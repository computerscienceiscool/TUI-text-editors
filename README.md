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

Run:

```sh
make run
```

### 02 — Guided Brief

An editor that begins with purpose rather than an empty document. A Charm Huh
form asks what is being written, for whom, in what tone, and which sections it
needs; it then creates an editable Markdown draft with the same save, open,
formatting, and Glamour-preview capabilities as Plain Default. It demonstrates
that terminal interaction can guide structure and intent, not merely expose a
blank canvas plus tools.

Run:

```sh
make guided-brief
```

### 03 — Command Palette / Color Field

A deliberately vivid editor that replaces menus with one command palette. Its
colored action chips group File, Format, Insert, and View; `Ctrl+K` opens a
searchable command list. It demonstrates color as an interaction language:
color can establish categories, orientation, and an action vocabulary rather
than simply decorate a terminal.

Run:

```sh
make color-field
```

### 04 — Ops Deck

A dense operational layout with a persistent outline, live draft, reading
view, metrics, and a compact key map. It demonstrates that a TUI can favor
rapid scanning and action over whitespace—useful for operational or
high-context work where several views must remain visible at once.

Run:

```sh
make ops-deck
```

### 05 — Signal / Accessible

A high-clarity editor that uses explicit words, symbols, borders, numbered
lines, a literal caret, and redundant status labels rather than relying on
color alone. It demonstrates that accessible TUI design can make state,
focus, and keyboard actions clear for everyone—not just users who perceive a
specific color, animation, or visual subtlety.

Run:

```sh
make signal-accessible
```

### 06 — Playground

An expressive editor with writing prompts, visual mood labels, playful language, and
small “sticker” actions for titles, steps, quotes, and sparkles. It shows that
terminal interfaces can invite experimentation and spark design ideas without
sacrificing a real Markdown editor and preview.

Run:

```sh
make playground
```

### 07 — Windows Desktop

A 1990s desktop-inspired treatment of the same editor: a teal desktop, gray
window surfaces, navy title bars, and raised controls. It demonstrates that
foreground and background colors can create spatial layers, ownership, and
action surfaces—not just accents on a single flat canvas.

Run:

```sh
make windows-desktop
```

### 08 — Blue Paper

A bright white-paper editor with Windows-blue chrome, blue ink, and a
mouse-enabled color-category menu. It demonstrates that a terminal can occupy
the whole screen as a deliberately light application surface—not merely a
dark console with styled widgets.

Run:

```sh
make blue-paper
```

## Project layout

- `cmd/<editor-name>/` contains one runnable editor variant.
- `internal/editor/` contains shared document behavior: supported file types,
  file loading/saving, and formatting insertion.

Future variants receive their own `cmd/` directory and a new entry in the
catalog above once complete. They keep the same document behavior while being
free to change menus, layout, interaction model, and visual style.

Plain Default uses `F10` or `Alt` menu mnemonics; the other editors document
their shortcuts in their own footer. All variants open and save `.md`,
`.markdown`, and `.txt` files.

## Verification

Run `make check` before presenting a variant. It runs unit tests, static
analysis, a build, and a real pseudo-terminal smoke test. Unit tests cover
menu access and the `Ctrl+C` quit command. The smoke test types into the editor,
checks that Glamour does not leak terminal color probes into the document, and
requires the terminal session to exit cleanly.
