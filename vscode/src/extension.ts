// VS Code activation entrypoint. Wires the hunter process to the status bar.
import * as vscode from 'vscode';
import { HunterProcess } from './hunter_process';
import { StatusBarRenderer } from './status_bar';
import { DetailsPanel } from './details_panel';
import { locateHunterBinary } from './binary_locator';

let hunterProcess: HunterProcess | null = null;
let statusBarRenderer: StatusBarRenderer | null = null;
let detailsPanel: DetailsPanel | null = null;
let diagnosticLog: vscode.OutputChannel | null = null;
let snapshotCount = 0;

export function activate(extensionContext: vscode.ExtensionContext): void {
  diagnosticLog = vscode.window.createOutputChannel('Claude Hunter');
  diagnosticLog.appendLine(`[activate] extensionPath=${extensionContext.extensionPath}`);

  statusBarRenderer = new StatusBarRenderer();
  detailsPanel = new DetailsPanel();
  extensionContext.subscriptions.push(
    vscode.commands.registerCommand('claudeHunter.showDetails', () => {
      detailsPanel?.show();
    }),
  );

  const settings = vscode.workspace.getConfiguration('claudeHunter');
  const configuredBinaryPath = settings.get<string>('binaryPath', '');
  const resolvedBinaryPath = locateHunterBinary(configuredBinaryPath, extensionContext.extensionPath);
  diagnosticLog.appendLine(`[activate] binary=${resolvedBinaryPath ?? 'NOT FOUND'}`);

  if (!resolvedBinaryPath) {
    statusBarRenderer.showError(
      'claude-hunter binary not found. Set claudeHunter.binaryPath or bundle a binary under bin/<platform>-<arch>/.',
    );
    return;
  }

  snapshotCount = 0;
  hunterProcess = new HunterProcess({
    binaryPath: resolvedBinaryPath,
    projectsDir: settings.get<string>('projectsDir', '') || undefined,
    windowHours: settings.get<number>('windowHours'),
    ceilingMillions: settings.get<number>('ceilingMillions'),
    onSnapshot: (snapshot) => {
      snapshotCount += 1;
      if (snapshotCount === 1 || snapshotCount % 40 === 0) {
        diagnosticLog?.appendLine(
          `[snapshot #${snapshotCount}] model=${snapshot.model} effective=${snapshot.window5h.effectiveTokens}`,
        );
      }
      statusBarRenderer?.render(snapshot);
      detailsPanel?.update(snapshot);
    },
    onError: (message) => diagnosticLog?.appendLine(`[stderr] ${message}`),
    onExit: (exitCode) => {
      diagnosticLog?.appendLine(`[exit] code=${exitCode ?? 'unknown'}`);
      statusBarRenderer?.showError(`claude-hunter exited (code ${exitCode ?? 'unknown'})`);
    },
  });
  hunterProcess.start();
  diagnosticLog.appendLine('[activate] hunter process started');

  extensionContext.subscriptions.push({
    dispose: () => {
      hunterProcess?.stop();
      hunterProcess = null;
    },
  });
}

export function deactivate(): void {
  hunterProcess?.stop();
  hunterProcess = null;
  statusBarRenderer?.dispose();
  statusBarRenderer = null;
  detailsPanel?.dispose();
  detailsPanel = null;
  diagnosticLog?.dispose();
  diagnosticLog = null;
}
