package window

import (
	"math"
	"testing"
	"time"

	"github.com/pradeep/claude-hunter/core/usage"
)

func TestLatestModelReturnsModelOfNewestTimestamp(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	baseTime := time.Now()

	rollingWindow.Add(usage.Record{
		Model:     "claude-haiku-4-5",
		Timestamp: baseTime.Add(-2 * time.Hour),
	})
	rollingWindow.Add(usage.Record{
		Model:     "claude-opus-4-7",
		Timestamp: baseTime.Add(-1 * time.Minute),
	})
	rollingWindow.Add(usage.Record{
		Model:     "claude-haiku-4-5",
		Timestamp: baseTime.Add(-30 * time.Minute),
	})

	if got := rollingWindow.LatestModel(); got != "claude-opus-4-7" {
		t.Errorf("latest model: got %q, want claude-opus-4-7", got)
	}
}

func TestLatestModelReturnsEmptyForEmptyWindow(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	if got := rollingWindow.LatestModel(); got != "" {
		t.Errorf("latest model on empty window: got %q, want empty", got)
	}
}

func TestPerModelAggregatesTokensAndCostByModel(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	baseTime := time.Now()

	rollingWindow.Add(usage.Record{
		Model:                    "claude-opus-4-7",
		Timestamp:                baseTime.Add(-10 * time.Minute),
		InputTokens:              1_000_000,
		OutputTokens:             500_000,
		CacheCreationInputTokens: 0,
		CacheReadInputTokens:     0,
	})
	rollingWindow.Add(usage.Record{
		Model:                    "claude-opus-4-7",
		Timestamp:                baseTime.Add(-5 * time.Minute),
		InputTokens:              1_000_000,
		OutputTokens:             500_000,
		CacheCreationInputTokens: 0,
		CacheReadInputTokens:     0,
	})
	rollingWindow.Add(usage.Record{
		Model:                    "claude-haiku-4-5",
		Timestamp:                baseTime.Add(-1 * time.Minute),
		InputTokens:              1_000_000,
		OutputTokens:             1_000_000,
		CacheCreationInputTokens: 0,
		CacheReadInputTokens:     0,
	})

	perModel := rollingWindow.PerModel()

	if len(perModel) != 2 {
		t.Fatalf("expected 2 model entries, got %d: %+v", len(perModel), perModel)
	}
	if perModel[0].Model != "claude-opus-4-7" {
		t.Errorf("expected opus first (highest cost), got %q", perModel[0].Model)
	}

	opusEntry := findEntry(perModel, "claude-opus-4-7")
	if opusEntry == nil {
		t.Fatal("no opus entry")
	}
	if opusEntry.InputTokens != 2_000_000 || opusEntry.OutputTokens != 1_000_000 {
		t.Errorf("opus tokens: got in=%d out=%d, want in=2M out=1M",
			opusEntry.InputTokens, opusEntry.OutputTokens)
	}
	expectedOpusCost := 2*15.0 + 1*75.0
	if math.Abs(opusEntry.CostUSD-expectedOpusCost) > 0.001 {
		t.Errorf("opus cost: got %.4f, want %.4f", opusEntry.CostUSD, expectedOpusCost)
	}

	haikuEntry := findEntry(perModel, "claude-haiku-4-5")
	if haikuEntry == nil {
		t.Fatal("no haiku entry")
	}
	expectedHaikuEffective := int64(1_000_000 + 1_000_000)
	if haikuEntry.EffectiveTokens != expectedHaikuEffective {
		t.Errorf("haiku effective: got %d, want %d", haikuEntry.EffectiveTokens, expectedHaikuEffective)
	}
}

func TestPerModelReturnsEmptyWhenNoRecords(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	if got := rollingWindow.PerModel(); len(got) != 0 {
		t.Errorf("empty window: got %d entries, want 0", len(got))
	}
}

func findEntry(entries []ModelBreakdown, model string) *ModelBreakdown {
	for i := range entries {
		if entries[i].Model == model {
			return &entries[i]
		}
	}
	return nil
}
