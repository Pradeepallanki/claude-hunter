# Changelog

All notable changes to Claude Hunter are documented in this file.

The format is loosely based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

Nothing yet.

## [0.1.0] — 2026-09-02

Initial public release.

### Added

- **Core** (`core/`) — Go binary that watches `~/.claude/projects/`,
  aggregates a rolling 5-hour window of assistant token usage, and
  streams NDJSON snapshots to stdout.
  - `usage` package: JSONL parser for one assistant record.
  - `pricing` package: per-model rate table (Opus / Sonnet / Haiku)
    and USD cost calculator.
  - `window` package: rolling aggregation, effective-token weighting
    (`cache_read × 0.1`), burn rate over 10-minute sample, latest
    model by timestamp, per-model breakdown sorted by cost.
  - `watcher` package: `FileTailer` with partial-line buffering and
    `ProjectsObserver` on top of `fsnotify`.
  - `snapshot` package: NDJSON payload schema.
- **VS Code extension** (`vscode/`) — TypeScript status-bar widget.
  - Live model name, percent-of-ceiling gauge, effective tokens,
    cost.
  - Markdown tooltip with input / output / cache-write / cache-read
    breakdown and per-model attribution table.
  - Diagnostic output channel `Claude Hunter`.
  - Binary staging to `$TMPDIR/claude-hunter/<version>/` on
    activation to work around the macOS Ventura+ App Management
    execution restriction under `~/Desktop/` and
    `~/.vscode/extensions/`.
  - `stopRequested` flag prevents the deactivate SIGTERM from
    clobbering the freshly-activated widget with an "exited" error.
- **IntelliJ plugin** (`intellij/`) — Kotlin status-bar widget.
  - Feature parity with VS Code: status text and HTML tooltip
    breakdown including per-model attribution.
  - Extracts the bundled binary from JAR resources into
    `$TMPDIR/claude-hunter/<version>/` on load.
  - Built against IntelliJ Platform 2024.1 via the JetBrains
    IntelliJ Platform Gradle Plugin v2.
- **Documentation** — README, installing, configuration, building,
  architecture, troubleshooting, contributing, and the original
  design plan.
- **MIT License**.

### Known limitations

- Plan ceiling is an estimate; recalibrate for your subscription — see
  [`docs/CONFIGURATION.md`](docs/CONFIGURATION.md#plan-ceilings).
- IntelliJ settings are exposed only via JVM system properties for
  now.
- No persistent history — restarts re-scan on-disk JSONL.

[Unreleased]: about:blank
[0.1.0]: about:blank
