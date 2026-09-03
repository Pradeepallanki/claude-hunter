package window

import (
	"sort"
	"strings"
	"time"

	"github.com/pradeep/claude-hunter/core/usage"
)

// SessionActivity is the per-session view surfaced to IDE clients so the
// details panel can list every Claude Code session currently burning tokens.
type SessionActivity struct {
	SessionID         string    `json:"sessionId"`
	Model             string    `json:"model"`
	ContextTokens     int64     `json:"contextTokens"`
	ContextWindowSize int64     `json:"contextWindowSize"`
	TotalTokens       int64     `json:"totalTokens"`
	SidechainTurns    int64     `json:"sidechainTurns"`
	LastActiveAt      time.Time `json:"lastActiveAt"`
}

// PerSession returns one entry per distinct sessionId observed in the window
// whose most recent record is not older than recencyCutoff. Sorted by
// descending LastActiveAt so the freshest session is first.
func (w *RollingWindow) PerSession(recencyCutoff time.Time) []SessionActivity {
	type accumulator struct {
		latestRecord   *usage.Record
		totalEffective int64
		sidechainTurns int64
	}

	bySession := make(map[string]*accumulator)
	for index := range w.retainedRecords {
		record := &w.retainedRecords[index]
		if record.SessionID == "" {
			continue
		}
		entry, exists := bySession[record.SessionID]
		if !exists {
			entry = &accumulator{}
			bySession[record.SessionID] = entry
		}
		entry.totalEffective += record.InputTokens +
			record.OutputTokens +
			record.CacheCreationInputTokens +
			int64(float64(record.CacheReadInputTokens)*CacheReadWeight)
		if record.IsSidechain {
			entry.sidechainTurns++
		}
		if entry.latestRecord == nil || record.Timestamp.After(entry.latestRecord.Timestamp) {
			entry.latestRecord = record
		}
	}

	activities := make([]SessionActivity, 0, len(bySession))
	for sessionID, entry := range bySession {
		latest := entry.latestRecord
		if latest.Timestamp.Before(recencyCutoff) {
			continue
		}
		activities = append(activities, SessionActivity{
			SessionID: sessionID,
			Model:     latest.Model,
			ContextTokens: latest.InputTokens +
				latest.CacheCreationInputTokens +
				latest.CacheReadInputTokens,
			ContextWindowSize: ContextWindowSizeFor(latest.Model),
			TotalTokens:       entry.totalEffective,
			SidechainTurns:    entry.sidechainTurns,
			LastActiveAt:      latest.Timestamp,
		})
	}

	sort.Slice(activities, func(leftIndex, rightIndex int) bool {
		return activities[leftIndex].LastActiveAt.After(activities[rightIndex].LastActiveAt)
	})
	return activities
}

// ContextWindowSizeFor returns the approximate context window (in tokens) for
// the given model name. Zero means unknown.
func ContextWindowSizeFor(modelName string) int64 {
	lowered := strings.ToLower(modelName)
	switch {
	case strings.Contains(lowered, "opus"):
		return 1_000_000
	case strings.Contains(lowered, "sonnet"):
		return 1_000_000
	case strings.Contains(lowered, "haiku"):
		return 200_000
	default:
		return 0
	}
}
