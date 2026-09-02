package usage

import (
	"testing"
	"time"
)

func TestParseLineExtractsUsageFromAssistantRecord(t *testing.T) {
	rawLine := []byte(`{"parentUuid":"p-1","isSidechain":false,"message":{"model":"claude-opus-4-7","id":"msg_1","type":"message","role":"assistant","content":[],"usage":{"input_tokens":6,"cache_creation_input_tokens":24794,"cache_read_input_tokens":100,"output_tokens":135}},"type":"assistant","sessionId":"session-abc","timestamp":"2026-08-20T04:38:24.503Z"}`)

	parsedRecord, err := ParseLine(rawLine)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedRecord == nil {
		t.Fatal("expected non-nil record for assistant line with usage")
	}

	expectedTimestamp := time.Date(2026, 8, 20, 4, 38, 24, 503_000_000, time.UTC)
	if !parsedRecord.Timestamp.Equal(expectedTimestamp) {
		t.Errorf("timestamp: got %v, want %v", parsedRecord.Timestamp, expectedTimestamp)
	}
	if parsedRecord.SessionID != "session-abc" {
		t.Errorf("sessionID: got %q, want %q", parsedRecord.SessionID, "session-abc")
	}
	if parsedRecord.Model != "claude-opus-4-7" {
		t.Errorf("model: got %q, want %q", parsedRecord.Model, "claude-opus-4-7")
	}
	if parsedRecord.InputTokens != 6 {
		t.Errorf("input tokens: got %d, want 6", parsedRecord.InputTokens)
	}
	if parsedRecord.OutputTokens != 135 {
		t.Errorf("output tokens: got %d, want 135", parsedRecord.OutputTokens)
	}
	if parsedRecord.CacheCreationInputTokens != 24794 {
		t.Errorf("cache creation tokens: got %d, want 24794", parsedRecord.CacheCreationInputTokens)
	}
	if parsedRecord.CacheReadInputTokens != 100 {
		t.Errorf("cache read tokens: got %d, want 100", parsedRecord.CacheReadInputTokens)
	}
}

func TestParseLineReturnsNilForNonAssistantRecord(t *testing.T) {
	rawLine := []byte(`{"type":"permission-mode","permissionMode":"default","sessionId":"session-abc"}`)

	parsedRecord, err := ParseLine(rawLine)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedRecord != nil {
		t.Errorf("expected nil for non-assistant record, got %+v", parsedRecord)
	}
}

func TestParseLineReturnsNilForAssistantRecordWithoutUsage(t *testing.T) {
	rawLine := []byte(`{"type":"assistant","sessionId":"s","timestamp":"2026-08-20T04:38:24.503Z","message":{"model":"claude-opus-4-7","content":[]}}`)

	parsedRecord, err := ParseLine(rawLine)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedRecord != nil {
		t.Errorf("expected nil for assistant record without usage, got %+v", parsedRecord)
	}
}

func TestParseLineReturnsErrorForMalformedJSON(t *testing.T) {
	rawLine := []byte(`{"type":"assistant",`)

	parsedRecord, err := ParseLine(rawLine)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if parsedRecord != nil {
		t.Errorf("expected nil record on parse error, got %+v", parsedRecord)
	}
}

func TestParseLineHandlesMissingCacheFieldsAsZero(t *testing.T) {
	rawLine := []byte(`{"type":"assistant","sessionId":"s","timestamp":"2026-08-20T04:38:24.503Z","message":{"model":"claude-sonnet-4-6","usage":{"input_tokens":10,"output_tokens":20}}}`)

	parsedRecord, err := ParseLine(rawLine)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedRecord == nil {
		t.Fatal("expected non-nil record")
	}
	if parsedRecord.CacheCreationInputTokens != 0 || parsedRecord.CacheReadInputTokens != 0 {
		t.Errorf("expected zero cache tokens when fields missing, got create=%d read=%d",
			parsedRecord.CacheCreationInputTokens, parsedRecord.CacheReadInputTokens)
	}
}
