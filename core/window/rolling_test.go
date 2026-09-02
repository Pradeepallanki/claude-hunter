package window

import (
	"math"
	"testing"
	"time"

	"github.com/reward360/claude-hunter/core/usage"
)

func recordAt(offset time.Duration, base time.Time) usage.Record {
	return usage.Record{
		SessionID:                "session-1",
		Model:                    "claude-opus-4-7",
		Timestamp:                base.Add(offset),
		InputTokens:              1_000,
		OutputTokens:             500,
		CacheCreationInputTokens: 4_000,
		CacheReadInputTokens:     10_000,
	}
}

func TestTotalsSumsRetainedRecords(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	baseTime := time.Now()

	rollingWindow.Add(recordAt(-2*time.Hour, baseTime))
	rollingWindow.Add(recordAt(-1*time.Hour, baseTime))

	totals := rollingWindow.Totals()

	if totals.InputTokens != 2_000 {
		t.Errorf("input tokens: got %d, want 2000", totals.InputTokens)
	}
	if totals.OutputTokens != 1_000 {
		t.Errorf("output tokens: got %d, want 1000", totals.OutputTokens)
	}
	if totals.CacheCreateTokens != 8_000 {
		t.Errorf("cache create tokens: got %d, want 8000", totals.CacheCreateTokens)
	}
	if totals.CacheReadTokens != 20_000 {
		t.Errorf("cache read tokens: got %d, want 20000", totals.CacheReadTokens)
	}
}

func TestTotalsAppliesCacheReadWeightForEffectiveTokens(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	baseTime := time.Now()

	rollingWindow.Add(usage.Record{
		Model:                    "claude-opus-4-7",
		Timestamp:                baseTime,
		InputTokens:              1_000,
		OutputTokens:             500,
		CacheCreationInputTokens: 4_000,
		CacheReadInputTokens:     10_000,
	})

	totals := rollingWindow.Totals()

	expectedEffective := int64(1_000 + 500 + 4_000 + int64(10_000*0.1))
	if totals.EffectiveTokens != expectedEffective {
		t.Errorf("effective tokens: got %d, want %d", totals.EffectiveTokens, expectedEffective)
	}
}

func TestPruneBeforeRemovesRecordsOlderThanCutoff(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	baseTime := time.Now()

	rollingWindow.Add(recordAt(-6*time.Hour, baseTime))
	rollingWindow.Add(recordAt(-4*time.Hour, baseTime))
	rollingWindow.Add(recordAt(-1*time.Hour, baseTime))

	rollingWindow.PruneBefore(baseTime.Add(-5 * time.Hour))

	totals := rollingWindow.Totals()
	if totals.InputTokens != 2_000 {
		t.Errorf("expected 2 records retained (2000 input), got %d", totals.InputTokens)
	}
}

func TestTotalsIncludesCostFromPricingTable(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	rollingWindow.Add(usage.Record{
		Model:        "claude-sonnet-4-6",
		Timestamp:    time.Now(),
		InputTokens:  1_000_000,
		OutputTokens: 1_000_000,
	})

	totals := rollingWindow.Totals()

	expectedCostUSD := 3.0 + 15.0
	if math.Abs(totals.CostUSD-expectedCostUSD) > 0.001 {
		t.Errorf("cost: got %.4f, want %.4f", totals.CostUSD, expectedCostUSD)
	}
}

func TestBurnRatePerMinuteMeasuresRecentEffectiveTokens(t *testing.T) {
	rollingWindow := NewRollingWindow(5 * time.Hour)
	baseTime := time.Now()

	rollingWindow.Add(usage.Record{
		Model:        "claude-opus-4-7",
		Timestamp:    baseTime.Add(-30 * time.Minute),
		InputTokens:  100_000,
		OutputTokens: 100_000,
	})
	rollingWindow.Add(usage.Record{
		Model:        "claude-opus-4-7",
		Timestamp:    baseTime.Add(-2 * time.Minute),
		InputTokens:  5_000,
		OutputTokens: 5_000,
	})

	tokensPerMinute := rollingWindow.BurnRatePerMinute(baseTime, 10*time.Minute)

	expectedTokensPerMinute := float64(10_000) / 10.0
	if math.Abs(tokensPerMinute-expectedTokensPerMinute) > 0.001 {
		t.Errorf("burn rate: got %.4f, want %.4f", tokensPerMinute, expectedTokensPerMinute)
	}
}
