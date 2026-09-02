# Configuring Claude Hunter

Every runtime knob is exposed both as a **core binary flag** (for
running the binary directly, for example when packaging) and as an
**IDE setting** (surfaced through VS Code and IntelliJ). Values set in
the IDE take precedence — they are forwarded to the binary as flags.

The defaults are tuned so that Claude Hunter is useful with zero
configuration; you only need to change something if the widget shows
the wrong percentage or points at an unusual projects location.

---

## Core binary flags

Run `./claude-hunter --help` for the current list. As of `0.1.0`:

| Flag | Default | Purpose |
|---|---|---|
| `--projects-dir <path>` | `~/.claude/projects` | Root directory Claude Code writes session JSONL to. Change only if Claude Code is configured with a non-standard data directory. |
| `--window-hours <float>` | `5.0` | Rolling window duration. Anthropic enforces 5 hours; changing this only affects the widget's arithmetic, not the actual limit. |
| `--emit-interval-ms <int>` | `250` | How often a snapshot is written to stdout. Lower values feel snappier; higher values reduce IDE main-thread wake-ups. |
| `--ceiling-millions <float>` | `88.0` | Estimated effective-token ceiling for your plan, in millions. This drives the percentage and the traffic-light colours. See [Plan ceilings](#plan-ceilings) below. |

Effective tokens use Anthropic's rate-limit weighting:

```
effective = input + output + cache_creation + (cache_read × 0.1)
```

Cost is computed from a per-model pricing table baked into the binary
(see `core/pricing/table.go`). Update the table and rebuild if
Anthropic changes rates.

---

## VS Code settings

Open **Settings** and search for **Claude Hunter**, or edit
`settings.json` directly.

| Setting | Type | Default | Purpose |
|---|---|---|---|
| `claudeHunter.binaryPath` | string | `""` | Absolute path to a `claude-hunter` binary. Empty means "use the bundled binary". |
| `claudeHunter.projectsDir` | string | `""` | Overrides `--projects-dir`. Empty means "use the binary's default". |
| `claudeHunter.windowHours` | number | `5` | Overrides `--window-hours`. |
| `claudeHunter.ceilingMillions` | number | `88` | Overrides `--ceiling-millions`. |

Changes are picked up on the next window reload.

---

## IntelliJ settings

The IntelliJ plugin currently accepts overrides only through JVM
system properties passed to the IDE. Add them to the IDE's
`Custom VM Options…` file (Help → Edit Custom VM Options…):

```
-Dclaude-hunter.binary-path=/absolute/path/to/claude-hunter
```

More settings will be exposed in a future release. See
[`ARCHITECTURE.md`](ARCHITECTURE.md#future-work) for the roadmap.

---

## Plan ceilings

Anthropic does not publish a single "effective tokens per 5 hours"
number for Claude Code, so the widget's percentage is an estimate.
`--ceiling-millions` lets you tune it.

Rough starting points observed in practice — measure and adjust for
your own plan:

| Plan | Suggested ceiling (millions of effective tokens / 5h) |
|---|---|
| Max 5× | ~22 |
| Max 20× | ~88 |
| Team | varies — start at your Max-equivalent |

To recalibrate: watch the widget until you hit the actual limit; note
the effective-token total at that moment; set
`--ceiling-millions` (or `claudeHunter.ceilingMillions`) to that
number in millions. That gives you an accurate 100% mark for the rest
of the window.

---

## Debugging output

Set `CLAUDE_HUNTER_DEBUG=1` in the environment of the process before
running the core binary directly for verbose stderr logs. The IDE
extensions do not currently propagate this variable; use the
diagnostic **Output → Claude Hunter** channel (VS Code) or `idea.log`
(IntelliJ) instead.
