package main

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func TestHighClarityViewUsesExplicitLabels(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	view := ansiEscape.ReplaceAllString(m.View(), "")
	for _, want := range []string{"[EDITING]", "[READING]", "[STATUS]", "[F2] New"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q: %q", want, view)
		}
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
	m.editor.SetValue("clear")
	if got := ansiEscape.ReplaceAllString(m.documentView(), ""); !strings.Contains(got, ">   1 clear▌") {
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
