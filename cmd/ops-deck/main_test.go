package main

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func TestOutlineFindsMarkdownHeadings(t *testing.T) {
	got := headings("# Incident\n\n## Decisions\ntext\n### Owner")
	want := []string{"001 Incident", "003 · Decisions", "005 · · Owner"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("headings = %#v, want %#v", got, want)
	}
}

func TestHeadingShortcutUpdatesDocument(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	m = updated.(model)
	if got := m.editor.Value(); got != "## Heading\n" {
		t.Fatalf("heading shortcut inserted %q", got)
	}
}

func TestDocumentViewShowsLiteralCaret(t *testing.T) {
	m := newModel()
	m.width, m.height = 100, 30
	m.resize()
	m.editor.SetValue("status")
	got := ansiEscape.ReplaceAllString(m.documentView(), "")
	if !strings.Contains(got, "▶001 status▌") {
		t.Fatalf("document view should show literal caret, got %q", got)
	}
}

func TestPreviewConsumesHeadingMarkers(t *testing.T) {
	m := newModel()
	m.width, m.height = 110, 36
	m.resize()
	m.editor.SetValue("## Decision")
	plain := ansiEscape.ReplaceAllString(m.previewView(), "")
	if strings.Contains(plain, "##") || !strings.Contains(plain, "Decision") {
		t.Fatalf("preview should consume heading markers: %q", plain)
	}
}
