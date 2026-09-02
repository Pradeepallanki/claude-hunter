package watcher

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func appendFile(t *testing.T, path, contents string) {
	t.Helper()
	fileHandle, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer fileHandle.Close()
	if _, err := fileHandle.WriteString(contents); err != nil {
		t.Fatalf("append %s: %v", path, err)
	}
}

func TestReadNewLinesReturnsAllCompleteLinesOnFirstCall(t *testing.T) {
	sessionPath := filepath.Join(t.TempDir(), "session.jsonl")
	writeFile(t, sessionPath, "one\ntwo\nthree\n")

	tailer := NewFileTailer(sessionPath)
	linesRead, err := tailer.ReadNewLines()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(linesRead); got != 3 {
		t.Fatalf("expected 3 lines, got %d (%q)", got, linesRead)
	}
	if string(linesRead[0]) != "one" || string(linesRead[2]) != "three" {
		t.Errorf("unexpected lines: %q", linesRead)
	}
}

func TestReadNewLinesReturnsEmptyWhenFileUnchanged(t *testing.T) {
	sessionPath := filepath.Join(t.TempDir(), "session.jsonl")
	writeFile(t, sessionPath, "a\nb\n")

	tailer := NewFileTailer(sessionPath)
	_, err := tailer.ReadNewLines()
	if err != nil {
		t.Fatalf("first read: %v", err)
	}

	linesRead, err := tailer.ReadNewLines()
	if err != nil {
		t.Fatalf("second read: %v", err)
	}
	if len(linesRead) != 0 {
		t.Errorf("expected 0 new lines, got %d (%q)", len(linesRead), linesRead)
	}
}

func TestReadNewLinesReturnsOnlyAppendedContent(t *testing.T) {
	sessionPath := filepath.Join(t.TempDir(), "session.jsonl")
	writeFile(t, sessionPath, "a\n")

	tailer := NewFileTailer(sessionPath)
	if _, err := tailer.ReadNewLines(); err != nil {
		t.Fatalf("first read: %v", err)
	}

	appendFile(t, sessionPath, "b\nc\n")

	linesRead, err := tailer.ReadNewLines()
	if err != nil {
		t.Fatalf("second read: %v", err)
	}
	if len(linesRead) != 2 || string(linesRead[0]) != "b" || string(linesRead[1]) != "c" {
		t.Errorf("expected [b c], got %q", linesRead)
	}
}

func TestReadNewLinesBuffersPartialTrailingLine(t *testing.T) {
	sessionPath := filepath.Join(t.TempDir(), "session.jsonl")
	writeFile(t, sessionPath, "complete\npart")

	tailer := NewFileTailer(sessionPath)
	firstBatch, err := tailer.ReadNewLines()
	if err != nil {
		t.Fatalf("first read: %v", err)
	}
	if len(firstBatch) != 1 || string(firstBatch[0]) != "complete" {
		t.Fatalf("expected only [complete] on first read, got %q", firstBatch)
	}

	appendFile(t, sessionPath, "ial\n")

	secondBatch, err := tailer.ReadNewLines()
	if err != nil {
		t.Fatalf("second read: %v", err)
	}
	if len(secondBatch) != 1 || string(secondBatch[0]) != "partial" {
		t.Errorf("expected [partial] after completion, got %q", secondBatch)
	}
}

func TestReadNewLinesReturnsErrorForMissingFile(t *testing.T) {
	tailer := NewFileTailer(filepath.Join(t.TempDir(), "does-not-exist.jsonl"))
	if _, err := tailer.ReadNewLines(); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
