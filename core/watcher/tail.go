// Package watcher observes Claude session files and emits new JSONL lines.
package watcher

import (
	"bufio"
	"errors"
	"io"
	"os"
)

// maxLineBytes bounds the size of a single JSONL record the tailer will
// accept. Claude Code session lines are typically well under this, but the
// cap protects the process from unbounded memory growth if a session file
// ever contains a pathological line without newline terminators.
const maxLineBytes = 4 * 1024 * 1024

// FileTailer reads new lines appended to a single file across successive
// calls to ReadNewLines. It buffers any trailing partial line so that
// callers only ever see complete newline-terminated records.
type FileTailer struct {
	filePath      string
	nextOffset    int64
	partialBuffer []byte
}

// NewFileTailer returns a tailer starting at offset zero.
func NewFileTailer(filePath string) *FileTailer {
	return &FileTailer{filePath: filePath}
}

// SeekToEnd advances the tailer past every byte currently in the file, so
// subsequent ReadNewLines calls only return content appended afterwards.
// Used at startup to avoid replaying months of historical session data.
func (t *FileTailer) SeekToEnd() error {
	fileInfo, err := os.Stat(t.filePath)
	if err != nil {
		return err
	}
	t.nextOffset = fileInfo.Size()
	t.partialBuffer = nil
	return nil
}

// ReadNewLines returns every complete line appended since the previous call.
// It streams the file line-by-line via bufio, so peak memory stays
// proportional to one record rather than the entire appended region.
func (t *FileTailer) ReadNewLines() ([][]byte, error) {
	fileHandle, err := os.Open(t.filePath)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	if _, err := fileHandle.Seek(t.nextOffset, io.SeekStart); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(fileHandle)
	var completeLines [][]byte
	var bytesConsumed int64
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			bytesConsumed += int64(len(line))
			terminated := line[len(line)-1] == '\n'
			if terminated {
				line = line[:len(line)-1]
			}
			if terminated {
				if merged, ok := t.mergePartial(line); ok {
					completeLines = append(completeLines, merged)
				}
			} else {
				if len(t.partialBuffer)+len(line) > maxLineBytes {
					return nil, errors.New("watcher: line exceeds max size")
				}
				t.partialBuffer = append(t.partialBuffer, line...)
			}
		}
		if readErr == nil {
			continue
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		return nil, readErr
	}
	t.nextOffset += bytesConsumed
	return completeLines, nil
}

// mergePartial combines any buffered trailing bytes from the previous read
// with the freshly-terminated line and returns the result. The bool is
// false when the resulting line is empty and should be skipped.
func (t *FileTailer) mergePartial(line []byte) ([]byte, bool) {
	if len(t.partialBuffer) == 0 {
		if len(line) == 0 {
			return nil, false
		}
		return line, true
	}
	merged := make([]byte, 0, len(t.partialBuffer)+len(line))
	merged = append(merged, t.partialBuffer...)
	merged = append(merged, line...)
	t.partialBuffer = nil
	if len(merged) == 0 {
		return nil, false
	}
	return merged, true
}
