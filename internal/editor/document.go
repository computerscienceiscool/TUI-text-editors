package editor

import (
	"os"
	"path/filepath"
	"strings"
)

func SupportedDocument(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown" || ext == ".txt"
}

func ReadDocument(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func WriteDocument(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o644)
}

func InsertFormatting(kind string) string {
	switch kind {
	case "heading-1":
		return "# Heading 1"
	case "heading-2":
		return "## Heading 2"
	case "bold":
		return "**bold text**"
	case "italic":
		return "_italic text_"
	case "underline":
		return "<u>underlined text</u>"
	case "link":
		return "[link text](https://example.com)"
	case "checklist":
		return "- [ ] A task"
	case "quote":
		return "> A quote"
	case "date":
		return "2026-10-07"
	}
	return ""
}
