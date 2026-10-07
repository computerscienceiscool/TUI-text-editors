// Command color-field is a command-palette editor with a deliberately vivid
// visual language for demonstrating color as terminal interaction design.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"tui-text-editors/internal/editor"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

type command struct {
	name, description, category string
}

func (c command) Title() string       { return c.name }
func (c command) Description() string { return c.category + " · " + c.description }
func (c command) FilterValue() string { return c.name + " " + c.category }

var commands = []list.Item{
	command{"New document", "Clear the canvas", "FILE"},
	command{"Open document", "Load Markdown or text", "FILE"},
	command{"Save document", "Write the current draft", "FILE"},
	command{"Heading", "Insert a level-two heading", "FORMAT"},
	command{"Bold", "Insert bold Markdown", "FORMAT"},
	command{"Italic", "Insert italic Markdown", "FORMAT"},
	command{"Underline", "Insert an underline extension", "FORMAT"},
	command{"Bullet list", "Insert a Markdown bullet", "INSERT"},
	command{"Checklist", "Insert an open task", "INSERT"},
	command{"Link", "Insert a Markdown link", "INSERT"},
	command{"Toggle preview", "Show or hide the reading view", "VIEW"},
	command{"Focus mode", "Give the draft the entire canvas", "VIEW"},
}

type model struct {
	editor                          textarea.Model
	palette                         list.Model
	input                           textinput.Model
	width, height                   int
	path, message, prompt           string
	paletteOpen, preview, focusMode bool
}

func newModel() model {
	ed := textarea.New()
	ed.Placeholder = "Start with an idea…"
	ed.ShowLineNumbers = false
	ed.Focus()
	p := list.New(commands, list.NewDefaultDelegate(), 44, 16)
	p.Title = "COMMAND PALETTE"
	p.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("57")).Padding(0, 1)
	p.Styles.StatusBar = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	p.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	p.SetShowHelp(false)
	p.SetShowStatusBar(true)
	return model{editor: ed, palette: p, input: textinput.New(), preview: true, message: "Ctrl+K opens the command palette"}
}

func (m model) Init() tea.Cmd { return m.editor.Focus() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = size.Width, size.Height
		m.resize()
		return m, nil
	}
	if m.prompt != "" {
		if key, ok := msg.(tea.KeyMsg); ok {
			return m.updatePrompt(key)
		}
		return m, nil
	}
	if m.paletteOpen {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "esc", "ctrl+k":
				m.paletteOpen = false
				m.editor.Focus()
				return m, nil
			case "enter":
				if selected, ok := m.palette.SelectedItem().(command); ok {
					return m.execute(selected.name)
				}
			}
		}
		var cmd tea.Cmd
		m.palette, cmd = m.palette.Update(msg)
		return m, cmd
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+k":
			m.paletteOpen = true
			m.palette.ResetFilter()
			return m, nil
		case "ctrl+p":
			m.preview = !m.preview
			m.resize()
			return m, nil
		case "ctrl+s":
			if m.path == "" {
				return m.ask("Save path", "color-field.md")
			}
			return m.save()
		case "ctrl+o":
			return m.ask("Open path", "")
		}
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m *model) resize() {
	width := max(30, m.width-8)
	if m.preview && !m.focusMode {
		width = max(28, (m.width-10)/2)
	}
	m.editor.SetWidth(width)
	m.editor.SetHeight(max(8, m.height-10))
	m.palette.SetSize(max(42, min(70, m.width-12)), max(12, m.height-10))
}

func (m model) execute(name string) (tea.Model, tea.Cmd) {
	m.paletteOpen = false
	m.editor.Focus()
	switch name {
	case "New document":
		m.editor.SetValue("")
		m.path = ""
		m.message = "New color-field document"
	case "Open document":
		return m.ask("Open path", "")
	case "Save document":
		if m.path == "" {
			return m.ask("Save path", "color-field.md")
		}
		return m.save()
	case "Heading":
		m.editor.InsertString("## Heading\n")
	case "Bold":
		m.editor.InsertString("**bold text**")
	case "Italic":
		m.editor.InsertString("*italic text*")
	case "Underline":
		m.editor.InsertString("<u>underlined text</u>")
	case "Bullet list":
		m.editor.InsertString("- list item\n")
	case "Checklist":
		m.editor.InsertString("- [ ] task\n")
	case "Link":
		m.editor.InsertString("[link text](https://example.com)")
	case "Toggle preview":
		m.preview = !m.preview
		m.resize()
	case "Focus mode":
		m.focusMode = !m.focusMode
		m.resize()
	}
	if name != "New document" {
		m.message = name + " selected"
	}
	return m, nil
}

func (m model) ask(prompt, value string) (tea.Model, tea.Cmd) {
	m.prompt = prompt
	m.input.SetValue(value)
	m.input.Focus()
	return m, nil
}

func (m model) updatePrompt(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.String() == "esc" {
		m.prompt = ""
		m.editor.Focus()
		return m, nil
	}
	if key.String() == "enter" {
		value, prompt := m.input.Value(), m.prompt
		m.prompt = ""
		m.editor.Focus()
		if !editor.SupportedDocument(value) {
			m.message = "Use .md, .markdown, or .txt"
			return m, nil
		}
		if prompt == "Open path" {
			contents, err := editor.ReadDocument(value)
			if err != nil {
				m.message = "Open failed: " + err.Error()
				return m, nil
			}
			m.editor.SetValue(contents)
			m.path = value
			m.message = "Opened " + filepath.Base(value)
			return m, nil
		}
		m.path = value
		return m.save()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(key)
	return m, cmd
}

func (m model) save() (tea.Model, tea.Cmd) {
	if err := editor.WriteDocument(m.path, m.editor.Value()); err != nil {
		m.message = "Save failed: " + err.Error()
		return m, nil
	}
	m.message = "Saved " + filepath.Base(m.path)
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.prompt != "" {
		return "\n  " + m.prompt + ": " + m.input.View() + "\n"
	}
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("57")).Padding(0, 1).Render("COLOR FIELD") + " " + chip("FILE", "86") + " " + chip("FORMAT", "212") + " " + chip("INSERT", "221") + " " + chip("VIEW", "141")
	if m.paletteOpen {
		return "\n" + header + "\n\n" + lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("57")).Padding(0, 1).Render(m.palette.View()) + "\n\n  ↑↓ select · Enter run · Esc close · / filter\n"
	}
	left := panel("DRAFT", m.documentView(), max(30, m.editor.Width()), "212")
	content := left
	if m.preview && !m.focusMode {
		content = lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", panel("READING VIEW", m.previewView(), max(30, m.editor.Width()), "86"))
	}
	name := "untitled.md"
	if m.path != "" {
		name = filepath.Base(m.path)
	}
	footer := fmt.Sprintf("%s  ·  %s  ·  Ln %d, Col %d\nctrl+k commands • ctrl+s save • ctrl+o open • ctrl+p preview • ctrl+c quit", name, m.message, m.editor.Line()+1, m.editor.LineInfo().StartColumn+m.editor.LineInfo().ColumnOffset+1)
	return "\n" + header + "\n\n" + content + "\n\n" + footer + "\n"
}

func (m model) documentView() string {
	lines := strings.Split(m.editor.Value(), "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	active := min(m.editor.Line(), len(lines)-1)
	position := m.editor.LineInfo().StartColumn + m.editor.LineInfo().ColumnOffset
	height := max(1, m.editor.Height())
	start := max(0, active-height+1)
	var out strings.Builder
	for row := 0; row < height; row++ {
		line := start + row
		if row > 0 {
			out.WriteByte('\n')
		}
		marker := " "
		if line == active {
			marker = "▶"
		}
		fmt.Fprintf(&out, "%s %3d ", marker, line+1)
		if line >= len(lines) {
			continue
		}
		text := []rune(lines[line])
		if line == active {
			column := min(position, len(text))
			text = append(text, 0)
			copy(text[column+1:], text[column:])
			text[column] = '▌'
		}
		out.WriteString(string(text))
	}
	return out.String()
}

const underlineStart = "COLORFIELDUNDERLINESTART"
const underlineEnd = "COLORFIELDUNDERLINEEND"

var underlineTag = regexp.MustCompile(`(?s)<u>(.*?)</u>`)

func (m model) previewView() string {
	source := underlineTag.ReplaceAllString(m.editor.Value(), underlineStart+"${1}"+underlineEnd)
	if !strings.HasSuffix(source, "\n") {
		source += "\n"
	}
	style := styles.DarkStyleConfig
	style.H1.Prefix, style.H2.Prefix, style.H3.Prefix = "", "", ""
	style.H4.Prefix, style.H5.Prefix, style.H6.Prefix = "", "", ""
	r, err := glamour.NewTermRenderer(glamour.WithStyles(style), glamour.WithWordWrap(max(26, m.editor.Width()-4)), glamour.WithPreservedNewLines())
	if err != nil {
		return err.Error()
	}
	out, err := r.Render(source)
	if err != nil {
		return err.Error()
	}
	return applyUnderlines(out)
}

func applyUnderlines(rendered string) string {
	for {
		start := strings.Index(rendered, underlineStart)
		if start == -1 {
			return rendered
		}
		endOffset := strings.Index(rendered[start+len(underlineStart):], underlineEnd)
		if endOffset == -1 {
			return strings.ReplaceAll(rendered, underlineStart, "")
		}
		end := start + len(underlineStart) + endOffset
		underlined := "\x1b[4m" + strings.ReplaceAll(rendered[start+len(underlineStart):end], "\x1b[0m", "\x1b[0m\x1b[4m") + "\x1b[24m"
		rendered = rendered[:start] + underlined + rendered[end+len(underlineEnd):]
	}
}

func chip(label, color string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color(color)).Padding(0, 1).Render(label)
}
func panel(title, body string, width int, color string) string {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(color)).Padding(0, 1).Width(width).Render(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color)).Render(title) + "\n" + body)
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
