// Package snapshot models the periodic status payload that Claude Hunter
// streams to its IDE clients over stdout.
package snapshot

import "time"

// Snapshot is one NDJSON record sent to the IDE client.
type Snapshot struct {
	Kind      string        `json:"type"`
	Timestamp time.Time     `json:"ts"`
	Model     string        `json:"model,omitempty"`
	Window    WindowSummary `json:"window5h"`
}

// WindowSummary carries the totals a client renders in the status bar.
type WindowSummary struct {
	InputTokens              int64            `json:"inputTokens"`
	OutputTokens             int64            `json:"outputTokens"`
	CacheCreateTokens        int64            `json:"cacheCreateTokens"`
	CacheReadTokens          int64            `json:"cacheReadTokens"`
	EffectiveTokens          int64            `json:"effectiveTokens"`
	CostUSD                  float64          `json:"costUSD"`
	BurnRatePerMinute        float64          `json:"burnRatePerMinute"`
	WindowStart              time.Time        `json:"windowStart"`
	WindowEnd                time.Time        `json:"windowEnd"`
	PercentOfCeilingEstimate float64          `json:"percentOfCeilingEstimate"`
	PerModel                 []ModelBreakdown `json:"perModel"`
}

// ModelBreakdown is one row of the tooltip's per-model attribution table.
type ModelBreakdown struct {
	Model             string  `json:"model"`
	InputTokens       int64   `json:"inputTokens"`
	OutputTokens      int64   `json:"outputTokens"`
	CacheCreateTokens int64   `json:"cacheCreateTokens"`
	CacheReadTokens   int64   `json:"cacheReadTokens"`
	EffectiveTokens   int64   `json:"effectiveTokens"`
	CostUSD           float64 `json:"costUSD"`
}
