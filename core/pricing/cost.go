package pricing

import "github.com/reward360/claude-hunter/core/usage"

const tokensPerMillion = 1_000_000.0

// Calculate returns the USD cost for the tokens recorded in a single usage
// record. Returns zero when the model has no known pricing entry.
func Calculate(record usage.Record) float64 {
	rates := LookupModel(record.Model)
	if rates == nil {
		return 0
	}
	inputCost := float64(record.InputTokens) / tokensPerMillion * rates.InputPerMillion
	outputCost := float64(record.OutputTokens) / tokensPerMillion * rates.OutputPerMillion
	cacheWriteCost := float64(record.CacheCreationInputTokens) / tokensPerMillion * rates.CacheWritePerMillion
	cacheReadCost := float64(record.CacheReadInputTokens) / tokensPerMillion * rates.CacheReadPerMillion
	return inputCost + outputCost + cacheWriteCost + cacheReadCost
}
