package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProjectsObserverEmitsLineAppendedToNewSessionFile(t *testing.T) {
	rootDir := t.TempDir()

	observer, err := NewProjectsObserver(rootDir)
	if err != nil {
		t.Fatalf("new observer: %v", err)
	}

	observerContext, cancelObserver := context.WithCancel(context.Background())
	defer cancelObserver()
	go func() {
		if runErr := observer.Run(observerContext); runErr != nil && observerContext.Err() == nil {
			t.Logf("observer run: %v", runErr)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	projectDir := filepath.Join(rootDir, "some-project")
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	sessionPath := filepath.Join(projectDir, "session.jsonl")
	if err := os.WriteFile(sessionPath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	select {
	case emittedEvent := <-observer.Events():
		if string(emittedEvent.Line) != "hello" {
			t.Errorf("expected line 'hello', got %q", emittedEvent.Line)
		}
		if emittedEvent.SessionFile != sessionPath {
			t.Errorf("expected session file %q, got %q", sessionPath, emittedEvent.SessionFile)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for line event")
	}
}
