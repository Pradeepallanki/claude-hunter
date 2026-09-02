# Architecture

Claude Hunter is deliberately small. This document explains the moving
parts so a new contributor can find their way to any behaviour in one
or two hops.

For the original design brief see [`PLAN.md`](PLAN.md). For rules
every change must obey see [`coding_standards.md`](coding_standards.md).

---

## 1. High-level shape

```
+----------------------------+       spawn (stdin/stdout NDJSON)         +----------------------+
|  VS Code extension (TS)    | <---------------------------------->      |                      |
+----------------------------+                                           |  claude-hunter (Go)  |
                                                                         |  - fsnotify watcher  |
+----------------------------+       spawn (stdin/stdout NDJSON)         |  - JSONL tailer      |
|  IntelliJ plugin (Kotlin)  | <---------------------------------->      |  - aggregator        |
+----------------------------+                                           +----------------------+
                                                                                    |
                                                                                    v
                                                                        ~/.claude/projects/**/*.jsonl
```

Two properties are load-bearing:

1. **One Go binary, two IDE clients.** All parsing, cost math,
   window arithmetic, and rate-limit weighting lives in one place.
   The extensions render only.
2. **Newline-delimited JSON over stdout.** The extensions spawn the
   binary as a subprocess and read lines. No sockets, no ports, no
   authentication.

The result: a status-bar readout that lags the JSONL write by
typically well under 300 ms and never touches the network.

---

## 2. Data source

Claude Code writes every conversational turn to
`~/.claude/projects/<slug>/<sessionId>.jsonl`. Assistant records carry
a `message.usage` object with `input_tokens`, `output_tokens`,
`cache_creation_input_tokens`, and `cache_read_input_tokens`. Claude
Hunter reads these files only — never Anthropic's API.

---

## 3. Core binary (`core/`)

Five feature-oriented Go packages. Each file has one responsibility;
see the [coding standards](coding_standards.md) for why.

| Package | Responsibility |
|---|---|
| `usage` | Parse a single JSONL line into a `Record`. |
| `pricing` | Per-model rate table + cost calculator. |
| `window` | Rolling 5-hour aggregation, per-model breakdown, burn rate, latest model. |
| `watcher` | fsnotify observer + per-file tailer with partial-line buffering. |
| `snapshot` | The NDJSON payload shape emitted to IDE clients. |

The `main.go` entrypoint wires them together:

1. Spawn `watcher.ProjectsObserver` in a goroutine.
2. For each `LineEvent` on the observer channel, `usage.ParseLine`
   into a `Record` (nil for non-assistant records), then
   `window.RollingWindow.Add`.
3. On every emit-interval tick, prune records older than the window,
   ask the window for totals, per-model breakdown, latest model, and
   burn rate, marshal into a `snapshot.Snapshot`, and write with
   `json.NewEncoder(os.Stdout).Encode`.

### Effective tokens

Anthropic's rate limit does not count cache-read tokens at their raw
value. The window applies the documented 10× discount:

```
effective = input + output + cache_creation + (cache_read × 0.1)
```

That's the number `window5h.effectiveTokens` reports, and the number
`window5h.percentOfCeilingEstimate` divides against
`--ceiling-millions`.

### Rolling window

- New records get appended; records may arrive out of order (seed of
  existing files happens after the observer is set up).
- `PruneBefore(cutoff)` scans the full slice — cheap in practice
  because the window rarely exceeds a few thousand records.
- `LatestModel()` returns the model of the record with the greatest
  timestamp. This is what the status bar shows so a mid-task switch
  from Opus to Haiku is reflected within one snapshot cycle.

---

## 4. Watcher

Two files, two responsibilities.

- `watcher/tail.go` — `FileTailer` tracks a byte offset per file, so
  re-opening on each `WRITE` event is `O(bytes appended)`, not
  `O(file size)`. Partial trailing bytes are buffered until the next
  newline.
- `watcher/observe.go` — `ProjectsObserver` seeds the tree with
  `filepath.WalkDir`, registers every discovered directory with
  `fsnotify.Watcher`, and dispatches lines to a buffered channel that
  `main.go` consumes.

macOS-specific note: `fsnotify` does not recurse. New session
subdirectories are watched by handling `CREATE` events on the parent.

---

## 5. Snapshot protocol

Each line on stdout is one JSON object. Fields, as of `0.1.0`:

```json
{
  "type": "snapshot",
  "ts": "2026-09-02T05:38:53.7Z",
  "model": "claude-opus-4-7",
  "window5h": {
    "inputTokens": 0,
    "outputTokens": 0,
    "cacheCreateTokens": 0,
    "cacheReadTokens": 0,
    "effectiveTokens": 0,
    "costUSD": 0,
    "burnRatePerMinute": 0,
    "windowStart": "2026-09-02T00:38:53.7Z",
    "windowEnd":   "2026-09-02T05:38:53.7Z",
    "percentOfCeilingEstimate": 0,
    "perModel": [
      { "model": "claude-opus-4-7", "effectiveTokens": 0, "costUSD": 0, "…": "…" }
    ]
  }
}
```

The protocol is intentionally versionless right now. Fields may be
added in a backwards-compatible manner; unknown fields are ignored
by both extensions. Renames or removals will require a bump.

---

## 6. VS Code extension (`vscode/`)

TypeScript, compiled to `out/`. Files reflect single responsibilities.

| File | Responsibility |
|---|---|
| `extension.ts` | Activation entrypoint, output-channel logging, wire-up. |
| `hunter_process.ts` | Spawn the core binary, buffer stdout, parse NDJSON. |
| `binary_locator.ts` | Find the bundled binary, stage it to `$TMPDIR` (macOS App-Management workaround). |
| `snapshot.ts` | TypeScript interfaces mirroring the JSON payload. |
| `status_bar.ts` | Populate the `StatusBarItem`; pick colour by percent. |
| `tooltip_renderer.ts` | `MarkdownString` builder for the hover tooltip. |
| `format.ts` | Compact token / percent / cost formatters. |

### Binary staging on macOS

macOS Ventura+ silently blocks execution of binaries that live under
protected locations such as `~/Desktop/` and `~/.vscode/extensions/`
when spawned from within a notarised app. On activation, the extension
copies the bundled binary out to `$TMPDIR/claude-hunter/<version>/`,
`chmod +x`es it, and spawns from there. See
[`TROUBLESHOOTING.md`](TROUBLESHOOTING.md#stuck-on-claude--macos)
if this ever breaks.

---

## 7. IntelliJ plugin (`intellij/`)

Kotlin + Gradle. Uses the JetBrains IntelliJ Platform Gradle Plugin v2.

| File | Responsibility |
|---|---|
| `StatusWidgetFactory.kt` | Registers the widget with the platform. |
| `StatusWidget.kt` | `StatusBarWidget.TextPresentation`; ties HunterProcess to the widget. |
| `HunterProcess.kt` | Spawns the core binary; two daemon threads for stdout / stderr. |
| `SnapshotParser.kt` | Gson-based decode of one JSONL line into `Snapshot`. |
| `Snapshot.kt` | Data classes mirroring the JSON payload. |
| `SnapshotFormat.kt` | Compact formatters (parallel to VS Code's `format.ts`). |
| `TooltipRenderer.kt` | HTML string builder (Swing tooltips need HTML for line breaks). |
| `BinaryLocator.kt` | Extracts the bundled binary from JAR resources into `$TMPDIR/claude-hunter/<version>/`. |

Bundled binaries live under `src/main/resources/bin/<platform>-<arch>/`
and are packaged into the plugin JAR as resources.

---

## 8. Future work

Non-committal ideas the architecture already leaves room for:

- **History persistence.** In-memory only today; a small append-only
  log would allow the widget to survive restarts without re-scanning
  every JSONL from disk.
- **Warn threshold hooks.** Fire a notification when
  `percentOfCeilingEstimate` crosses a user-defined line.
- **Per-project scoping.** Emit a separate snapshot for the current
  workspace when the IDE asks for it.
- **Panel view.** Sparkline of the last 5 hours, opened by clicking
  the status bar.

All would be additive to the snapshot protocol.
