// Snapshot mirrors the NDJSON payload emitted by the claude-hunter core.
export interface ModelBreakdown {
  model: string;
  inputTokens: number;
  outputTokens: number;
  cacheCreateTokens: number;
  cacheReadTokens: number;
  effectiveTokens: number;
  costUSD: number;
}

export interface WindowSummary {
  inputTokens: number;
  outputTokens: number;
  cacheCreateTokens: number;
  cacheReadTokens: number;
  effectiveTokens: number;
  costUSD: number;
  burnRatePerMinute: number;
  windowStart: string;
  windowEnd: string;
  percentOfCeilingEstimate: number;
  perModel: ModelBreakdown[];
}

export interface Snapshot {
  type: 'snapshot';
  ts: string;
  model?: string;
  window5h: WindowSummary;
}
