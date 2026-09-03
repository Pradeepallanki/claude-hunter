# Building Claude Hunter from Source

This document is for developers who want to build the core binary and
the IDE extensions themselves. End users should follow
[`INSTALLING.md`](INSTALLING.md) instead.

---

## Prerequisites

| Tool | Version | Used for |
|---|---|---|
| Go | 1.22 or newer | The core binary. |
| Node.js | 18 or newer | The VS Code extension. |
| npm | comes with Node | VS Code dependency install. |
| JDK | 17 | The IntelliJ plugin (Gradle toolchain). |
| Git | any recent | Clone. |

On macOS, `brew install go node openjdk@17` covers everything. The
IntelliJ plugin's Gradle wrapper downloads the JetBrains SDK on first
run (large; expect a few minutes).

---

## Core binary

```bash
cd core
go test ./...
go build -o bin/claude-hunter .
```

The result is a single ~3.5 MB static binary. You can inspect its
NDJSON output directly:

```bash
./bin/claude-hunter --projects-dir ~/.claude/projects --emit-interval-ms 500
```

### Running the tests

TDD is a non-negotiable standard here — see
[`coding_standards.md`](coding_standards.md). Every core change is
tested first. To run:

```bash
cd core
go test ./...
```

Package-level tests exist for `usage`, `pricing`, `window`, `watcher`,
and `snapshot`. The `watcher` package includes an integration test
that exercises the fsnotify observer end-to-end in a temporary
directory.

---

## VS Code extension

```bash
cd vscode
npm install
npm run compile
```

Output lands in `vscode/out/`. To exercise the extension:

- **Development host** — open `vscode/` in VS Code and press `F5`. A
  new window launches with the extension activated.
- **Local install** — symlink `vscode/` into
  `~/.vscode/extensions/pradeep.claude-hunter-<version>` and reload
  the parent VS Code window.
- **Package a VSIX** — install `vsce` (`npm i -g @vscode/vsce`), then
  `vsce package` inside `vscode/`. This produces
  `pradeep-claude-hunter-<version>.vsix`.

The extension expects the core binary to be present under
`vscode/bin/<platform>-<arch>/claude-hunter`. Copy it there after
building the core (see [Bundling binaries](#bundling-binaries-for-release)).

---

## IntelliJ plugin

```bash
cd intellij
./gradlew buildPlugin
```

Output lands under `intellij/build/distributions/`. Install the ZIP
via **Settings → Plugins → ⚙ → Install Plugin from Disk…**.

The plugin expects the core binary under
`intellij/src/main/resources/bin/<platform>-<arch>/claude-hunter`.
Copy it there before packaging.

### Notes on the toolchain

- We use the **IntelliJ Platform Gradle Plugin v2** (`org.jetbrains.intellij.platform`
  `2.1.0`) rather than the older `org.jetbrains.intellij`. v2 supports
  Gradle 8.x/9.x cleanly.
- Gradle wrapper is pinned to 8.10 for reproducibility.
- The plugin declares `intellijIdeaCommunity("2024.1")` and
  `instrumentationTools()`. First `buildPlugin` downloads the IDE
  distribution — this is normal.

---

## Bundling binaries for release

Cross-compile the core once per target and drop the result into each
extension's `bin/…` folder before packaging.

```bash
cd core

# darwin
GOOS=darwin  GOARCH=arm64 go build -o ../vscode/bin/darwin-arm64/claude-hunter .
GOOS=darwin  GOARCH=amd64 go build -o ../vscode/bin/darwin-x64/claude-hunter .

# linux
GOOS=linux   GOARCH=amd64 go build -o ../vscode/bin/linux-x64/claude-hunter .
GOOS=linux   GOARCH=arm64 go build -o ../vscode/bin/linux-arm64/claude-hunter .

# windows
GOOS=windows GOARCH=amd64 go build -o ../vscode/bin/win-x64/claude-hunter.exe .
```

Repeat the same output paths against
`../intellij/src/main/resources/bin/…`. A small helper script is a
reasonable follow-up if you find yourself doing this often.

---

## End-to-end smoke test

Handy when validating a fresh build against a scratch dataset:

```bash
TMP_PROJECTS=$(mktemp -d)
mkdir -p "$TMP_PROJECTS/proj-a"
./core/bin/claude-hunter --projects-dir "$TMP_PROJECTS" --emit-interval-ms 300 &
HUNTER_PID=$!
sleep 0.4
printf '%s\n' '{"type":"assistant","sessionId":"s","timestamp":"'"$(date -u +%FT%TZ)"'","message":{"model":"claude-opus-4-7","usage":{"input_tokens":1000,"output_tokens":500,"cache_creation_input_tokens":4000,"cache_read_input_tokens":10000}}}' \
  >> "$TMP_PROJECTS/proj-a/session.jsonl"
sleep 1
kill $HUNTER_PID
```

You should see one or more `{"type":"snapshot",…}` lines on stdout,
with `effectiveTokens: 6500` after the append. If not, see
[`TROUBLESHOOTING.md`](TROUBLESHOOTING.md).
