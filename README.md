# TUI Text Editors

A collection of small Charm-based editors for a team learning session. Each
editor keeps the same core document actions while exploring a different visual
and interaction design.

For a presenter-facing live-demo script, see
[PRESENTATION-WALKTHROUGH.md](PRESENTATION-WALKTHROUGH.md).

## Editors

### 01 — Plain Default

The control version: a deliberately neutral Markdown-first editor with a normal
menu bar, document area, Glamour preview, status feedback, and keyboard help.
It exists so later editors have a clear baseline to challenge.

### 02 — Guided Brief

An editor that begins with purpose rather than an empty document. A Charm Huh
form asks what is being written, for whom, in what tone, and which sections it
needs; it then creates an editable Markdown draft with the same save, open,
formatting, and Glamour-preview capabilities as Plain Default. It demonstrates
that terminal interaction can guide structure and intent, not merely expose a
blank canvas plus tools.

### 03 — Command Palette / Color Field

A deliberately vivid editor that replaces menus with one command palette. Its
colored action chips group File, Format, Insert, and View; `Ctrl+K` opens a
searchable command list. It demonstrates color as an interaction language:
color can establish categories, orientation, and an action vocabulary rather
than simply decorate a terminal.

### 04 — Ops Deck

A dense operational layout with a persistent outline, live draft, reading
view, metrics, and a compact key map. It demonstrates that a TUI can favor
rapid scanning and action over whitespace—useful for operational or
high-context work where several views must remain visible at once.

## Project layout

- `cmd/<editor-name>/` contains one runnable editor variant.
- `internal/editor/` contains shared document behavior: supported file types,
  file loading/saving, and formatting insertion.

Future variants receive their own `cmd/` directory and a new entry in the
catalog above once complete. They keep the same document behavior while being
free to change menus, layout, interaction model, and visual style.

## Run

```sh
go run ./cmd/plain-default
```

Run Guided Brief with:

```sh
go run ./cmd/guided-brief
```

Run Command Palette / Color Field with:

```sh
go run ./cmd/color-field
```

Run Ops Deck with:

```sh
go run ./cmd/ops-deck
```

Press `F10` to open the menu (or use `Alt+F`, `Alt+E`, `Alt+O`, `Alt+I`,
`Alt+V`, or `Alt+H` for a specific menu). Arrow keys navigate menus; `Enter`
selects an action; `Esc` closes a menu or prompt. `Ctrl+P` toggles preview and
`Ctrl+C` quits. Open and save `.md`,
`.markdown`, and `.txt` files.

## Verification

Run `make check` before presenting a variant. It runs unit tests, static
analysis, a build, and a real pseudo-terminal smoke test. Unit tests cover
menu access and the `Ctrl+C` quit command. The smoke test types into the editor,
checks that Glamour does not leak terminal color probes into the document, and
requires the terminal session to exit cleanly.
