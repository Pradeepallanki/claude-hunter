# Claude Hunter

**Live Claude Code token-usage monitor for VS Code and IntelliJ.**

Claude Hunter puts a small widget in your IDE status bar that continuously
shows how much of your Claude Code 5-hour rolling limit you have already
spent, what it has cost, which model is being used, and how fast you are
burning tokens right now — so you can wrap up a task cleanly instead of
being interrupted mid-flow by a *usage limit exceeded* error.
```
opus-4-7 🔋 73% · 25.8M/5h · $67.20
```

Hover the widget for a per-turn breakdown of input / output / cache
tokens and a per-model cost attribution.

- **No cloud, no telemetry, no API keys.** Everything is read from the
  JSONL session files Claude Code already writes to
  `~/.claude/projects/`.
- **Tiny.** A ~3.5 MB Go binary watches the filesystem; the IDE
  extensions are thin wrappers.
- **Cross-IDE.** VS Code and IntelliJ share one core so both status
  bars show identical numbers.

---

## Why this exists

Claude Code enforces a rolling 5-hour token limit and, on some plans, a
weekly ceiling. When you hit it mid-task, the current conversation
context is effectively lost — you cannot continue until the window
rolls over. A live counter turns that from a hard cliff into a visible
gauge you can plan around.

See [`docs/PLAN.md`](docs/PLAN.md) for the original design brief and
open questions.

---

## Quick start

1. **Install for your IDE** —
   see [`docs/INSTALLING.md`](docs/INSTALLING.md) for VS Code and
   IntelliJ walk-throughs.
2. **(Optional) Tune the plan ceiling** to match your Claude Code
   subscription — see
   [`docs/CONFIGURATION.md`](docs/CONFIGURATION.md).
3. **Verify** by looking at the status bar; within about a second the
   `⏳` placeholder is replaced by the live readout above.

If nothing appears, jump to
[`docs/TROUBLESHOOTING.md`](docs/TROUBLESHOOTING.md).

---

## Documentation

| Document | What it covers |
|---|---|
| [`docs/INSTALLING.md`](docs/INSTALLING.md) | End-user install for VS Code and IntelliJ. |
| [`docs/CONFIGURATION.md`](docs/CONFIGURATION.md) | Every flag and IDE setting, with defaults and typical values. |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | How the core, snapshot stream, and IDE clients fit together. |
| [`docs/BUILDING.md`](docs/BUILDING.md) | Build from source and cross-compile for release. |
| [`docs/TROUBLESHOOTING.md`](docs/TROUBLESHOOTING.md) | Empty status bar, macOS Desktop restriction, wrong model, and other known issues. |
| [`docs/CONTRIBUTING.md`](docs/CONTRIBUTING.md) | Development workflow, TDD requirement, PR expectations. |
| [`docs/coding_standards.md`](docs/coding_standards.md) | Non-negotiable code style — SRP, TDD, naming, layout. |
| [`docs/PLAN.md`](docs/PLAN.md) | Original design document. |
| [`CHANGELOG.md`](CHANGELOG.md) | Release notes. |

---

## Project layout

```
claude-hunter/
├── core/         Go binary — fsnotify watcher, JSONL parser, aggregator,
│                 NDJSON snapshot emitter.
├── vscode/       VS Code extension (TypeScript). Bundles the core binary
│                 per platform under bin/<platform>-<arch>/.
├── intellij/     IntelliJ plugin (Kotlin + Gradle). Bundles the core
│                 binary under src/main/resources/bin/<platform>-<arch>/.
├── docs/         All contributor and user documentation.
├── LICENSE       MIT License.
└── CHANGELOG.md  Release notes.
```

---

## Compatibility

| Component | Minimum version |
|---|---|
| Claude Code | any version that writes session JSONL to `~/.claude/projects/` |
| VS Code | 1.80 |
| IntelliJ Platform | 2024.1 (build 241) — Community or Ultimate |
| Go (build only) | 1.22 |
| Node.js (build only) | 18 |
| JDK (build only) | 17 |

Prebuilt binaries are shipped for `darwin-arm64`, `darwin-x64`,
`linux-x64`, `linux-arm64`, and `win-x64`. See
[`docs/BUILDING.md`](docs/BUILDING.md) if your platform is not covered.

---

## License

Claude Hunter is licensed under the MIT License. See
[`LICENSE`](LICENSE) for the full text.

Maintained by Reward360.
