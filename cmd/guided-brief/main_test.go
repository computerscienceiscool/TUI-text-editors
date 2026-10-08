package main

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func TestTemplateUsesBriefChoices(t *testing.T) {
	m := newModel()
	m.brief.documentType = "incident-update"
	m.brief.audience = "leadership"
	m.brief.tone = []string{"concise", "technical"}
	m.brief.includes = []string{"summary", "risks"}
	got := m.template()
	for _, want := range []string{"# Incident Update", "For leadership", "concise, technical", "## Summary", "## Risks"} {
		if !strings.Contains(got, want) {
			t.Fatalf("template missing %q: %q", want, got)
		}
	}
}

func TestEditorShortcutsInsertMarkdown(t *testing.T) {
	m := newModel()
	m.screen = editorScreen
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	m = updated.(model)
	if got := m.editor.Value(); got != "## Heading" {
		t.Fatalf("heading shortcut inserted %q", got)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	m = updated.(model)
	if !strings.Contains(m.editor.Value(), "- list item") {
		t.Fatalf("list shortcut did not insert a list item: %q", m.editor.Value())
	}
}

func TestGuidedBriefStartsWithPurposeForm(t *testing.T) {
	m := newModel()
	if m.screen != briefScreen {
		t.Fatalf("screen = %v, want brief screen", m.screen)
	}
	if m.form == nil {
		t.Fatal("guided brief should configure a Huh form")
	}
	if m.brief == nil {
		t.Fatal("guided brief should retain shared form answers")
	}
}

func TestEditorMenuNavigatesAndRunsFormatAction(t *testing.T) {
	m := newModel()
	m.screen = editorScreen
	m.width, m.height = 110, 40
	m.resize()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Alt: true, Runes: []rune{'m'}})
	m = updated.(model)
	if !m.menuOpen {
		t.Fatal("Alt+M did not open the editor menu")
	}
	for range 2 {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
		m = updated.(model)
	}
	if got := guidedMenus[m.menuIndex].name; got != "Format" {
		t.Fatalf("selected menu %q, want Format", got)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.menuOpen || m.editor.Value() != "**bold text**" {
		t.Fatalf("menu did not insert bold text: open=%t text=%q", m.menuOpen, m.editor.Value())
	}
}

func TestGeneratedBriefPreviewConsumesMarkdownMarkers(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.brief.documentType = "proposal"
	m.brief.audience = "team"
	m.brief.includes = []string{"summary"}
	m.editor.SetValue(m.template())
	plain := ansiEscape.ReplaceAllString(m.previewView(), "")
	for _, marker := range []string{"# Proposal", "## Summary", "_For team"} {
		if strings.Contains(plain, marker) {
			t.Fatalf("preview leaked source marker %q: %q", marker, plain)
		}
	}
	for _, want := range []string{"Proposal", "Summary", "For team"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("preview missing %q: %q", want, plain)
		}
	}
}
