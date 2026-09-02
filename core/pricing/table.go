// Package pricing exposes per-model USD cost calculation for usage records.
// Rates are per one million tokens.
package pricing

import "strings"

// ModelRates carries the four token-pricing dimensions Anthropic bills on.
// All values are USD per one million tokens.
type ModelRates struct {
	InputPerMillion      float64
	OutputPerMillion     float64
	CacheWritePerMillion float64
	CacheReadPerMillion  float64
}

var modelRatesByPrefix = []struct {
	prefix string
	rates  ModelRates
}{
	{
		prefix: "claude-opus",
		rates: ModelRates{
			InputPerMillion:      15.00,
			OutputPerMillion:     75.00,
			CacheWritePerMillion: 18.75,
			CacheReadPerMillion:  1.50,
		},
	},
	{
		prefix: "claude-sonnet",
		rates: ModelRates{
			InputPerMillion:      3.00,
			OutputPerMillion:     15.00,
			CacheWritePerMillion: 3.75,
			CacheReadPerMillion:  0.30,
		},
	},
	{
		prefix: "claude-haiku",
		rates: ModelRates{
			InputPerMillion:      1.00,
			OutputPerMillion:     5.00,
			CacheWritePerMillion: 1.25,
			CacheReadPerMillion:  0.10,
		},
	},
}

// LookupModel returns the pricing entry that matches modelName by prefix,
// or nil when no known model family matches.
func LookupModel(modelName string) *ModelRates {
	for _, entry := range modelRatesByPrefix {
		if strings.HasPrefix(modelName, entry.prefix) {
			return &entry.rates
		}
	}
	return nil
}
