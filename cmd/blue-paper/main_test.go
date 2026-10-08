package main

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func TestControlKOpensCommandPalette(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	m = updated.(model)
	if !m.paletteOpen || !strings.Contains(m.View(), "BLUE PAPER MENU") {
		t.Fatalf("Ctrl+K should open command palette, got %q", m.View())
	}
}

func TestAltMMenuOpensCategoryCommands(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Alt: true, Runes: []rune{'m'}})
	m = updated.(model)
	if !m.paletteOpen || !strings.Contains(m.View(), "BLUE PAPER MENU") {
		t.Fatalf("Alt+M did not open actionable menu: %q", m.View())
	}
}

func TestMouseClickOpensFormatMenu(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.MouseMsg{X: 22, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = updated.(model)
	if !m.paletteOpen || m.category != "FORMAT" {
		t.Fatalf("click did not open Format menu: open=%t category=%q", m.paletteOpen, m.category)
	}
}

func TestSecondMenuClickReplacesTheFirstMenu(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.MouseMsg{X: 14, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = updated.(model)
	updated, _ = m.Update(tea.MouseMsg{X: 22, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = updated.(model)
	if !m.paletteOpen || m.category != "FORMAT" || len(m.palette.Items()) != 4 {
		t.Fatalf("second click did not replace menu: open=%t category=%q items=%d", m.paletteOpen, m.category, len(m.palette.Items()))
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

func TestPaletteCategoryFilterShowsOnlyMatchingCommands(t *testing.T) {
	m := newModel()
	updated, _ := m.setCategory("FORMAT")
	m = updated.(model)
	if m.category != "FORMAT" || len(m.palette.Items()) != 4 {
		t.Fatalf("format filter = %q with %d items", m.category, len(m.palette.Items()))
	}
	for _, item := range m.palette.Items() {
		if item.(command).category != "FORMAT" {
			t.Fatalf("unexpected filtered command: %#v", item)
		}
	}
}

func TestPaletteCommandTitleIsPlainForFiltering(t *testing.T) {
	entry := command{name: "Save document", category: "FILE"}
	if got := entry.Title(); got != "Save document" {
		t.Fatalf("title must be plain command text, got %q", got)
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

func TestPreviewAndMenuRestoreWhitePaperAfterStyleResets(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.editor.SetValue("typed text")
	if got := m.previewView(); !strings.Contains(got, "48;5;15") || strings.Contains(got, "\x1b[40m") {
		t.Fatalf("preview did not preserve white paper: %q", got)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	m = updated.(model)
	if got := paperSurface(m.palette.View()); !strings.Contains(got, "48;5;15") || strings.Contains(got, "\x1b[40m") {
		t.Fatalf("menu did not preserve white paper: %q", got)
	}
}

func TestCanvasPaintsEntireViewport(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	m := newModel()
	m.width, m.height = 110, 36
	view := m.paperCanvas("blue paper")
	if !strings.Contains(view, "48;5;15") && !strings.Contains(view, "107m") {
		t.Fatalf("canvas does not paint white paper: %q", view)
	}
}

func TestCanvasNeverLeaksBlackAndPaintsEveryRow(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	m := newModel()
	m.width, m.height = 90, 12
	view := m.paperCanvas("left panel\npreview panel")
	if strings.Contains(view, "\x1b[40m") || strings.Contains(view, "48;5;0") {
		t.Fatalf("canvas leaked a black background: %q", view)
	}
	rows := strings.Split(view, "\n")
	if len(rows) != m.height {
		t.Fatalf("canvas has %d rows, want %d", len(rows), m.height)
	}
	for row, line := range rows {
		if !strings.Contains(line, "107m") && !strings.Contains(line, "48;5;15") {
			t.Fatalf("row %d has no white-paper background: %q", row, line)
		}
	}
}
