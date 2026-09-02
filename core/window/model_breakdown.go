package window

import (
	"sort"

	"github.com/reward360/claude-hunter/core/pricing"
	"github.com/reward360/claude-hunter/core/usage"
)

// ModelBreakdown is the per-model aggregation surfaced to IDE clients so the
// tooltip can attribute cost and tokens across mid-task model switches.
type ModelBreakdown struct {
	Model             string  `json:"model"`
	InputTokens       int64   `json:"inputTokens"`
	OutputTokens      int64   `json:"outputTokens"`
	CacheCreateTokens int64   `json:"cacheCreateTokens"`
	CacheReadTokens   int64   `json:"cacheReadTokens"`
	EffectiveTokens   int64   `json:"effectiveTokens"`
	CostUSD           float64 `json:"costUSD"`
}

// LatestModel returns the model of the retained record with the greatest
// timestamp, or an empty string when the window has no records. IDE clients
// display this as the "live" model.
func (w *RollingWindow) LatestModel() string {
	var latestRecord *usage.Record
	for index := range w.retainedRecords {
		candidate := &w.retainedRecords[index]
		if latestRecord == nil || candidate.Timestamp.After(latestRecord.Timestamp) {
			latestRecord = candidate
		}
	}
	if latestRecord == nil {
		return ""
	}
	return latestRecord.Model
}

// PerModel returns one aggregated entry per distinct model observed in the
// current window, sorted by descending cost so the biggest spender is first.
func (w *RollingWindow) PerModel() []ModelBreakdown {
	byModel := make(map[string]*ModelBreakdown)
	for _, record := range w.retainedRecords {
		entry, exists := byModel[record.Model]
		if !exists {
			entry = &ModelBreakdown{Model: record.Model}
			byModel[record.Model] = entry
		}
		entry.InputTokens += record.InputTokens
		entry.OutputTokens += record.OutputTokens
		entry.CacheCreateTokens += record.CacheCreationInputTokens
		entry.CacheReadTokens += record.CacheReadInputTokens
		entry.CostUSD += pricing.Calculate(record)
	}

	perModel := make([]ModelBreakdown, 0, len(byModel))
	for _, entry := range byModel {
		entry.EffectiveTokens = entry.InputTokens +
			entry.OutputTokens +
			entry.CacheCreateTokens +
			int64(float64(entry.CacheReadTokens)*CacheReadWeight)
		perModel = append(perModel, *entry)
	}
	sort.Slice(perModel, func(leftIndex, rightIndex int) bool {
		return perModel[leftIndex].CostUSD > perModel[rightIndex].CostUSD
	})
	return perModel
}
