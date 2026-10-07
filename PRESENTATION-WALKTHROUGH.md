# TUI Text Editors — Presentation Walkthrough

This is a live-demo script, not product documentation. The point is to help a
design team see that a terminal interface can be calm, guided, expressive, or
dense while preserving the same basic document capabilities.

## Before the presentation

From the repository root, verify the demos once:

```sh
make check
```

Launch each demo in its own terminal when you reach it:

```sh
make run
make guided-brief
make color-field
make ops-deck
```

The demos all accept `Ctrl+C` to quit. Use a sufficiently wide terminal for
the split-pane variants; 140 columns is comfortable for Ops Deck.

## What is in this repository

All four editors use the Charm ecosystem:

| Component | Where it appears | What it demonstrates |
| --- | --- | --- |
| Bubble Tea | Every editor | The event loop and state model behind an interactive TUI. |
| Bubbles `textarea` | Every editor | Real text input, cursor movement, and editing behavior. |
| Bubbles `textinput` | Every editor | Open/save path prompts. |
| Bubbles `list` | Color Field | A filterable command palette. |
| Huh | Guided Brief | A structured, keyboard-first form that creates a document brief. |
| Lip Gloss | Every editor | Layout, borders, color, labels, panels, and visual hierarchy. |
| Glamour | Every editor | Rendered Markdown reading views. |

**TUIos is not used in these programs.** It is a separate framework direction
for dashboard/application shells. These examples use Charm components directly
so the team can see the building blocks that can live inside many kinds of TUI
applications, including a future TUIos shell.

## Suggested opening — 30 seconds

Say:

> “These are not four products. They are four answers to the same interaction
> question: how can a person write and read a Markdown document in a terminal?
> The capabilities stay familiar; the interface changes what it emphasizes.”

Then name the progression:

1. Familiarity — Plain Default.
2. Purpose — Guided Brief.
3. Expression and action vocabulary — Color Field.
4. Scanning and operational context — Ops Deck.

## 01 — Plain Default

Launch:

```sh
make run
```

### Demonstrate

1. Point out the conventional menu bar and split document/preview layout.
2. Type this in the left pane:

   ```md
   ## Release update

   - [ ] Confirm owner
   - [x] Draft is ready

   **Bold**, *italic*, and <u>underlined</u>.
   ```

3. Point to the Markdown rendering in the right pane.
4. Press `F10`; use arrow keys and `Enter` to choose an action.
5. Mention `Ctrl+S` / `Ctrl+O`, `Ctrl+P`, and the visible literal caret.

### Say

> “This is the control version. It proves a TUI can be conventional and
> legible. It is our baseline—not the only possible terminal experience.”

### Design lesson

**Restraint and familiarity.** A terminal can offer menus, panels, status,
help, and a reading view without becoming an IDE or a web app imitation.

## 02 — Guided Brief

Launch:

```sh
make guided-brief
```

### Demonstrate

1. In **What are you writing?**, choose `Proposal` or `Incident update`.
2. Choose an audience.
3. In the two multi-select steps, use `Space` to select items and `Enter` to
   continue. Pick at least two tones and two sections.
4. Show the generated Markdown draft and the reading view.
5. Press `Ctrl+G` to return to the brief form; point out that this is an
   intentional alternative to presenting an empty canvas.

### Say

> “Before asking someone to write, we can ask what they are trying to achieve.
> The interface captures purpose, audience, tone, and structure—then still
> hands the person a normal editable document.”

### Design lesson

**Guidance and purpose.** Some workflows benefit from structured choices
before freeform editing. A form can be an authoring tool, not only data entry.

## 03 — Command Palette / Color Field

Launch:

```sh
make color-field
```

### Demonstrate

1. Point out the colored categories: `1 FILE`, `2 FORMAT`, `3 INSERT`, and
   `4 VIEW`.
2. Press `Ctrl+K` to open the palette.
3. Press `2`; show that only Format commands remain.
4. Press `0` for all commands.
5. Press `/`, type `view`, then use arrows and `Enter` to choose a command.
6. Return to the editor and point out that category color appears both in the
   header and each command row.

### Say

> “Here color is not decoration. It is an action vocabulary: a category has a
> color, a keyboard filter, and matching commands. If color did not map to
> action, it would be visual noise.”

### Design lesson

**Color as interaction language.** Designers can use color to create a stable
mental map across commands, shortcuts, and state—not merely to make a terminal
look lively.

## 04 — Ops Deck

Launch:

```sh
make ops-deck
```

### Demonstrate

1. Type headings with `Ctrl+H`; watch the left outline populate.
2. Type a list with `Ctrl+L` and tasks with `Ctrl+T`.
3. Point to the right reading pane and the footer metrics: words, tasks,
   headings, cursor location, and shortcuts.
4. Press `Ctrl+F` to enter focus mode; press it again to restore the deck.
5. Mention `Ctrl+N`, `Ctrl+O`, and `Ctrl+S` for document actions.

### Say

> “This version intentionally uses more of the screen. It is for an operator
> who needs to scan structure, edit, verify Markdown, and see document health
> without changing screens.”

### Design lesson

**Operational density.** Whitespace is not always the goal. For high-context
work, persistent context can be kinder than repeated navigation.

## Closing question for the team

Ask:

> “For the workflow we are designing, should the terminal feel familiar,
> guided, expressive, or operational? What information must always be visible,
> and what can stay behind a command?”

The useful output is not a vote for one demo. It is a shared vocabulary for
choosing an interface shape deliberately.
