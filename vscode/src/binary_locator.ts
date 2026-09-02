// Locates the claude-hunter binary and returns a path safe to execute.
// macOS blocks execution of binaries that live inside ~/.vscode/extensions/…
// (Ventura+ App Management), so we stage a copy in the OS temp dir and
// return that copy's path.
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
  return stagedPath;
}

function isStagedCopyCurrent(sourcePath: string, stagedPath: string): boolean {
  if (!fs.existsSync(stagedPath)) return false;
  const sourceStats = fs.statSync(sourcePath);
  const stagedStats = fs.statSync(stagedPath);
  return sourceStats.size === stagedStats.size && stagedStats.mtimeMs >= sourceStats.mtimeMs;
}
