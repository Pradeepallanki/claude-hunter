// Package usage models a single token-usage observation extracted from a
// Claude Code session file and exposes a parser for one JSONL line.
package usage

import (
	"encoding/json"
	"time"
)

// Record captures the token counts for a single assistant turn.
type Record struct {
	SessionID                string
	Model                    string
	Timestamp                time.Time
	InputTokens              int64
	OutputTokens             int64
	CacheCreationInputTokens int64
	CacheReadInputTokens     int64
}

type rawLine struct {
	Type      string    `json:"type"`
	SessionID string    `json:"sessionId"`
	Timestamp time.Time `json:"timestamp"`
	Message   *struct {
		Model string `json:"model"`
		Usage *struct {
			InputTokens              int64 `json:"input_tokens"`
			OutputTokens             int64 `json:"output_tokens"`
			CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// ParseLine returns the usage record from a single JSONL line, or nil if the
// line is not an assistant record with usage data. A non-nil error indicates
// the line could not be decoded as JSON.
func ParseLine(rawJSON []byte) (*Record, error) {
	var decoded rawLine
	if err := json.Unmarshal(rawJSON, &decoded); err != nil {
		return nil, err
	}
	if decoded.Type != "assistant" || decoded.Message == nil || decoded.Message.Usage == nil {
		return nil, nil
	}
	return &Record{
		SessionID:                decoded.SessionID,
		Model:                    decoded.Message.Model,
		Timestamp:                decoded.Timestamp,
		InputTokens:              decoded.Message.Usage.InputTokens,
		OutputTokens:             decoded.Message.Usage.OutputTokens,
		CacheCreationInputTokens: decoded.Message.Usage.CacheCreationInputTokens,
		CacheReadInputTokens:     decoded.Message.Usage.CacheReadInputTokens,
	}, nil
}
