package pricing

import (
	"math"
	"testing"

	"github.com/pradeep/claude-hunter/core/usage"
)

func TestCalculateAppliesOpusRates(t *testing.T) {
	record := usage.Record{
		Model:                    "claude-opus-4-7",
		InputTokens:              1_000_000,
		OutputTokens:             1_000_000,
		CacheCreationInputTokens: 1_000_000,
		CacheReadInputTokens:     1_000_000,
	}

	calculatedCostUSD := Calculate(record)

	expectedCostUSD := 15.0 + 75.0 + 18.75 + 1.50
	if math.Abs(calculatedCostUSD-expectedCostUSD) > 0.001 {
		t.Errorf("opus cost: got %.4f, want %.4f", calculatedCostUSD, expectedCostUSD)
	}
}

func TestCalculateAppliesSonnetRates(t *testing.T) {
	record := usage.Record{
		Model:        "claude-sonnet-4-6",
		InputTokens:  1_000_000,
		OutputTokens: 1_000_000,
	}

	calculatedCostUSD := Calculate(record)

	expectedCostUSD := 3.0 + 15.0
	if math.Abs(calculatedCostUSD-expectedCostUSD) > 0.001 {
		t.Errorf("sonnet cost: got %.4f, want %.4f", calculatedCostUSD, expectedCostUSD)
	}
}

func TestCalculateReturnsZeroForUnknownModel(t *testing.T) {
	record := usage.Record{
		Model:        "some-future-model",
		InputTokens:  1_000_000,
		OutputTokens: 1_000_000,
	}

	if calculated := Calculate(record); calculated != 0 {
		t.Errorf("unknown model: expected 0 cost, got %.4f", calculated)
	}
}

func TestCalculateHandlesFractionalTokens(t *testing.T) {
	record := usage.Record{
		Model:        "claude-haiku-4-5-20251001",
		InputTokens:  500_000,
		OutputTokens: 100_000,
	}

	calculatedCostUSD := Calculate(record)

	expectedCostUSD := 0.5*1.0 + 0.1*5.0
	if math.Abs(calculatedCostUSD-expectedCostUSD) > 0.001 {
		t.Errorf("haiku cost: got %.4f, want %.4f", calculatedCostUSD, expectedCostUSD)
	}
}
