package snapshot

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSnapshotMarshalsExpectedFields(t *testing.T) {
	fixedTime := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	snapshot := Snapshot{
		Kind:      "snapshot",
		Timestamp: fixedTime,
		Model:     "claude-opus-4-7",
		Window: WindowSummary{
			InputTokens:        1_000,
			OutputTokens:       500,
			CacheCreateTokens:  4_000,
			CacheReadTokens:    10_000,
			EffectiveTokens:    5_500,
			CostUSD:            0.123456,
			BurnRatePerMinute:  42.5,
			WindowStart:        fixedTime.Add(-5 * time.Hour),
			WindowEnd:          fixedTime,
			PercentOfCeilingEstimate: 12.3,
		},
	}

	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decodedInto map[string]any
	if err := json.Unmarshal(encoded, &decodedInto); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decodedInto["type"] != "snapshot" {
		t.Errorf("expected top-level type=snapshot, got %v", decodedInto["type"])
	}
	if decodedInto["model"] != "claude-opus-4-7" {
		t.Errorf("expected model, got %v", decodedInto["model"])
	}
	window, isMap := decodedInto["window5h"].(map[string]any)
	if !isMap {
		t.Fatalf("expected window5h object, got %T", decodedInto["window5h"])
	}
	if window["effectiveTokens"].(float64) != 5_500 {
		t.Errorf("effectiveTokens mismatch: %v", window["effectiveTokens"])
	}
	if window["burnRatePerMinute"].(float64) != 42.5 {
		t.Errorf("burn rate mismatch: %v", window["burnRatePerMinute"])
	}
}
