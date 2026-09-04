package window

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/pradeep/claude-hunter/core/pricing"
)

// projectRootCache memoises the walk-up-for-.git lookup so it runs once per
// unique cwd rather than every 250ms emission tick.
var projectRootCache sync.Map

// resolveProjectRoot returns the enclosing git-repo root for cwd, or cwd
// itself when no .git directory is found on the way up to the filesystem
// root. Two different subdirectories of the same repo therefore collapse
// onto the same project name.
func resolveProjectRoot(cwd string) string {
	if cwd == "" {
		return ""
	}
	if cached, ok := projectRootCache.Load(cwd); ok {
		return cached.(string)
	}
	current := cwd
	for {
		if fileInfo, err := os.Stat(filepath.Join(current, ".git")); err == nil && (fileInfo.IsDir() || fileInfo.Mode().IsRegular()) {
			projectRootCache.Store(cwd, current)
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			projectRootCache.Store(cwd, cwd)
			return cwd
		}
		current = parent
	}
}

// ProjectActivity is the per-project rollup surfaced to IDE clients.
// Projects are grouped by the basename of the working directory recorded on
// each assistant turn (`cwd` in the JSONL), so two clones of the same repo in
// different paths still share one row.
type ProjectActivity struct {
	Project      string    `json:"project"`
	CWD          string    `json:"cwd"`
	Sessions     int       `json:"sessions"`
	TotalTokens  int64     `json:"totalTokens"`
	CostUSD      float64   `json:"costUSD"`
	LastActiveAt time.Time `json:"lastActiveAt"`
}

// PerProject returns one entry per distinct project observed in the window.
// Sorted by descending CostUSD so the priciest project is first.
func (w *RollingWindow) PerProject() []ProjectActivity {
	type accumulator struct {
		project      string
		cwd          string
		sessionIDs   map[string]struct{}
		totalTokens  int64
		costUSD      float64
		lastActiveAt time.Time
	}

	byProject := make(map[string]*accumulator)
	for index := range w.retainedRecords {
		record := &w.retainedRecords[index]
		if record.CWD == "" {
			continue
		}
		projectRoot := resolveProjectRoot(record.CWD)
		projectName := filepath.Base(projectRoot)
		entry, exists := byProject[projectRoot]
		if !exists {
			entry = &accumulator{
				project:    projectName,
				cwd:        projectRoot,
				sessionIDs: make(map[string]struct{}),
			}
			byProject[projectRoot] = entry
		}
		entry.totalTokens += record.InputTokens +
			record.OutputTokens +
			record.CacheCreationInputTokens +
			int64(float64(record.CacheReadInputTokens)*CacheReadWeight)
		entry.costUSD += pricing.Calculate(*record)
		if record.SessionID != "" {
			entry.sessionIDs[record.SessionID] = struct{}{}
		}
		if record.Timestamp.After(entry.lastActiveAt) {
			entry.lastActiveAt = record.Timestamp
		}
	}

	activities := make([]ProjectActivity, 0, len(byProject))
	for _, entry := range byProject {
		activities = append(activities, ProjectActivity{
			Project:      entry.project,
			CWD:          entry.cwd,
			Sessions:     len(entry.sessionIDs),
			TotalTokens:  entry.totalTokens,
			CostUSD:      entry.costUSD,
			LastActiveAt: entry.lastActiveAt,
		})
	}
	sort.Slice(activities, func(leftIndex, rightIndex int) bool {
		return activities[leftIndex].CostUSD > activities[rightIndex].CostUSD
	})
	return activities
}
