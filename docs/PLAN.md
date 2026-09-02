# Claude Hunter — Live Token Usage Monitor

A lightweight, cross-IDE (VS Code + IntelliJ) status-bar extension that shows Claude Code token usage in real time so you never hit "usage limit exceeded" mid-flow.

---

## 1. Problem

Claude Code silently accumulates tokens across a conversation until it hits a hard rolling limit (5-hour window on Max plans, plus daily / weekly ceilings). Discovering the limit *after* it trips wastes the current context and interrupts work. We want an always-visible readout that updates within a second or two of every assistant turn.

## 2. Data Source (no API calls, purely local)

Claude Code writes every turn to `~/.claude/projects/<slug>/<sessionId>.jsonl`. Assistant lines carry a `message.usage` object:

```json
{
  "message": {
    "model": "claude-opus-4-7",
    "usage": {
      "input_tokens": 6,
      "output_tokens": 135,
      "cache_creation_input_tokens": 24794,
      "cache_read_input_tokens": 0
    }
  },
  "timestamp": "2026-08-20T04:38:24.503Z",
  "sessionId": "..."
}
```

Everything we need is on disk. No network, no auth, no rate limits.

## 3. Architecture

```
+-------------------------+       spawn (stdin/stdout NDJSON)       +----------------------+
|  VS Code extension (TS) | <---------------------------------->    |                      |
+-------------------------+                                         |  claude-hunter (Go)  |
                                                                    |  - fsnotify watcher  |
+-------------------------+       spawn (stdin/stdout NDJSON)       |  - JSONL tailer      |
|  IntelliJ plugin (Kt)   | <---------------------------------->    |  - aggregator        |
+-------------------------+                                         +----------------------+
                                                                             |
                                                                             v
                                                                    ~/.claude/projects/**/*.jsonl
```

**Single Go core binary + two thin IDE clients.**

### Why Go for the core?

| Criterion | Go | Rust | Node | Python |
|---|---|---|---|---|
| Single static binary | yes | yes | no (needs runtime) | no |
| Startup latency | ~5 ms | ~5 ms | ~150 ms | ~200 ms |
| Memory footprint | ~10 MB | ~5 MB | ~40 MB | ~30 MB |
| fsnotify cross-platform | first-class (`fsnotify/fsnotify`) | good | ok | ok |
| Dev velocity | high | medium | high | high |
| Distribution size | ~6 MB | ~3 MB | bundled Node ~50 MB | interpreter needed |

Go wins on the balance of speed, footprint, single-binary distribution, and dev velocity. Rust is marginally leaner but slower to iterate for an MVP; Node loses on cold-start; Python isn't a fit for a background daemon.

### Why not just do it in the extension's native language?

Two reasons:
1. **No duplicated aggregation logic** across TS and Kotlin — the parser, cache-token math, and 5h-window rolling logic live in one place.
2. **Consistent behavior**: both IDEs render identical numbers because they read the same stream.

### Communication protocol

The IDE extension spawns `claude-hunter --watch` as a child process. The core writes newline-delimited JSON to stdout:

```json
{"type":"snapshot","window5h":{"input":123456,"output":45678,"cacheRead":900000,"cacheCreate":200000,"tokensEffective":389134,"pctOfMaxPlanEst":42.1,"windowStart":"2026-09-01T10:00:00Z","windowEnd":"2026-09-01T15:00:00Z"},"session":{"id":"...","tokensEffective":58231,"cost":0.87},"model":"claude-opus-4-7","ts":"2026-09-01T14:22:07Z"}
```

Emits an initial `snapshot` on start, then a fresh snapshot within ~200 ms of any JSONL append. NDJSON over stdout means:
- No TCP ports, no firewall prompts, no port collisions between two running IDEs.
- Trivial to test (`./claude-hunter --watch | jq`).
- Extensions just do `readline → JSON.parse → update status bar`.

## 4. Aggregation Logic

**Effective tokens** (what actually counts against the rate limit) uses Anthropic's documented weighting:

```
tokensEffective = input_tokens
                + output_tokens
                + cache_creation_input_tokens        // full price
                + cache_read_input_tokens * 0.1      // ~10% weight
```

Model-specific pricing table hard-coded (opus / sonnet / haiku, cache-write & cache-read rates) — small and rarely changes; can be tweaked in a config file later.

**5-hour rolling window**: aggregate all assistant records with `timestamp >= now - 5h`, across all sessions in `~/.claude/projects/`. Claude Code's rolling limit is global, not per-session.

**Plan estimation**: we don't know the user's plan, so we show:
- Absolute effective tokens in the current 5h window
- Percentage against a configurable ceiling (defaults tuned for Max 20x; user can override in settings)
- A "burn rate" (tokens/min over last 10 min) so the user can see whether they'll hit the ceiling before the window rolls over

## 5. File Watching Strategy

- On startup: scan `~/.claude/projects/**/*.jsonl`, record `(path, size)` for each. Only process the last N days of data initially (default 7) to keep first-paint fast.
- Register a recursive `fsnotify` watcher on `~/.claude/projects/`.
- On `WRITE`: seek to the previously recorded offset, read new bytes, split on `\n`, parse each JSON line, update in-memory aggregates, emit fresh snapshot. This is O(bytes appended), not O(file size) — a live conversation adds tens of KB per turn, so update latency stays well under 200 ms.
- On `CREATE` (new session file): add to watch set, offset = 0.
- Debounce snapshot emits to 100 ms so a burst of writes coalesces into one status-bar update.

## 6. IDE Client Surfaces

### VS Code (TypeScript)

- Status-bar item, right-aligned, high priority.
- Text: `Claude 42% ▓▓▓▓░░░░░░ · 389k/5h · $12.30`
- Colors: green < 60%, yellow 60–85%, red > 85%.
- Tooltip: full breakdown (input / output / cache-read / cache-write, per-model, top session).
- Click: opens a webview panel with a sparkline of the last 5 hours, per-project totals, and the current session breakdown.
- Command: `Claude Hunter: Reset Window Estimate` (lets user recalibrate).

### IntelliJ (Kotlin)

- `StatusBarWidgetFactory` → custom widget bottom-right.
- Same text, tooltip, click-to-open-tool-window as VS Code.
- Tool window mirrors the webview: sparkline + breakdown.

Both extensions:
- On activation: locate bundled `claude-hunter` binary for the current OS/arch (we ship darwin-arm64, darwin-x64, linux-x64, linux-arm64, win-x64).
- Spawn it, pipe stdout through a line reader, update UI on each event.
- On IDE shutdown: SIGTERM the child, it flushes and exits.

## 7. Performance Budget

| Metric | Target |
|---|---|
| Core binary cold start | < 50 ms |
| Time from JSONL write → status bar update | < 300 ms |
| Steady-state CPU (idle) | < 0.1% |
| Steady-state RSS | < 20 MB |
| Extension activation overhead | < 100 ms |

## 8. Project Layout

```
claude-hunter/
├── docs/
│   └── PLAN.md                    (this file)
├── core/                          Go module: `claude-hunter` binary
│   ├── main.go
│   ├── watcher.go
│   ├── parser.go
│   ├── aggregator.go
│   ├── pricing.go
│   └── go.mod
├── vscode/                        VS Code extension
│   ├── src/extension.ts
│   ├── package.json
│   └── tsconfig.json
├── intellij/                      IntelliJ plugin (Gradle + Kotlin)
│   ├── src/main/kotlin/...
│   └── build.gradle.kts
└── bin/                           prebuilt core binaries (per OS/arch), bundled into each extension
```

## 9. Milestones

1. **Core MVP**: watcher + parser + aggregator + NDJSON emitter, hand-tested via `| jq`.
2. **VS Code MVP**: status bar reading the stream, tooltip breakdown.
3. **IntelliJ MVP**: status bar widget consuming the same stream.
4. **Polish**: webview / tool-window with sparkline, config for plan ceiling, burn rate warning.
5. **Distribution**: `.vsix` for VS Code Marketplace, `.zip` plugin for JetBrains Marketplace.

Milestones 1–3 constitute the "usable today" version. 4–5 come after you've lived with it for a day and told me what's missing.

## 10. Open Questions for Review

- **Plan ceiling default**: I'll seed it with a Max-20x-ish number. Confirm your plan so we set a sensible default? (User can override anyway.)
- **Multi-project vs current-project scope**: default is global (all projects). Want a toggle to scope to the currently open workspace/project? Slightly complicates the IDE side.
- **History persistence**: current design keeps aggregates in-memory only. If the daemon restarts mid-window, it re-scans on-disk JSONL and rebuilds — no separate DB. OK?
- **Cost display**: hard-code pricing table, or omit dollar figures entirely and show tokens only? Pricing tables drift.

Awaiting your review — reply with any changes and I'll start on Milestone 1.
