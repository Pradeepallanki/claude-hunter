// TooltipRenderer turns a Snapshot into the MarkdownString the VS Code
// status-bar item shows on hover.
import * as vscode from 'vscode';
import { Snapshot } from './snapshot';
import { formatCostUSD, formatTokensCompact } from './format';

export function renderTooltip(snapshot: Snapshot): vscode.MarkdownString {
  const window5h = snapshot.window5h;
  const weightedCacheRead = Math.round(window5h.cacheReadTokens * 0.1);

  const markdownTooltip = new vscode.MarkdownString();
  markdownTooltip.supportThemeIcons = true;
  markdownTooltip.isTrusted = false;

  markdownTooltip.appendMarkdown(
    `**Claude Hunter** · ${snapshot.model ?? 'unknown model'}\n\n`,
  );
  markdownTooltip.appendMarkdown(
    `**5h window** — ${formatTokensCompact(window5h.effectiveTokens)} effective ` +
      `(${window5h.percentOfCeilingEstimate.toFixed(1)}% of ceiling)\n\n`,
  );

  markdownTooltip.appendMarkdown('| | |\n|---|---|\n');
  markdownTooltip.appendMarkdown(
    `| Input | ${formatTokensCompact(window5h.inputTokens)} |\n` +
      `| Output | ${formatTokensCompact(window5h.outputTokens)} |\n` +
      `| Cache write | ${formatTokensCompact(window5h.cacheCreateTokens)} |\n` +
      `| Cache read | ${formatTokensCompact(window5h.cacheReadTokens)} ` +
      `(×0.1 → ${formatTokensCompact(weightedCacheRead)}) |\n` +
      `| Cost | ${formatCostUSD(window5h.costUSD)} |\n` +
      `| Burn rate | ${Math.round(window5h.burnRatePerMinute)} tok/min |\n`,
  );

  if (window5h.perModel && window5h.perModel.length > 0) {
    markdownTooltip.appendMarkdown('\n**By model**\n\n');
    markdownTooltip.appendMarkdown('| Model | Tokens | Cost |\n|---|---:|---:|\n');
    for (const modelEntry of window5h.perModel) {
      markdownTooltip.appendMarkdown(
        `| ${modelEntry.model} | ` +
          `${formatTokensCompact(modelEntry.effectiveTokens)} | ` +
          `${formatCostUSD(modelEntry.costUSD)} |\n`,
      );
    }
  }

  return markdownTooltip;
}
