// Package watcher observes Claude session files and emits new JSONL lines.
package watcher

import (
	"bytes"
	"io"
	"os"
)

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

// ReadNewLines returns every complete line appended since the previous call.
func (t *FileTailer) ReadNewLines() ([][]byte, error) {
	fileHandle, err := os.Open(t.filePath)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	if _, err := fileHandle.Seek(t.nextOffset, io.SeekStart); err != nil {
		return nil, err
	}

	appendedBytes, err := io.ReadAll(fileHandle)
	if err != nil {
		return nil, err
	}
	t.nextOffset += int64(len(appendedBytes))

	combined := append(t.partialBuffer, appendedBytes...)
	splitLines := bytes.Split(combined, []byte("\n"))
	trailingPartial := splitLines[len(splitLines)-1]
	completeLines := splitLines[:len(splitLines)-1]

	t.partialBuffer = append(t.partialBuffer[:0], trailingPartial...)

	copiedLines := make([][]byte, 0, len(completeLines))
	for _, line := range completeLines {
		if len(line) == 0 {
			continue
		}
		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)
		copiedLines = append(copiedLines, lineCopy)
	}
	return copiedLines, nil
}
