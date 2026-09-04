package window

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pradeep/claude-hunter/core/usage"
)

func TestPerProjectRollsUpSubdirectoriesToTheGitRoot(t *testing.T) {
	// Create a temp repo with two subdirectories. Records with cwds in
	// different subdirs must merge under a single project keyed to the
	// repo root's basename.
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	frontendDir := filepath.Join(repoRoot, "frontend")
	backendDir := filepath.Join(repoRoot, "backend")
	for _, dir := range []string{frontendDir, backendDir} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	anchor := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	rollingWindow := NewRollingWindow(5 * time.Hour)
	rollingWindow.Add(usage.Record{
		SessionID:    "sess-frontend",
		Model:        "claude-opus-4-7",
		Timestamp:    anchor,
		CWD:          frontendDir,
		InputTokens:  100,
		OutputTokens: 50,
	})
	rollingWindow.Add(usage.Record{
		SessionID:    "sess-backend",
		Model:        "claude-opus-4-7",
		Timestamp:    anchor.Add(1 * time.Minute),
		CWD:          backendDir,
		InputTokens:  10,
		OutputTokens: 5,
	})

	projects := rollingWindow.PerProject()
	if len(projects) != 1 {
		t.Fatalf("expected 1 project (git root rollup), got %d: %+v", len(projects), projects)
	}
	rolled := projects[0]
	if rolled.Project != filepath.Base(repoRoot) {
		t.Errorf("project name: got %q, want %q", rolled.Project, filepath.Base(repoRoot))
	}
	if rolled.Sessions != 2 {
		t.Errorf("sessions: got %d, want 2", rolled.Sessions)
	}
	if rolled.TotalTokens != 100+50+10+5 {
		t.Errorf("totalTokens: got %d, want 165", rolled.TotalTokens)
	}
}

func TestPerProjectFallsBackToCWDWhenNoGitRoot(t *testing.T) {
	nonRepoDir := t.TempDir()
	rollingWindow := NewRollingWindow(5 * time.Hour)
	rollingWindow.Add(usage.Record{
		SessionID:    "s1",
		Model:        "claude-opus-4-7",
		Timestamp:    time.Now(),
		CWD:          nonRepoDir,
		InputTokens:  1,
		OutputTokens: 1,
	})

	projects := rollingWindow.PerProject()
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Project != filepath.Base(nonRepoDir) {
		t.Errorf("expected basename of cwd (%q), got %q", filepath.Base(nonRepoDir), projects[0].Project)
	}
}

func TestPerProjectSortedByCostDescending(t *testing.T) {
	cheapDir := t.TempDir()
	priceyDir := t.TempDir()
	anchor := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	rollingWindow := NewRollingWindow(5 * time.Hour)

	rollingWindow.Add(usage.Record{
		SessionID:    "s1",
		Model:        "claude-haiku-4-5",
		Timestamp:    anchor,
		CWD:          cheapDir,
		InputTokens:  1_000_000,
		OutputTokens: 0,
	})
	rollingWindow.Add(usage.Record{
		SessionID:    "s2",
		Model:        "claude-opus-4-7",
		Timestamp:    anchor,
		CWD:          priceyDir,
		InputTokens:  1_000_000,
		OutputTokens: 0,
	})

	projects := rollingWindow.PerProject()
	if projects[0].Project != filepath.Base(priceyDir) {
		t.Errorf("expected pricey first, got %q (%+v)", projects[0].Project, projects)
	}
}

func TestPerProjectSkipsRecordsWithoutCWD(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	rollingWindow.Add(usage.Record{
		SessionID: "s1",
		Model:     "claude-opus-4-7",
		Timestamp: time.Now(),
	})
	if projects := rollingWindow.PerProject(); len(projects) != 0 {
		t.Errorf("expected 0 projects for cwd-less records, got %+v", projects)
	}
}
