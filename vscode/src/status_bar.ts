// Renders snapshots into a VS Code status-bar item.
import * as vscode from 'vscode';
import { Snapshot } from './snapshot';
import { formatCostUSD, formatTokensCompact } from './format';
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
    const percentRemaining = Math.max(
      0,
      Math.min(100, 100 - window5h.percentOfCeilingEstimate),
    );
    const batteryColour = pickBatteryColour(percentRemaining);
    const shortModel = shortenModelName(snapshot.model);

    this.statusBarItem.text =
      `$(pulse) ${shortModel} ` +
      `🔋 ${percentRemaining.toFixed(0)}% ` +
      `· ${formatTokensCompact(window5h.effectiveTokens)}/5h ` +
      `· ${formatCostUSD(window5h.costUSD)}`;
    this.statusBarItem.color = batteryColour;
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

function pickBatteryColour(percentRemaining: number): vscode.ThemeColor {
  if (percentRemaining <= 10) return new vscode.ThemeColor('errorForeground');
  if (percentRemaining <= 20) return new vscode.ThemeColor('editorWarning.foreground');
  return new vscode.ThemeColor('charts.green');
}

function shortenModelName(model: string | undefined): string {
  if (!model) return 'Claude';
  return model.replace(/^claude-/, '').replace(/-\d{8}$/, '');
}
