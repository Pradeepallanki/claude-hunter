// Renders snapshot details as a Webview panel opened from the status bar.
import * as vscode from 'vscode';
import { Snapshot } from './snapshot';
import { formatCostUSD, formatTokensCompact } from './format';

export class DetailsPanel {
  private panel: vscode.WebviewPanel | null = null;
  private latestSnapshot: Snapshot | null = null;

  update(snapshot: Snapshot): void {
    this.latestSnapshot = snapshot;
    if (this.panel) {
      this.panel.webview.html = renderHtml(snapshot);
    }
  }

  show(): void {
    if (!this.latestSnapshot) {
      vscode.window.showInformationMessage('Claude Hunter is still starting…');
      return;
    }
    if (this.panel) {
      this.panel.reveal(vscode.ViewColumn.Beside, true);
      return;
    }
    this.panel = vscode.window.createWebviewPanel(
      'claudeHunterDetails',
      'Claude Hunter',
      { viewColumn: vscode.ViewColumn.Beside, preserveFocus: true },
      { enableScripts: true, retainContextWhenHidden: false },
    );
    this.panel.webview.html = renderHtml(this.latestSnapshot);
    this.panel.webview.onDidReceiveMessage((message) => {
      if (message?.command === 'close') this.panel?.dispose();
    });
    this.panel.onDidDispose(() => {
      this.panel = null;
    });
  }

  dispose(): void {
    this.panel?.dispose();
    this.panel = null;
  }
}

function renderHtml(snapshot: Snapshot): string {
  const window5h = snapshot.window5h;
  const percentUsed = window5h.percentOfCeilingEstimate;
  const percentRemaining = Math.max(0, Math.min(100, 100 - percentUsed));
  const weightedCacheRead = Math.round(window5h.cacheReadTokens * 0.1);
  const modelLabel = escapeHtml(snapshot.model ?? 'unknown model');

  const perModelRows = (window5h.perModel ?? [])
    .map(
      (entry) =>
        `<tr><td>${escapeHtml(entry.model)}</td>` +
        `<td class="num">${formatTokensCompact(entry.effectiveTokens)}</td>` +
        `<td class="num">${formatCostUSD(entry.costUSD)}</td></tr>`,
    )
    .join('');

  const agents = snapshot.agents ?? [];
  const agentRows = agents
    .map((agent) => {
      const contextPct =
        agent.contextWindowSize > 0
          ? Math.round((agent.contextTokens / agent.contextWindowSize) * 100)
          : 0;
      const contextLabel =
        agent.contextWindowSize > 0
          ? `${formatTokensCompact(agent.contextTokens)} / ${formatTokensCompact(agent.contextWindowSize)} (${contextPct}%)`
          : formatTokensCompact(agent.contextTokens);
      const shortId = escapeHtml(agent.sessionId.slice(0, 8));
      const subagentBadge =
        agent.sidechainTurns > 0
          ? `<span class="badge">+${agent.sidechainTurns} subagent</span>`
          : '';
      return (
        `<tr><td><code>${shortId}</code> ${subagentBadge}</td>` +
        `<td>${escapeHtml(shortenModel(agent.model))}</td>` +
        `<td class="num">${contextLabel}</td>` +
        `<td class="num">${formatTokensCompact(agent.totalTokens)}</td>` +
        `<td class="num">${relativeTime(agent.lastActiveAt)}</td></tr>`
      );
    })
    .join('');

  return `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <style>
      body {
        font-family: var(--vscode-font-family);
        color: var(--vscode-foreground);
        padding: 16px 20px;
        font-size: var(--vscode-font-size);
      }
      h2 { margin: 0 0 4px 0; font-size: 1.1em; }
      .subtitle { opacity: 0.75; margin-bottom: 16px; }
      .battery {
        font-size: 1.3em;
        margin-bottom: 16px;
        color: var(--vscode-charts-green);
      }
      table { border-collapse: collapse; width: 100%; margin-bottom: 12px; }
      th, td {
        text-align: left;
        padding: 4px 8px;
        border-bottom: 1px solid var(--vscode-panel-border);
      }
      td.num, th.num { text-align: right; font-variant-numeric: tabular-nums; }
      h3 { margin: 16px 0 4px 0; font-size: 1em; }
      button {
        margin-top: 12px;
        padding: 6px 14px;
        background: var(--vscode-button-background);
        color: var(--vscode-button-foreground);
        border: none;
        border-radius: 2px;
        cursor: pointer;
      }
      button:hover { background: var(--vscode-button-hoverBackground); }
      code { font-family: var(--vscode-editor-font-family); font-size: 0.9em; opacity: 0.85; }
      .badge {
        display: inline-block;
        padding: 1px 6px;
        margin-left: 6px;
        font-size: 0.75em;
        border-radius: 8px;
        background: var(--vscode-badge-background);
        color: var(--vscode-badge-foreground);
      }
    </style>
  </head>
  <body>
    <h2>Claude Hunter</h2>
    <div class="subtitle">${modelLabel}</div>
    <div class="battery">🔋 ${percentRemaining.toFixed(0)}% remaining
      <span style="opacity:0.6; font-size:0.85em;">
        (${percentUsed.toFixed(1)}% of 5h ceiling used)
      </span>
    </div>
    <table>
      <tr><td>Input</td><td class="num">${formatTokensCompact(window5h.inputTokens)}</td></tr>
      <tr><td>Output</td><td class="num">${formatTokensCompact(window5h.outputTokens)}</td></tr>
      <tr><td>Cache write</td><td class="num">${formatTokensCompact(window5h.cacheCreateTokens)}</td></tr>
      <tr><td>Cache read</td><td class="num">${formatTokensCompact(window5h.cacheReadTokens)} (×0.1 → ${formatTokensCompact(weightedCacheRead)})</td></tr>
      <tr><td>Effective</td><td class="num">${formatTokensCompact(window5h.effectiveTokens)}</td></tr>
      <tr><td>Cost</td><td class="num">${formatCostUSD(window5h.costUSD)}</td></tr>
      <tr><td>Burn rate</td><td class="num">${Math.round(window5h.burnRatePerMinute)} tok/min</td></tr>
    </table>
    ${
      agentRows
        ? `<h3>Active agents (${agents.length})</h3>
    <table>
      <tr><th>Session</th><th>Model</th><th class="num">Context</th><th class="num">5h tokens</th><th class="num">Last active</th></tr>
      ${agentRows}
    </table>`
        : '<h3>Active agents</h3><div style="opacity:0.7; margin-bottom:12px;">No sessions active in the last 5 minutes.</div>'
    }
    ${
      perModelRows
        ? `<h3>By model</h3>
    <table>
      <tr><th>Model</th><th class="num">Tokens</th><th class="num">Cost</th></tr>
      ${perModelRows}
    </table>`
        : ''
    }
    <button id="close-button">Close</button>
    <script>
      const vscodeApi = acquireVsCodeApi();
      document.getElementById('close-button').addEventListener('click', () => {
        vscodeApi.postMessage({ command: 'close' });
      });
    </script>
  </body>
</html>`;
}

function shortenModel(model: string): string {
  if (!model) return 'unknown';
  return model.replace(/^claude-/, '').replace(/-\d{8}$/, '');
}

function relativeTime(iso: string): string {
  const then = Date.parse(iso);
  if (Number.isNaN(then)) return '—';
  const seconds = Math.max(0, Math.round((Date.now() - then) / 1000));
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.round(minutes / 60);
  return `${hours}h ago`;
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}
