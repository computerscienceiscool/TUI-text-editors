package main

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func TestPlaygroundShowsPromptAndActions(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	view := ansiEscape.ReplaceAllString(m.View(), "")
	for _, want := range []string{"PLAYGROUND", "WRITING SPARK", "[Tab] mood", "[F1] next prompt", "[F8] sparkle"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q: %q", want, view)
		}
	}
}

func TestF1CyclesWritingPrompt(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyF1})
	m = updated.(model)
	if m.promptIndex != 1 || m.message != prompts[1] {
		t.Fatalf("F1 should cycle prompt, got index %d and %q", m.promptIndex, m.message)
	}
}
func TestF8InsertsSparkle(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyF8})
	m = updated.(model)
	if got := m.editor.Value(); got != "✨ " {
		t.Fatalf("sparkle inserted %q", got)
	}
}
func TestDocumentViewShowsLiteralCaret(t *testing.T) {
	m := newModel()
	m.width, m.height = 80, 20
	m.resize()
	m.editor.SetValue("idea")
	got := ansiEscape.ReplaceAllString(m.documentView(), "")
	if !strings.Contains(got, "✦   1 idea▌") {
		t.Fatalf("caret missing: %q", got)
	}
}
func TestPreviewConsumesHeadingMarkers(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.editor.SetValue("## Idea")
	plain := ansiEscape.ReplaceAllString(m.previewView(), "")
	if strings.Contains(plain, "##") || !strings.Contains(plain, "Idea") {
		t.Fatalf("preview = %q", plain)
	}
}
