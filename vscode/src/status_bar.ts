// Renders snapshots into a VS Code status-bar item.
import * as vscode from 'vscode';
import { Snapshot } from './snapshot';
import { formatCostUSD, formatPercentBar, formatTokensCompact } from './format';
import { renderTooltip } from './tooltip_renderer';

export class StatusBarRenderer {
  private readonly statusBarItem: vscode.StatusBarItem;

  constructor() {
    this.statusBarItem = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Right,
      100,
    );
    this.statusBarItem.text = 'Claude ⏳';
    this.statusBarItem.tooltip = 'Claude Hunter starting…';
    this.statusBarItem.show();
  }

  render(snapshot: Snapshot): void {
    const window5h = snapshot.window5h;
    const percentColour = pickColourForPercent(window5h.percentOfCeilingEstimate);
    const shortModel = shortenModelName(snapshot.model);

    this.statusBarItem.text =
      `$(pulse) ${shortModel} ` +
      `${window5h.percentOfCeilingEstimate.toFixed(0)}% ` +
      `${formatPercentBar(window5h.percentOfCeilingEstimate)} ` +
      `· ${formatTokensCompact(window5h.effectiveTokens)}/5h ` +
      `· ${formatCostUSD(window5h.costUSD)}`;
    this.statusBarItem.color = percentColour;
    this.statusBarItem.tooltip = renderTooltip(snapshot);
  }

  showError(message: string): void {
    this.statusBarItem.text = '$(alert) Claude Hunter';
    this.statusBarItem.tooltip = message;
  }

  dispose(): void {
    this.statusBarItem.dispose();
  }
}

function pickColourForPercent(percent: number): vscode.ThemeColor | undefined {
  if (percent >= 85) return new vscode.ThemeColor('errorForeground');
  if (percent >= 60) return new vscode.ThemeColor('editorWarning.foreground');
  return undefined;
}

function shortenModelName(model: string | undefined): string {
  if (!model) return 'Claude';
  return model.replace(/^claude-/, '').replace(/-\d{8}$/, '');
}
