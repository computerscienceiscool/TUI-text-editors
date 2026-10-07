package main

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func TestControlKOpensCommandPalette(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	m = updated.(model)
	if !m.paletteOpen || !strings.Contains(m.View(), "COMMAND PALETTE") {
		t.Fatalf("Ctrl+K should open command palette, got %q", m.View())
	}
}

func TestPaletteCommandInsertsMarkdown(t *testing.T) {
	m := newModel()
	updated, _ := m.execute("Heading")
	m = updated.(model)
	if got := m.editor.Value(); got != "## Heading\n" {
		t.Fatalf("heading command inserted %q", got)
	}
}

func TestDocumentViewShowsLiteralCaret(t *testing.T) {
	m := newModel()
	m.width, m.height = 80, 20
	m.resize()
	m.editor.SetValue("idea")
	if got := ansiEscape.ReplaceAllString(m.documentView(), ""); !strings.Contains(got, "▶   1 idea▌") {
		t.Fatalf("document view should show a literal caret, got %q", got)
	}
}

func TestPreviewConsumesHeadingMarkers(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.editor.SetValue("## Heading")
	plain := ansiEscape.ReplaceAllString(m.previewView(), "")
	if strings.Contains(plain, "##") || !strings.Contains(plain, "Heading") {
		t.Fatalf("preview should render heading without source markers: %q", plain)
	}
}
