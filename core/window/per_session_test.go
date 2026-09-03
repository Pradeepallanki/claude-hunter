package window

import (
	"testing"
	"time"

	"github.com/reward360/claude-hunter/core/usage"
)

func TestPerSessionGroupsBySessionIDAndKeepsLatestContext(t *testing.T) {
	anchor := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	rollingWindow := NewRollingWindow(5 * time.Hour)

	rollingWindow.Add(usage.Record{
		SessionID:                "sess-A",
		Model:                    "claude-opus-4-7",
		Timestamp:                anchor,
		InputTokens:              100,
		OutputTokens:              50,
		CacheReadInputTokens:      200,
		CacheCreationInputTokens: 300,
	})
	rollingWindow.Add(usage.Record{
		SessionID:                "sess-A",
		Model:                    "claude-opus-4-7",
		Timestamp:                anchor.Add(1 * time.Minute),
		InputTokens:              9,
		OutputTokens:              1,
		CacheReadInputTokens:      50_000,
		CacheCreationInputTokens: 800_000,
	})
	rollingWindow.Add(usage.Record{
		SessionID:   "sess-B",
		Model:       "claude-haiku-4-5",
		Timestamp:   anchor.Add(30 * time.Second),
		InputTokens: 10,
		IsSidechain: true,
	})

	activities := rollingWindow.PerSession(anchor.Add(-1 * time.Minute))
	if len(activities) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(activities))
	}

	// Sorted by LastActiveAt desc: sess-A is fresher than sess-B.
	if activities[0].SessionID != "sess-A" {
		t.Errorf("expected sess-A first, got %q", activities[0].SessionID)
	}
	// Context tokens should reflect the LATEST record only, not the sum.
	if activities[0].ContextTokens != 9+50_000+800_000 {
		t.Errorf("contextTokens = %d, want %d", activities[0].ContextTokens, 9+50_000+800_000)
	}
	if activities[0].ContextWindowSize != 1_000_000 {
		t.Errorf("contextWindowSize = %d, want 1_000_000", activities[0].ContextWindowSize)
	}
	if activities[1].SessionID != "sess-B" {
		t.Errorf("expected sess-B second, got %q", activities[1].SessionID)
	}
	if activities[1].SidechainTurns != 1 {
		t.Errorf("sidechainTurns = %d, want 1", activities[1].SidechainTurns)
	}
	if activities[1].ContextWindowSize != 200_000 {
		t.Errorf("haiku contextWindowSize = %d, want 200_000", activities[1].ContextWindowSize)
	}
}

func TestPerSessionDropsSessionsOlderThanCutoff(t *testing.T) {
	anchor := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	rollingWindow := NewRollingWindow(5 * time.Hour)
	rollingWindow.Add(usage.Record{
		SessionID: "stale",
		Model:     "claude-opus-4-7",
		Timestamp: anchor.Add(-2 * time.Hour),
	})
	rollingWindow.Add(usage.Record{
		SessionID: "fresh",
		Model:     "claude-opus-4-7",
		Timestamp: anchor.Add(-10 * time.Second),
	})

	activities := rollingWindow.PerSession(anchor.Add(-1 * time.Minute))
	if len(activities) != 1 || activities[0].SessionID != "fresh" {
		t.Errorf("expected only fresh session, got %+v", activities)
	}
}

func TestContextWindowSizeForKnownFamilies(t *testing.T) {
	cases := map[string]int64{
		"claude-opus-4-7":          1_000_000,
		"claude-sonnet-4-6":        1_000_000,
		"claude-haiku-4-5-2025":    200_000,
		"unknown-model":             0,
	}
	for modelName, expected := range cases {
		if got := ContextWindowSizeFor(modelName); got != expected {
			t.Errorf("ContextWindowSizeFor(%q) = %d, want %d", modelName, got, expected)
		}
	}
}
