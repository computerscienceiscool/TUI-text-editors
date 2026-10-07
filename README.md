# TUI Text Editors

A collection of small Charm-based editors for a team learning session. Each
editor keeps the same core document actions while exploring a different visual
and interaction design.

## Editors

### 01 — Plain Default

The control version: a deliberately neutral Markdown-first editor with a normal
menu bar, document area, Glamour preview, status feedback, and keyboard help.
It exists so later editors have a clear baseline to challenge.

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

Press `Alt+M` to open the menu. Arrow keys navigate menus; `Enter` selects an
action. `Ctrl+P` toggles preview and `Ctrl+C` quits. Open and save `.md`,
`.markdown`, and `.txt` files.
