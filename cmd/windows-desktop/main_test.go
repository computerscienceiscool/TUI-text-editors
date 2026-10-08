package main

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func TestDesktopViewUsesWindowChrome(t *testing.T) {
	m := newModel()
	m.width, m.height = 120, 36
	m.resize()
	view := ansiEscape.ReplaceAllString(m.View(), "")
	for _, want := range []string{"Untitled - Terminal Write", "File  Edit  View", "F2 New", "F3 Save", "PgUp/PgDn scroll", "Document", "Preview"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q: %q", want, view)
		}
	}
}

func TestCtrlGMenuNavigatesAndRunsFormatAction(t *testing.T) {
	m := newModel()
	m.width, m.height = 120, 40
	m.resize()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlG})
	m = updated.(model)
	if !m.menuOpen || m.menuIndex != 0 {
		t.Fatalf("Ctrl+G did not open File menu: open=%t index=%d", m.menuOpen, m.menuIndex)
	}
	for range 3 {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
		m = updated.(model)
	}
	if got := desktopMenus[m.menuIndex].name; got != "Format" {
		t.Fatalf("right navigation selected %q, want Format", got)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.menuOpen || m.editor.Value() != "**bold text**" || m.message != "Bold inserted" {
		t.Fatalf("Format menu action failed: open=%t text=%q message=%q", m.menuOpen, m.editor.Value(), m.message)
	}
}

func TestMenuEscapeClosesWithoutRunningAction(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlG})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.menuOpen {
		t.Fatal("Esc did not close the menu")
	}
}

func TestOpenMenuRendersSelectablePopup(t *testing.T) {
	m := newModel()
	m.width, m.height = 120, 40
	m.resize()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlG})
	m = updated.(model)
	view := ansiEscape.ReplaceAllString(m.View(), "")
	for _, want := range []string{"File", "Edit", "View", "Format", "Help", "► New", "Save As…"} {
		if !strings.Contains(view, want) {
			t.Fatalf("open menu missing %q: %q", want, view)
		}
	}
}

func TestTypedPreviewHasNoBackgroundFill(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.editor.SetValue("typed text")
	preview := m.previewView()
	if !strings.Contains(preview, "\x1b[30;47m") {
		t.Fatalf("typed preview does not restore gray paper: %q", preview)
	}
	if strings.Contains(preview, "\x1b[40m") || strings.Contains(preview, "\x1b[48;") {
		t.Fatalf("typed preview unexpectedly sets a dark background: %q", preview)
	}
}

func TestPageDownMovesTheCursorAndShowsScrollTrack(t *testing.T) {
	m := newModel()
	m.width, m.height = 80, 20
	m.resize()
	m.editor.SetValue(strings.Repeat("line\n", 40))
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updated.(model)
	if m.editor.Line() == 0 || m.message != "Scrolled down" {
		t.Fatalf("page down did not move cursor: line %d, message %q", m.editor.Line(), m.message)
	}
	if got := ansiEscape.ReplaceAllString(m.documentView(), ""); !strings.Contains(got, "█") {
		t.Fatalf("scroll track missing: %q", got)
	}
}

func TestPreviewWindowFillsTheDocumentHeight(t *testing.T) {
	m := newModel()
	m.width, m.height = 120, 36
	m.resize()
	preview := ansiEscape.ReplaceAllString(window("Preview", m.previewView(), m.editor.Width(), m.editor.Height()), "")
	if got, want := len(strings.Split(preview, "\n")), m.editor.Height()+3; got != want {
		t.Fatalf("preview window has %d lines, want %d", got, want)
	}
}

func TestF5InsertsHeading(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyF5})
	m = updated.(model)
	if got := m.editor.Value(); got != "## Heading\n" {
		t.Fatalf("heading inserted %q", got)
	}
}

func TestDocumentViewShowsLiteralCaret(t *testing.T) {
	m := newModel()
	m.width, m.height = 80, 20
	m.resize()
	m.editor.SetValue("memo")
	got := ansiEscape.ReplaceAllString(m.documentView(), "")
	if !strings.Contains(got, "►   1 memo▌") {
		t.Fatalf("caret missing: %q", got)
	}
}

func TestPreviewConsumesHeadingMarkers(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.editor.SetValue("## Heading")
	plain := ansiEscape.ReplaceAllString(m.previewView(), "")
	if strings.Contains(plain, "##") || !strings.Contains(plain, "Heading") {
		t.Fatalf("heading preview = %q", plain)
	}
}
