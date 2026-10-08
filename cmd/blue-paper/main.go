// Command blue-paper is a white-paper, Windows-blue editor demonstration.
package main

import (
	"fmt"
	"io"
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
	"github.com/muesli/termenv"
)

type command struct {
	name, description, category string
}

func (c command) Title() string       { return c.name }
func (c command) Description() string { return c.category + " · " + c.description }
func (c command) FilterValue() string { return c.name + " " + c.category }

// paletteDelegate applies colors after list filtering has processed plain
// command text. Rendering ANSI sequences inside Title caused the stock
// delegate to display escape fragments while navigating filtered results.
type paletteDelegate struct{}

func (paletteDelegate) Height() int                         { return 2 }
func (paletteDelegate) Spacing() int                        { return 1 }
func (paletteDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (paletteDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(command)
	if !ok {
		return
	}
	prefix := "  "
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Background(lipgloss.Color("15"))
	if index == m.Index() {
		prefix = "› "
		nameStyle = nameStyle.Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
	}
	title := prefix + chip(entry.category, categoryColor(entry.category)) + " " + entry.name
	detail := "   " + entry.description
	fmt.Fprint(w, nameStyle.Render(title)+"\n"+lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Background(lipgloss.Color("15")).Render(detail))
}

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
	menuOpen                        bool
	menuIndex                       int
	category                        string
}

func newModel() model {
	ed := textarea.New()
	ed.Placeholder = "Start with an idea…"
	ed.ShowLineNumbers = false
	ed.Focus()
	p := list.New(commands, paletteDelegate{}, 44, 16)
	p.Title = "BLUE PAPER MENU"
	p.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4")).Padding(0, 1)
	p.Styles.StatusBar = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	p.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	p.SetShowHelp(false)
	p.SetShowStatusBar(true)
	return model{editor: ed, palette: p, input: textinput.New(), preview: true, category: "ALL", message: "Alt+M opens menus"}
}

func (m model) Init() tea.Cmd { return m.editor.Focus() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if mouse, ok := msg.(tea.MouseMsg); ok && mouse.Button == tea.MouseButtonLeft && mouse.Action == tea.MouseActionPress && mouse.Y == 1 {
		for index, start := range []int{13, 20, 29, 38} {
			if mouse.X >= start && mouse.X < start+8 {
				m.paletteOpen = true
				return m.setCategory([]string{"FILE", "FORMAT", "INSERT", "VIEW"}[index])
			}
		}
	}
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
			case "0":
				return m.setCategory("ALL")
			case "1":
				return m.setCategory("FILE")
			case "2":
				return m.setCategory("FORMAT")
			case "3":
				return m.setCategory("INSERT")
			case "4":
				return m.setCategory("VIEW")
			}
		}
		var cmd tea.Cmd
		m.palette, cmd = m.palette.Update(msg)
		return m, cmd
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "alt+m":
			m.paletteOpen = true
			m.palette.ResetFilter()
			m.category = "ALL"
			return m.setCategory("ALL")
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

func (m model) setCategory(category string) (tea.Model, tea.Cmd) {
	m.category = category
	m.palette.ResetFilter()
	items := make([]list.Item, 0, len(commands))
	for _, item := range commands {
		entry := item.(command)
		if category == "ALL" || entry.category == category {
			items = append(items, entry)
		}
	}
	cmd := m.palette.SetItems(items)
	m.message = "Palette: " + category
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
		return m.paperCanvas("\n  " + m.prompt + ": " + m.input.View() + "\n")
	}
	labels := []string{"FILE", "FORMAT", "INSERT", "VIEW"}
	parts := []string{lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4")).Padding(0, 1).Render("BLUE PAPER")}
	for _, label := range labels {
		s := chip(label, categoryColor(label))
		parts = append(parts, s)
	}
	header := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	if m.paletteOpen {
		return m.paperCanvas("\n" + header + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Background(lipgloss.Color("15")).Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("4")).Padding(0, 1).Render(paperSurface(m.palette.View())) + "\n\n  0 all · 1 file · 2 format · 3 insert · 4 view · ↑↓ select · Enter run · Esc close · / filter\n")
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
	footer := fmt.Sprintf("%s  ·  %s  ·  Ln %d, Col %d\nalt+m / ctrl+k menus • 1 file • 2 format • 3 insert • 4 view • ctrl+c quit", name, m.message, m.editor.Line()+1, m.editor.LineInfo().StartColumn+m.editor.LineInfo().ColumnOffset+1)
	return m.paperCanvas("\n" + header + "\n\n" + content + "\n\n" + footer + "\n")
}

// paperCanvas paints the entire terminal viewport. Styling only the content
// leaves unused cells at the terminal's default (often black) background.
func (m model) paperCanvas(content string) string {
	paper := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Background(lipgloss.Color("15"))
	lines := strings.Split(paperSurface(content), "\n")
	for len(lines) < m.height {
		lines = append(lines, "")
	}
	if len(lines) > m.height {
		lines = lines[:m.height]
	}
	for i, line := range lines {
		padding := max(0, m.width-lipgloss.Width(line))
		lines[i] = paper.Render(line + strings.Repeat(" ", padding))
	}
	return strings.Join(lines, "\n")
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
var paperBackground = regexp.MustCompile(`\x1b\[(?:4[0-9]|48;5;[0-9]+)m`)

func (m model) previewView() string {
	source := underlineTag.ReplaceAllString(m.editor.Value(), underlineStart+"${1}"+underlineEnd)
	if !strings.HasSuffix(source, "\n") {
		source += "\n"
	}
	style := styles.LightStyleConfig
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
	return paperSurface(applyUnderlines(out))
}

func paperSurface(rendered string) string {
	rendered = paperBackground.ReplaceAllString(rendered, "")
	rendered = strings.ReplaceAll(rendered, "\x1b[0m", "\x1b[0;38;5;4;48;5;15m")
	return "\x1b[38;5;4;48;5;15m" + rendered + "\x1b[0m"
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
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color(color)).Padding(0, 1).Render(label)
}
func categoryColor(category string) string {
	switch category {
	case "FILE":
		return "4"
	case "FORMAT":
		return "12"
	case "INSERT":
		return "33"
	case "VIEW":
		return "25"
	default:
		return "250"
	}
}
func panel(title, body string, width int, color string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Background(lipgloss.Color("15")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(color)).Padding(0, 1).Width(width).Render(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color)).Render(title) + "\n" + body)
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
	// Color Field is specifically a color-language demo. Do not let an
	// indeterminate terminal profile silently turn it into monochrome output.
	lipgloss.SetColorProfile(termenv.ANSI256)
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
