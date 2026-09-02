// Spawns the claude-hunter binary and exposes a callback stream of snapshots.
import { ChildProcess, spawn } from 'child_process';
import { Snapshot } from './snapshot';

export interface HunterProcessOptions {
  binaryPath: string;
  projectsDir?: string;
  windowHours?: number;
  ceilingMillions?: number;
  onSnapshot: (snapshot: Snapshot) => void;
  onError: (message: string) => void;
  onExit: (exitCode: number | null) => void;
}

export class HunterProcess {
  private childProcess: ChildProcess | null = null;
  private partialStdoutBuffer = '';
  private stopRequested = false;

  constructor(private readonly options: HunterProcessOptions) {}

  start(): void {
    const commandArgs = this.buildCommandArgs();
    const spawnedChild = spawn(this.options.binaryPath, commandArgs, {
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    this.childProcess = spawnedChild;

    spawnedChild.stdout?.on('data', (chunk: Buffer) => this.consumeStdoutChunk(chunk));
    spawnedChild.stderr?.on('data', (chunk: Buffer) => {
      this.options.onError(chunk.toString('utf8').trimEnd());
    });
    spawnedChild.on('exit', (exitCode) => {
      if (this.stopRequested) return;
      this.options.onExit(exitCode);
    });
    spawnedChild.on('error', (spawnError) => {
      if (this.stopRequested) return;
      this.options.onError(spawnError.message);
    });
  }

  stop(): void {
    this.stopRequested = true;
    if (this.childProcess) {
      this.childProcess.kill('SIGTERM');
      this.childProcess = null;
    }
  }

  private buildCommandArgs(): string[] {
    const commandArgs: string[] = [];
    if (this.options.projectsDir) {
      commandArgs.push('--projects-dir', this.options.projectsDir);
    }
    if (typeof this.options.windowHours === 'number') {
      commandArgs.push('--window-hours', String(this.options.windowHours));
    }
    if (typeof this.options.ceilingMillions === 'number') {
      commandArgs.push('--ceiling-millions', String(this.options.ceilingMillions));
    }
    return commandArgs;
  }

  private consumeStdoutChunk(chunk: Buffer): void {
    this.partialStdoutBuffer += chunk.toString('utf8');
    const stdoutLines = this.partialStdoutBuffer.split('\n');
    this.partialStdoutBuffer = stdoutLines.pop() ?? '';
    for (const line of stdoutLines) {
      if (!line.trim()) continue;
      try {
        const decoded = JSON.parse(line) as Snapshot;
        if (decoded.type === 'snapshot') {
          this.options.onSnapshot(decoded);
        }
      } catch (parseError) {
        this.options.onError(`bad snapshot line: ${(parseError as Error).message}`);
      }
    }
  }
}
