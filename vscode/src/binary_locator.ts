// Locates the claude-hunter binary and returns a path safe to execute.
// macOS blocks execution of binaries that live inside ~/.vscode/extensions/…
// (Ventura+ App Management), so we stage a copy in the OS temp dir and
// return that copy's path.
import { execFileSync } from 'child_process';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';

const STAGED_BINARY_VERSION_TAG = 'v0.1.0';

export function locateHunterBinary(configuredPath: string, extensionRoot: string): string | null {
  if (configuredPath && fs.existsSync(configuredPath)) {
    return configuredPath;
  }
  const bundledBinaryPath = resolveBundledBinaryPath(extensionRoot);
  if (!bundledBinaryPath) {
    return null;
  }
  return stageBinaryForExecution(bundledBinaryPath);
}

function resolveBundledBinaryPath(extensionRoot: string): string | null {
  const platformFolder = `${process.platform}-${process.arch}`;
  const binaryName = process.platform === 'win32' ? 'claude-hunter.exe' : 'claude-hunter';
  const bundledPath = path.join(extensionRoot, 'bin', platformFolder, binaryName);
  return fs.existsSync(bundledPath) ? bundledPath : null;
}

function stageBinaryForExecution(sourcePath: string): string {
  const stageDirectory = path.join(os.tmpdir(), 'claude-hunter', STAGED_BINARY_VERSION_TAG);
  fs.mkdirSync(stageDirectory, { recursive: true });
  const stagedPath = path.join(stageDirectory, path.basename(sourcePath));

  if (isStagedCopyCurrent(sourcePath, stagedPath)) {
    return stagedPath;
  }
  fs.copyFileSync(sourcePath, stagedPath);
  fs.chmodSync(stagedPath, 0o755);
  reSignAdHocOnMacOS(stagedPath);
  return stagedPath;
}

// macOS invalidates a Mach-O linker signature when a Go binary is copied out
// of a location protected by App Management (e.g. ~/.vscode/extensions/), and
// launchd will silently SIGKILL the resulting process. Re-signing ad-hoc on
// the staged copy restores a signature the OS accepts.
function reSignAdHocOnMacOS(stagedPath: string): void {
  if (process.platform !== 'darwin') return;
  try {
    execFileSync('codesign', ['--force', '-s', '-', stagedPath], { stdio: 'ignore' });
  } catch {
    // codesign is part of the standard macOS command-line tools; if it's
    // missing, we surface the underlying spawn failure via the process exit
    // path rather than blocking activation.
  }
}

function isStagedCopyCurrent(sourcePath: string, stagedPath: string): boolean {
  if (!fs.existsSync(stagedPath)) return false;
  const sourceStats = fs.statSync(sourcePath);
  const stagedStats = fs.statSync(stagedPath);
  return sourceStats.size === stagedStats.size && stagedStats.mtimeMs >= sourceStats.mtimeMs;
}
