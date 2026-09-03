# Installing Claude Hunter

This document walks an end user through installing Claude Hunter in
either **VS Code** or **IntelliJ**. To build from source instead, see
[`BUILDING.md`](BUILDING.md).

---

## Prerequisites

- **Claude Code** is installed and has produced at least one session
  file under `~/.claude/projects/…/*.jsonl`. Nothing to configure —
  Claude Code writes these files by default.
- The IDE you are installing into is **running on the same machine**
  that produces those session files. Claude Hunter is a local-only
  tool; it does not read anything over the network.

Nothing else is required for use. No API keys, no cloud accounts, no
Anthropic tokens.

---

## VS Code

### Option A — install from a prebuilt `.vsix`

1. Download `claude-hunter-<version>.vsix` from your internal release
   channel (or produce one with `vsce package` — see
   [`BUILDING.md`](BUILDING.md)).
2. In VS Code, open the Command Palette (`Cmd+Shift+P` /
   `Ctrl+Shift+P`) and run **Extensions: Install from VSIX…**. Pick
   the file.
3. Reload the window when prompted.
4. The widget appears on the right side of the status bar within a
   second.

### Option B — install a linked development copy

Useful when iterating on the extension source.

```bash
cd claude-hunter/vscode
npm install
npm run compile
```

Then either open the `vscode/` folder in VS Code and press `F5` to
launch an **Extension Development Host**, or symlink the folder into
`~/.vscode/extensions/`:

```bash
ln -s "$(pwd)" ~/.vscode/extensions/pradeep.claude-hunter-0.1.0
```

Reload VS Code. This method is what you want if you are actively
editing the extension.

> **macOS note.** VS Code stages the bundled binary out of the
> extension directory into `$TMPDIR/claude-hunter/<version>/` on
> activation, because macOS Ventura+ blocks execution of binaries
> that live under `~/Desktop/`, `~/.vscode/extensions/`, and a few
> other protected locations. This is automatic — no user action
> required. See
> [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md#stuck-on-claude--macos)
> if the widget ever gets stuck.

### Verifying VS Code

Open **View → Output** and pick **Claude Hunter** in the dropdown.
Within a second you should see:

```
[activate] extensionPath=…
[activate] binary=…/claude-hunter
[activate] hunter process started
[snapshot #1] model=claude-opus-4-7 effective=…
```

If the `[snapshot #1]` line never appears, see
[`TROUBLESHOOTING.md`](TROUBLESHOOTING.md).

---

## IntelliJ

Claude Hunter works with any JetBrains IDE built on IntelliJ Platform
2024.1 or newer — IntelliJ IDEA, PyCharm, GoLand, WebStorm, Rider, etc.
The instructions below use IntelliJ IDEA; they apply verbatim to the
other IDEs.

### Install

1. Obtain `claude-hunter-intellij-<version>.zip`. Either download it
   from your internal release channel or build it locally
   (`./gradlew buildPlugin` — see [`BUILDING.md`](BUILDING.md)); the
   output lands under `intellij/build/distributions/`.
2. In IntelliJ, open **Settings → Plugins**, click the ⚙ gear icon,
   then **Install Plugin from Disk…**. Pick the ZIP.
3. Restart the IDE when prompted.

### Enabling the widget

If the widget does not appear after restart:

- Right-click any empty area of the status bar → confirm
  **Claude Hunter** is checked.
- Or: **Settings → Appearance & Behavior → Menus and Toolbars →
  Status Bar → Claude Hunter → Enable**.

### Verifying IntelliJ

Open **Help → Show Log in Finder / Explorer**, then tail
`idea.log`. On startup you should see lines from
`com.pradeep.claudehunter.*`. Warnings from `SnapshotParser` mean the
core binary emitted an unexpected line — usually harmless.

---

## Uninstalling

- **VS Code**: Extensions view → find **Claude Hunter** → **Uninstall**.
- **IntelliJ**: Settings → Plugins → find **Claude Hunter** → **Uninstall**.

Both installers leave a small cache under
`$TMPDIR/claude-hunter/` (macOS/Linux) or
`%TEMP%\claude-hunter\` (Windows). It is safe to delete at any time
and will be recreated on next activation.

---

## Next

- Adjust the plan ceiling and other options —
  [`CONFIGURATION.md`](CONFIGURATION.md).
- Something wrong? — [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md).
