package watcher

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// LineEvent is a single newline-terminated record read from a session file.
type LineEvent struct {
	SessionFile string
	Line        []byte
}

// ProjectsObserver watches a Claude projects root and streams every JSONL
// line appended to any session file below it.
type ProjectsObserver struct {
	rootDir       string
	fsWatcher     *fsnotify.Watcher
	tailersByPath map[string]*FileTailer
	tailerMutex   sync.Mutex
	events        chan LineEvent
	errors        chan error
}

// NewProjectsObserver constructs an observer over rootDir. Callers own the
// lifecycle via Run.
func NewProjectsObserver(rootDir string) (*ProjectsObserver, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &ProjectsObserver{
		rootDir:       rootDir,
		fsWatcher:     fsWatcher,
		tailersByPath: make(map[string]*FileTailer),
		events:        make(chan LineEvent, 256),
		errors:        make(chan error, 8),
	}, nil
}

// Events returns the channel of newly-observed lines.
func (o *ProjectsObserver) Events() <-chan LineEvent {
	return o.events
}

// Errors returns non-fatal errors encountered while tailing.
func (o *ProjectsObserver) Errors() <-chan error {
	return o.errors
}

// Run blocks until ctx is cancelled, watching rootDir and every discovered
// subdirectory, seeding tailers for pre-existing session files, and emitting
// new lines as files change.
func (o *ProjectsObserver) Run(ctx context.Context) error {
	defer o.fsWatcher.Close()

	if err := o.seedExistingTree(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case fsEvent := <-o.fsWatcher.Events:
			o.handleFsEvent(fsEvent)
		case fsErr := <-o.fsWatcher.Errors:
			o.reportError(fsErr)
		}
	}
}

func (o *ProjectsObserver) seedExistingTree() error {
	if err := os.MkdirAll(o.rootDir, 0o755); err != nil {
		return err
	}
	if err := o.fsWatcher.Add(o.rootDir); err != nil {
		return err
	}
	return filepath.WalkDir(o.rootDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() {
			if path == o.rootDir {
				return nil
			}
			_ = o.fsWatcher.Add(path)
			return nil
		}
		if isSessionFile(path) {
			o.registerTailer(path)
			o.drainTailer(path)
		}
		return nil
	})
}

func (o *ProjectsObserver) handleFsEvent(fsEvent fsnotify.Event) {
	if fsEvent.Op&fsnotify.Create != 0 {
		fileInfo, statErr := os.Stat(fsEvent.Name)
		if statErr == nil && fileInfo.IsDir() {
			_ = o.fsWatcher.Add(fsEvent.Name)
			return
		}
		if isSessionFile(fsEvent.Name) {
			o.registerTailer(fsEvent.Name)
		}
	}
	if fsEvent.Op&(fsnotify.Write|fsnotify.Create) != 0 && isSessionFile(fsEvent.Name) {
		o.drainTailer(fsEvent.Name)
	}
}

func (o *ProjectsObserver) registerTailer(sessionPath string) {
	o.tailerMutex.Lock()
	defer o.tailerMutex.Unlock()
	if _, exists := o.tailersByPath[sessionPath]; !exists {
		o.tailersByPath[sessionPath] = NewFileTailer(sessionPath)
	}
}

func (o *ProjectsObserver) drainTailer(sessionPath string) {
	o.tailerMutex.Lock()
	tailer, exists := o.tailersByPath[sessionPath]
	o.tailerMutex.Unlock()
	if !exists {
		return
	}

	newLines, err := tailer.ReadNewLines()
	if err != nil {
		o.reportError(err)
		return
	}
	for _, line := range newLines {
		o.events <- LineEvent{SessionFile: sessionPath, Line: line}
	}
}

func (o *ProjectsObserver) reportError(err error) {
	select {
	case o.errors <- err:
	default:
	}
}

func isSessionFile(path string) bool {
	return strings.HasSuffix(path, ".jsonl")
}
