// Renders snapshot details as a tabbed Webview panel opened from the status bar.
import * as vscode from 'vscode';
import { ProjectActivity, SessionActivity, Snapshot } from './snapshot';
import { formatCostUSD, formatTokensCompact } from './format';

export class DetailsPanel {
  private panel: vscode.WebviewPanel | null = null;
  private latestSnapshot: Snapshot | null = null;

  update(snapshot: Snapshot): void {
    this.latestSnapshot = snapshot;
    this.pushSnapshotToWebview(snapshot);
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
    this.panel.webview.html = renderShellHtml();
    this.panel.webview.onDidReceiveMessage((message) => {
      if (message?.command === 'close') this.panel?.dispose();
      if (message?.command === 'ready') this.pushSnapshotToWebview(this.latestSnapshot);
    });
    this.panel.onDidDispose(() => {
      this.panel = null;
    });
  }

  private pushSnapshotToWebview(snapshot: Snapshot | null): void {
    if (!this.panel || !snapshot) return;
    this.panel.webview.postMessage({
      command: 'snapshot',
      tabs: renderTabBodies(snapshot),
    });
  }

  dispose(): void {
    this.panel?.dispose();
    this.panel = null;
  }
}

function renderTabBodies(snapshot: Snapshot): Record<string, string> {
  return {
    overview: renderOverviewTab(snapshot),
    agents: renderAgentsTab(snapshot.agents ?? []),
    projects: renderProjectsTab(snapshot.projects ?? []),
    breakdown: renderBreakdownTab(snapshot),
  };
}

function renderShellHtml(): string {
  return `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <style>
      body {
        font-family: var(--vscode-font-family);
        color: var(--vscode-foreground);
        padding: 0;
        margin: 0;
        font-size: var(--vscode-font-size);
      }
      .tabbar {
        display: flex;
        border-bottom: 1px solid var(--vscode-panel-border);
        background: var(--vscode-editorGroupHeader-tabsBackground);
      }
      .tab {
        padding: 10px 18px;
        cursor: pointer;
        border: none;
        background: transparent;
        color: var(--vscode-foreground);
        opacity: 0.65;
        border-bottom: 2px solid transparent;
        font-size: var(--vscode-font-size);
        font-family: inherit;
      }
      .tab:hover { opacity: 0.9; }
      .tab.active {
        opacity: 1;
        border-bottom-color: var(--vscode-focusBorder);
      }
      .panel { display: none; padding: 16px 20px; }
      .panel.active { display: block; }
      h2 { margin: 0 0 4px 0; font-size: 1.1em; }
      h3 { margin: 16px 0 6px 0; font-size: 1em; }
      .subtitle { opacity: 0.75; margin-bottom: 16px; }
      .battery {
        font-size: 1.5em;
        margin-bottom: 8px;
        color: var(--vscode-charts-green);
      }
      .battery.warn { color: var(--vscode-editorWarning-foreground); }
      .battery.danger { color: var(--vscode-errorForeground); }
      .eta { opacity: 0.8; margin-bottom: 14px; }
      .banner {
        padding: 8px 12px;
        border-radius: 4px;
        margin: 4px 0 14px 0;
        background: var(--vscode-inputValidation-warningBackground);
        border: 1px solid var(--vscode-inputValidation-warningBorder);
        color: var(--vscode-foreground);
      }
      .banner.danger {
        background: var(--vscode-inputValidation-errorBackground);
        border-color: var(--vscode-inputValidation-errorBorder);
      }
      .stats-grid {
        display: grid;
        grid-template-columns: 1fr 1fr 1fr;
        gap: 12px;
        margin-bottom: 12px;
      }
      .stat {
        background: var(--vscode-editor-inactiveSelectionBackground);
        padding: 10px 12px;
        border-radius: 4px;
      }
      .stat .label { opacity: 0.7; font-size: 0.85em; }
      .stat .value { font-size: 1.2em; margin-top: 2px; font-variant-numeric: tabular-nums; }
      table { border-collapse: collapse; width: 100%; margin-bottom: 12px; }
      th, td {
        text-align: left;
        padding: 6px 8px;
        border-bottom: 1px solid var(--vscode-panel-border);
      }
      td.num, th.num { text-align: right; font-variant-numeric: tabular-nums; }
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
      .empty { opacity: 0.65; padding: 8px 0; }
      .footer {
        display: flex;
        justify-content: flex-end;
        padding: 12px 20px;
        border-top: 1px solid var(--vscode-panel-border);
      }
      button {
        padding: 6px 14px;
        background: var(--vscode-button-background);
        color: var(--vscode-button-foreground);
        border: none;
        border-radius: 2px;
        cursor: pointer;
      }
      button:hover { background: var(--vscode-button-hoverBackground); }
    </style>
  </head>
  <body>
    <div class="tabbar" role="tablist">
      <button class="tab active" data-target="overview">Overview</button>
      <button class="tab" data-target="agents">Agents</button>
      <button class="tab" data-target="projects">Projects</button>
      <button class="tab" data-target="breakdown">Breakdown</button>
    </div>
    <section id="overview" class="panel active"><div class="empty">Loading…</div></section>
    <section id="agents" class="panel"><div class="empty">Loading…</div></section>
    <section id="projects" class="panel"><div class="empty">Loading…</div></section>
    <section id="breakdown" class="panel"><div class="empty">Loading…</div></section>
    <div class="footer">
      <button id="close-button">Close</button>
    </div>
    <script>
      const vscodeApi = acquireVsCodeApi();
      document.getElementById('close-button').addEventListener('click', () => {
        vscodeApi.postMessage({ command: 'close' });
      });
      document.querySelectorAll('.tab').forEach((tabButton) => {
        tabButton.addEventListener('click', () => {
          document.querySelectorAll('.tab').forEach((t) => t.classList.remove('active'));
          document.querySelectorAll('.panel').forEach((p) => p.classList.remove('active'));
          tabButton.classList.add('active');
          document.getElementById(tabButton.dataset.target).classList.add('active');
        });
      });
      window.addEventListener('message', (event) => {
        const message = event.data;
        if (message?.command === 'snapshot' && message.tabs) {
          for (const tabId of Object.keys(message.tabs)) {
            const section = document.getElementById(tabId);
            if (section) section.innerHTML = message.tabs[tabId];
          }
        }
      });
      vscodeApi.postMessage({ command: 'ready' });
    </script>
  </body>
</html>`;
}

function renderOverviewTab(snapshot: Snapshot): string {
  const window5h = snapshot.window5h;
  const percentUsed = window5h.percentOfCeilingEstimate;
  const percentRemaining = Math.max(0, Math.min(100, 100 - percentUsed));
  const batteryClass =
    percentRemaining <= 10 ? 'danger' : percentRemaining <= 20 ? 'warn' : '';
  const etaLabel = formatEta(window5h.secondsToLimit);
  const modelLabel = escapeHtml(snapshot.model ?? 'unknown model');

  const contextWarning = (snapshot.agents ?? []).find(
    (agent) =>
      agent.contextWindowSize > 0 &&
      agent.contextTokens / agent.contextWindowSize >= 0.8,
  );
  const warningBanner = contextWarning
    ? `<div class="banner${
        contextWarning.contextTokens / contextWarning.contextWindowSize >= 0.95
          ? ' danger'
          : ''
      }">⚠ Session <code>${escapeHtml(contextWarning.sessionId.slice(0, 8))}</code> context is at ${Math.round(
        (contextWarning.contextTokens / contextWarning.contextWindowSize) * 100,
      )}% of ${formatTokensCompact(contextWarning.contextWindowSize)} — consider <code>/compact</code>.</div>`
    : '';

  return `
    <h2>Claude Hunter</h2>
    <div class="subtitle">${modelLabel}</div>
    <div class="battery ${batteryClass}">🔋 ${percentRemaining.toFixed(0)}% remaining</div>
    <div class="eta">${etaLabel} · ${percentUsed.toFixed(1)}% of 5h ceiling used</div>
    ${warningBanner}
    <div class="stats-grid">
      <div class="stat"><div class="label">Effective tokens</div><div class="value">${formatTokensCompact(window5h.effectiveTokens)}</div></div>
      <div class="stat"><div class="label">Cost</div><div class="value">${formatCostUSD(window5h.costUSD)}</div></div>
      <div class="stat"><div class="label">Burn rate</div><div class="value">${Math.round(window5h.burnRatePerMinute)} <span style="font-size:0.7em; opacity:0.7;">tok/min</span></div></div>
    </div>`;
}

function renderAgentsTab(agents: SessionActivity[]): string {
  if (agents.length === 0) {
    return '<h3>Active agents</h3><div class="empty">No sessions active in the last 5 minutes.</div>';
  }
  const rows = agents.map((agent) => renderAgentRow(agent)).join('');
  return `
    <h3>Active agents (${agents.length})</h3>
    <table>
      <tr>
        <th>Session</th>
        <th>Model</th>
        <th class="num">Context</th>
        <th class="num">5h tokens</th>
        <th class="num">Last active</th>
      </tr>
      ${rows}
    </table>`;
}

function renderAgentRow(agent: SessionActivity): string {
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
    `<tr><td><code>${shortId}</code>${subagentBadge}</td>` +
    `<td>${escapeHtml(shortenModel(agent.model))}</td>` +
    `<td class="num">${contextLabel}</td>` +
    `<td class="num">${formatTokensCompact(agent.totalTokens)}</td>` +
    `<td class="num">${relativeTime(agent.lastActiveAt)}</td></tr>`
  );
}

function renderProjectsTab(projects: ProjectActivity[]): string {
  if (projects.length === 0) {
    return '<h3>Projects</h3><div class="empty">No project activity in the current window.</div>';
  }
  const rows = projects
    .map(
      (project) =>
        `<tr>` +
        `<td>${escapeHtml(project.project)}</td>` +
        `<td class="num">${project.sessions}</td>` +
        `<td class="num">${formatTokensCompact(project.totalTokens)}</td>` +
        `<td class="num">${formatCostUSD(project.costUSD)}</td>` +
        `<td class="num">${relativeTime(project.lastActiveAt)}</td>` +
        `</tr>`,
    )
    .join('');
  return `
    <h3>Projects (${projects.length})</h3>
    <table>
      <tr>
        <th>Project</th>
        <th class="num">Sessions</th>
        <th class="num">Tokens</th>
        <th class="num">Cost</th>
        <th class="num">Last active</th>
      </tr>
      ${rows}
    </table>`;
}

function renderBreakdownTab(snapshot: Snapshot): string {
  const window5h = snapshot.window5h;
  const weightedCacheRead = Math.round(window5h.cacheReadTokens * 0.1);
  const perModelRows = (window5h.perModel ?? [])
    .map(
      (entry) =>
        `<tr><td>${escapeHtml(entry.model)}</td>` +
        `<td class="num">${formatTokensCompact(entry.effectiveTokens)}</td>` +
        `<td class="num">${formatCostUSD(entry.costUSD)}</td></tr>`,
    )
    .join('');

  return `
    <h3>Tokens</h3>
    <table>
      <tr><td>Input</td><td class="num">${formatTokensCompact(window5h.inputTokens)}</td></tr>
      <tr><td>Output</td><td class="num">${formatTokensCompact(window5h.outputTokens)}</td></tr>
      <tr><td>Cache write</td><td class="num">${formatTokensCompact(window5h.cacheCreateTokens)}</td></tr>
      <tr><td>Cache read</td><td class="num">${formatTokensCompact(window5h.cacheReadTokens)} (×0.1 → ${formatTokensCompact(weightedCacheRead)})</td></tr>
      <tr><td>Effective</td><td class="num">${formatTokensCompact(window5h.effectiveTokens)}</td></tr>
      <tr><td>Cost</td><td class="num">${formatCostUSD(window5h.costUSD)}</td></tr>
    </table>
    ${
      perModelRows
        ? `<h3>By model</h3>
    <table>
      <tr><th>Model</th><th class="num">Tokens</th><th class="num">Cost</th></tr>
      ${perModelRows}
    </table>`
        : ''
    }`;
}

function formatEta(secondsToLimit: number): string {
  if (secondsToLimit < 0) return 'ETA to limit: — (no burn detected)';
  if (secondsToLimit === 0) return 'ETA to limit: limit reached';
  const minutes = Math.round(secondsToLimit / 60);
  if (minutes < 60) return `≈${minutes}m to limit at current pace`;
  const hours = Math.floor(minutes / 60);
  const remainderMinutes = minutes % 60;
  return `≈${hours}h ${remainderMinutes}m to limit at current pace`;
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
