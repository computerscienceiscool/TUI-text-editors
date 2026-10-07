package main

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"tui-text-editors/internal/editor"
)

func TestTypingReachesFocusedEditor(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	got := updated.(model).editor.Value()
	if got != "hello" {
		t.Fatalf("editor value = %q, want %q", got, "hello")
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
