// Package window aggregates usage records over a rolling time window so
// callers can inspect current-window token totals, cost, and burn rate.
package window

import (
	"time"

	"github.com/reward360/claude-hunter/core/pricing"
	"github.com/reward360/claude-hunter/core/usage"
)

// CacheReadWeight is the fraction of a cached-read token that counts toward
// Anthropic's effective-token rate limit.
const CacheReadWeight = 0.1

// Totals is the aggregated view a RollingWindow reports for its retained
// records.
type Totals struct {
	InputTokens       int64
	OutputTokens      int64
	CacheCreateTokens int64
	CacheReadTokens   int64
	EffectiveTokens   int64
	CostUSD           float64
}

// RollingWindow keeps usage records for a configurable duration and exposes
// aggregated views. It is not safe for concurrent use; callers must
// synchronise access externally.
type RollingWindow struct {
	windowDuration  time.Duration
	retainedRecords []usage.Record
}

// NewRollingWindow returns a window that retains records for windowDuration.
func NewRollingWindow(windowDuration time.Duration) *RollingWindow {
	return &RollingWindow{windowDuration: windowDuration}
}

// Add appends a record. Callers are expected to call PruneBefore periodically
// to bound memory.
func (w *RollingWindow) Add(record usage.Record) {
	w.retainedRecords = append(w.retainedRecords, record)
}

// PruneBefore drops records whose timestamp is strictly before cutoff.
// Records may arrive out of order, so the whole slice is scanned rather
// than assuming a sorted prefix.
func (w *RollingWindow) PruneBefore(cutoff time.Time) {
	kept := w.retainedRecords[:0]
	for _, record := range w.retainedRecords {
		if !record.Timestamp.Before(cutoff) {
			kept = append(kept, record)
		}
	}
	w.retainedRecords = kept
}

// Totals sums every retained record.
func (w *RollingWindow) Totals() Totals {
	var totals Totals
	for _, record := range w.retainedRecords {
		totals.InputTokens += record.InputTokens
		totals.OutputTokens += record.OutputTokens
		totals.CacheCreateTokens += record.CacheCreationInputTokens
		totals.CacheReadTokens += record.CacheReadInputTokens
		totals.CostUSD += pricing.Calculate(record)
	}
	totals.EffectiveTokens = totals.InputTokens +
		totals.OutputTokens +
		totals.CacheCreateTokens +
		int64(float64(totals.CacheReadTokens)*CacheReadWeight)
	return totals
}

// BurnRatePerMinute returns effective tokens per minute observed over the
// sampleDuration ending at now.
func (w *RollingWindow) BurnRatePerMinute(now time.Time, sampleDuration time.Duration) float64 {
	sampleStart := now.Add(-sampleDuration)
	var effectiveInSample int64
	for _, record := range w.retainedRecords {
		if record.Timestamp.Before(sampleStart) {
			continue
		}
		effectiveInSample += record.InputTokens +
			record.OutputTokens +
			record.CacheCreationInputTokens +
			int64(float64(record.CacheReadInputTokens)*CacheReadWeight)
	}
	minutesInSample := sampleDuration.Minutes()
	if minutesInSample == 0 {
		return 0
	}
	return float64(effectiveInSample) / minutesInSample
}
