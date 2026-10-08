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
make signal-accessible
make playground
make windows-desktop
```

The demos all accept `Ctrl+C` to quit. Use a sufficiently wide terminal for
the split-pane variants; 140 columns is comfortable for Ops Deck.

## What is in this repository

All seven editors use the Charm ecosystem:

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

> “These are not seven products. They are seven answers to the same interaction
> question: how can a person write and read a Markdown document in a terminal?
> The capabilities stay familiar; the interface changes what it emphasizes.”

Then name the progression:

1. Familiarity — Plain Default.
2. Purpose — Guided Brief.
3. Expression and action vocabulary — Color Field.
4. Scanning and operational context — Ops Deck.
5. Clarity and redundant cues — Signal / Accessible.
6. Play and emotional tone — Playground.
7. Spatial hierarchy through foreground and background — Windows Desktop.

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

## 05 — Signal / Accessible

Launch:

```sh
make signal-accessible
```

### Demonstrate

1. Point out the literal labels: `[EDITING]`, `[READING]`, and `[STATUS]`.
2. Show the permanent F-key action map at the bottom.
3. Use `F5` for a heading, `F6` for a bullet, and `F7` to toggle preview.
4. Mention the numbered gutter and literal caret as cues that do not depend on
   subtle color changes or a blinking terminal cursor.

### Say

> “Accessibility is not a monochrome fallback. It is a design discipline of
> making state, focus, and actions legible through more than one channel.”

### Design lesson

**Redundant cues and clarity.** Words, symbols, borders, position, and
keyboard labels can reinforce one another so meaning does not require a
specific visual ability or terminal theme.

## 06 — Playground

Launch:

```sh
make playground
```

### Demonstrate

1. Point out the current mood and writing prompt in the left panel.
2. Press `Tab` to cycle the mood label; it is deliberately a visual/tonal
   treatment, not a different editor mode.
3. Press `F1` to cycle a writing prompt.
4. Use `F5`, `F6`, `F7`, and `F8` to insert a title, step, quote, and sparkle
   at the caret.
5. Show that the draft remains a normal editable Markdown document with a
   separate reading view and ordinary save/open actions.

### Say

> “Not every demo needs to solve a production workflow. This one is a sketch
> meant to spark an idea: terminals can have tone, warmth, and a little joy,
> while the underlying editor stays fully real.”

### Design lesson

**Play and emotional tone.** Language, visual mood, micro-feedback, and
optional ritual can spark new design ideas without pretending to be a complete
product workflow.

## 07 — Windows Desktop

Launch:

```sh
make windows-desktop
```

### Demonstrate

1. Point out the teal desktop, gray window surface, navy title bar, and raised
   toolbar buttons before typing anything.
2. Use `F2` for a new document, `F3` to save, `F4` to open, `F5` to insert a
   heading, and `F6` to insert a list item.
3. Press `F7` to hide and restore the reading view.
4. Emphasize that the document editing and Markdown behavior are the same as
   the other demos; only the visual language has changed.

### Say

> “A terminal does not have to be one undifferentiated black rectangle.
> Foreground and background colors can construct a desktop, a window, a title
> bar, controls, and content surfaces—even with the same small set of actions.”

### Design lesson

**Foreground/background as spatial hierarchy.** Color can define nested
surfaces and ownership. A familiar visual reference, such as an early desktop
window, can make those layers immediately legible without changing the task.

## Closing question for the team

Ask:

> “For the workflow we are designing, should the terminal feel familiar,
> guided, expressive, or operational? What information must always be visible,
> and what can stay behind a command?”

The useful output is not a vote for one demo. It is a shared vocabulary for
choosing an interface shape deliberately.
