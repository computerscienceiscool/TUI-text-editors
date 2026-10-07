package editor

import (
	"path/filepath"
	"testing"
)

func TestSupportedDocument(t *testing.T) {
	for _, path := range []string{"note.md", "NOTE.MARKDOWN", "plain.txt"} {
		if !SupportedDocument(path) {
			t.Errorf("%q should be supported", path)
		}
	}
	if SupportedDocument("image.png") {
		t.Error("image.png should not be supported")
	}
}

func TestReadWriteDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.md")
	want := "# A note\n\nHello"
	if err := WriteDocument(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := ReadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInsertFormatting(t *testing.T) {
	if got := InsertFormatting("bold"); got != "**bold text**" {
		t.Fatalf("got %q", got)
	}
	if got := InsertFormatting("unknown"); got != "" {
		t.Fatalf("got %q", got)
	}
}
