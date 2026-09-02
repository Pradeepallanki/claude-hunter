// Formatting helpers for status-bar and tooltip rendering.

export function formatTokensCompact(tokens: number): string {
  if (tokens >= 1_000_000) {
    return (tokens / 1_000_000).toFixed(1) + 'M';
  }
  if (tokens >= 1_000) {
    return Math.round(tokens / 1_000) + 'k';
  }
  return String(tokens);
}

export function formatPercentBar(percent: number, widthCells = 10): string {
  const clamped = Math.max(0, Math.min(100, percent));
  const filledCells = Math.round((clamped / 100) * widthCells);
  return '▓'.repeat(filledCells) + '░'.repeat(widthCells - filledCells);
}

export function formatCostUSD(usd: number): string {
  if (usd < 0.01) return '$0';
  if (usd < 1) return '$' + usd.toFixed(2);
  return '$' + usd.toFixed(2);
}
