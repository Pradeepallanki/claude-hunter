# Troubleshooting

Symptom-first. Each entry points at what to check and, where
possible, how to confirm the fix. If you hit something not listed
here, open the diagnostic log first (VS Code: **Output → Claude
Hunter**; IntelliJ: `idea.log`) and file an issue with the last ~20
lines attached.

---

## Widget shows only `Claude ⏳` and never updates

The extension activated but no snapshot line ever arrived from the
core binary. Diagnose in this order:

1. **Confirm the extension activated.**
   - VS Code: **Output → Claude Hunter** should show
     `[activate] hunter process started`.
   - IntelliJ: `idea.log` should show entries from
     `co.reward360.claudehunter.*`.
2. **Confirm the binary was found.** The `[activate] binary=…` line
   in VS Code Output tells you which path was chosen. If it reads
   `NOT FOUND`, set `claudeHunter.binaryPath` explicitly (see
   [`CONFIGURATION.md`](CONFIGURATION.md)) or rebundle
   (see [`BUILDING.md`](BUILDING.md)).
3. **Confirm the binary runs on its own.**
   ```bash
   /path/to/claude-hunter --projects-dir ~/.claude/projects --emit-interval-ms 500 | head -1
   ```
   If this prints a JSON snapshot, the binary is healthy; the
   problem is in the extension's plumbing.
   If it hangs silently, jump to
   [Stuck on Claude ⏳ (macOS)](#stuck-on-claude--macos).

---

## Stuck on `Claude ⏳` (macOS)

**Symptom.** VS Code Output shows `[activate] hunter process started`
but never `[snapshot #1]`. The process appears in `ps` with state
`UE` and RSS around 32 KB.

**Cause.** macOS Ventura+ silently blocks execution of binaries that
live under protected locations (`~/Desktop/`,
`~/.vscode/extensions/`, `~/Documents/`) when spawned by a notarised
parent process such as VS Code. The spawn appears to succeed but the
process wedges before its first syscall.

**Fix.** The VS Code extension already stages the bundled binary to
`$TMPDIR/claude-hunter/<version>/` on activation. If you are seeing
this on `0.1.0`, confirm you are running compiled TypeScript that
includes `stageBinaryForExecution` in
`out/binary_locator.js`:

```bash
grep -c stageBinaryForExecution vscode/out/binary_locator.js
```

Should print `1` or more. If it prints `0`, run `npm run compile` in
`vscode/`.

**Verification.** After a window reload, the diagnostic log's
`[activate] binary=…` line should point at a path under
`/var/folders/…/T/claude-hunter/…`, not the extension directory.

**IntelliJ.** The plugin already extracts to `$TMPDIR` on every load,
so this failure mode is not applicable there.

---

## Widget shows the wrong model after IDE reload

**Symptom.** You were on Opus a moment ago; you reload the IDE and
the widget now says `haiku-4-5`.

**Cause.** Prior to `0.1.0` the widget displayed the model of the
*last record processed*, not the model of the record with the
greatest timestamp. On startup the core binary re-reads history from
disk in filesystem order; whichever session file happened to be
scanned last dictated the model — often a Haiku record used
internally by Claude Code for auto-titling.

**Fix.** Upgrade to `0.1.0` or newer. The core now exposes
`LatestModel()`, which picks the record with the greatest timestamp.

---

## Burn rate shows `0 tok/min`

Three legitimate reasons and one gotcha.

- **No assistant turn has completed in the last 10 min.** Claude Code
  writes usage records when a turn finishes; if you have been idle
  or the response is still streaming, burn is genuinely zero.
- **The extension just activated.** Give the seed scan a moment to
  populate the window. First useful burn arrives ~1–2 seconds after
  activation.
- **You are looking at a cached tooltip.** VS Code freezes the
  tooltip content when you first hover; move away, wait a beat, and
  hover again to see fresh numbers.
- **Gotcha:** if `--window-hours` is set to something absurdly small,
  the 10-min sample can fall outside it. Reset to `5`.

Verify the core is producing non-zero values with:

```bash
/path/to/claude-hunter --projects-dir ~/.claude/projects | head -1 | \
  python3 -c 'import sys, json; s = json.loads(sys.stdin.read()); print(s["window5h"]["burnRatePerMinute"])'
```

---

## Percentage looks way off

**Symptom.** You feel you've used far more (or less) than the widget
shows.

**Cause.** The ceiling is an *estimate*. Anthropic doesn't publish an
exact "effective tokens per 5 hours" number for Claude Code, so the
plugin ships with `88M` — the observed limit on a Max-20x plan — and
the percentage is (effective tokens) / (ceiling) × 100.

**Fix.** Recalibrate against your plan. See
[Plan ceilings in CONFIGURATION.md](CONFIGURATION.md#plan-ceilings).

---

## `claude-hunter exited (code unknown)` after reload

**Symptom.** Immediately after reloading VS Code, the widget shows
this error briefly.

**Cause.** Prior to `0.1.0` the extension's `onExit` callback ran
even when the process was terminated intentionally by our own
`stop()`. The old process's SIGTERM would fire `onExit`, which
overwrote the freshly-activated widget with the error text.

**Fix.** Upgrade to `0.1.0`. `HunterProcess` now sets a
`stopRequested` flag and skips the callback for intentional stops.

---

## Nothing in `~/.claude/projects`

Claude Hunter has nothing to display until Claude Code has produced at
least one session file. Send a single message through Claude Code and
the widget will populate on the next snapshot tick.

---

## Filing a bug

Please include:

- Claude Hunter version (in the plugin/extension manifest).
- IDE name and version.
- Output of `uname -a` and, on macOS, `sw_vers`.
- The last ~30 lines of the diagnostic log.
- Whether the core binary works stand-alone
  (`./claude-hunter --projects-dir ~/.claude/projects | head -1`).
