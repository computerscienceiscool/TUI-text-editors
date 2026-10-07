package main

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"tui-text-editors/internal/editor"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func TestInitialViewWaitsForTerminalSize(t *testing.T) {
	m := newModel()
	if got := m.View(); got != "" {
		t.Fatalf("initial view should wait for terminal size, got %q", got)
	}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 110, Height: 36})
	m = updated.(model)
	if m.width != 110 || m.height != 36 {
		t.Fatalf("size = %dx%d, want 110x36", m.width, m.height)
	}
	if got := m.View(); !strings.Contains(got, "File") {
		t.Fatalf("sized view should render the editor, got %q", got)
	}
}

func TestTypingReachesFocusedEditor(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	got := updated.(model).editor.Value()
	if got != "hello" {
		t.Fatalf("editor value = %q, want %q", got, "hello")
	}
}

func TestEditorUsesVisibleStaticCursor(t *testing.T) {
	m := newModel()
	if !m.editor.Focused() {
		t.Fatal("editor should start focused")
	}
	if m.editor.Cursor.Mode() != cursor.CursorStatic {
		t.Fatalf("cursor mode = %v, want static", m.editor.Cursor.Mode())
	}
	if m.editor.Cursor.Blink {
		t.Fatal("static cursor should be continuously visible")
	}
}

func TestEditorGutterShowsEveryVisibleLineAndCurrentLine(t *testing.T) {
	m := newModel()
	m.editor.SetWidth(40)
	m.editor.SetHeight(3)
	got := ansiEscape.ReplaceAllString(m.editor.View(), "")
	for _, want := range []string{"▶   1", "    2", "    3"} {
		if !strings.Contains(got, want) {
			t.Fatalf("gutter missing %q in %q", want, got)
		}
	}
}

func TestDocumentViewShowsLiteralCaretAtEditingPosition(t *testing.T) {
	m := newModel()
	m.width, m.height = 80, 16
	m.resize()
	m.editor.SetValue("hello")
	got := ansiEscape.ReplaceAllString(m.documentView(), "")
	if !strings.Contains(got, "▶   1 hello▌") {
		t.Fatalf("document view should show a literal caret after the text, got %q", got)
	}
}

func TestFormattingMenuActionInsertsMarkdown(t *testing.T) {
	m := newModel()
	updated, _ := m.execute("Bold")
	got := updated.(model).editor.Value()
	if got != "**bold text**" {
		t.Fatalf("editor value = %q", got)
	}
}

func TestF10OpensFileMenu(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyF10})
	got := updated.(model)
	if !got.menuOpen || !strings.Contains(got.View(), "› New") {
		t.Fatal("F10 should open the File menu")
	}
}

func TestAltKeyOpensMatchingMenu(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e"), Alt: true})
	got := updated.(model)
	if !got.menuOpen || got.menuIndex != 1 {
		t.Fatal("alt+e should open the Edit menu")
	}
}

func TestMenuNavigationRunsSelectedAction(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.menuOpen = true
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	if m.itemIndex != 1 {
		t.Fatalf("menu item = %d, want Open… at index 1", m.itemIndex)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.prompt != "Open path" {
		t.Fatalf("prompt = %q, want Open path", m.prompt)
	}
}

func TestControlCQuits(t *testing.T) {
	m := newModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should return a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+c should emit tea.QuitMsg")
	}
}

func TestControlCQuitsFromPrompt(t *testing.T) {
	m := newModel()
	m.prompt = "Open path"
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should return a quit command from a prompt")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+c should emit tea.QuitMsg from a prompt")
	}
}

func TestPreviewPreservesLinesAndRendersUnderline(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.editor.SetValue("first line\nsecond line\n\n<u>underlined</u>")
	got := m.renderPreview()
	plain := ansiEscape.ReplaceAllString(got, "")
	for _, want := range []string{"first line", "second line", "underlined"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("preview missing %q: %q", want, plain)
		}
	}
	if strings.Contains(plain, underlineStart) || strings.Contains(plain, underlineEnd) || strings.Contains(plain, "<u>") {
		t.Fatalf("preview leaked underline markup: %q", plain)
	}
	if !strings.Contains(got, "\x1b[4m") {
		t.Fatalf("preview missing terminal underline style: %q", got)
	}
	if !regexp.MustCompile(`first line\s*\n\s*second line`).MatchString(plain) {
		t.Fatalf("preview should preserve ordinary line breaks: %q", plain)
	}
}

func TestPreviewRendersMarkdownHeading(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.editor.SetValue("### Heading")
	plain := ansiEscape.ReplaceAllString(m.renderPreview(), "")
	if !strings.Contains(plain, "Heading") || strings.Contains(plain, "###") {
		t.Fatalf("heading preview should render Markdown, got %q", plain)
	}
}

func TestPreviewRendersMarkdownBullets(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.editor.SetValue("- first item\n- second item\n- [ ] open task\n- [x] done task")
	plain := ansiEscape.ReplaceAllString(m.renderPreview(), "")
	for _, want := range []string{"first item", "second item", "open task", "done task"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("bullet preview missing %q: %q", want, plain)
		}
	}
	if strings.Contains(plain, "- first item") {
		t.Fatalf("bullet preview should not show source dash: %q", plain)
	}
}

func TestSaveWritesCurrentDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.md")
	m := newModel()
	m.path = path
	m.editor.SetValue("# Saved")
	_, _ = m.save()
	got, err := editor.ReadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "# Saved" {
		t.Fatalf("file contents = %q", got)
	}
}
