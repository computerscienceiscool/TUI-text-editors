// Command signal-accessible demonstrates redundant, high-clarity TUI cues.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"tui-text-editors/internal/editor"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

type menu struct {
	name  string
	items []menuItem
}
type menuItem struct{ label, action string }

var signalMenus = []menu{
	{"File", []menuItem{{"New", "new"}, {"Open…", "open"}, {"Save", "save"}, {"Save As…", "save-as"}}},
	{"Edit", []menuItem{{"Cursor up", "up"}, {"Cursor down", "down"}}},
	{"View", []menuItem{{"Split preview", "preview"}}},
	{"Format", []menuItem{{"Heading", "heading"}, {"Bullet", "bullet"}, {"Bold", "bold"}, {"Italic", "italic"}, {"Underline", "underline"}}},
	{"Help", []menuItem{{"Controls", "controls"}, {"About", "about"}}},
}

type model struct {
	editor                textarea.Model
	input                 textinput.Model
	width, height         int
	path, message, prompt string
	preview               bool
	menuIndex, itemIndex  int
	menuOpen              bool
}

func newModel() model {
	ed := textarea.New()
	ed.Placeholder = "Start writing…"
	ed.ShowLineNumbers = false
	ed.Focus()
	return model{editor: ed, input: textinput.New(), preview: true, message: "Editor ready"}
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
	if key, ok := msg.(tea.KeyMsg); ok {
		if m.menuOpen {
			return m.updateMenu(key)
		}
		switch key.String() {
		case "f10", "alt+m":
			m.menuOpen, m.itemIndex = true, 0
			m.resize()
			return m, nil
		case "alt+f", "alt+e", "alt+v", "alt+r", "alt+h":
			m.menuIndex = map[string]int{"alt+f": 0, "alt+e": 1, "alt+v": 2, "alt+r": 3, "alt+h": 4}[key.String()]
			m.menuOpen, m.itemIndex = true, 0
			m.resize()
			return m, nil
		case "f2":
			m.editor.SetValue("")
			m.path = ""
			m.message = "New document created"
			return m, nil
		case "f3":
			if m.path == "" {
				return m.ask("Save path", "signal.md")
			}
			return m.save()
		case "f4":
			return m.ask("Open path", "")
		case "f5":
			m.editor.InsertString("## Heading\n")
			m.message = "Heading inserted"
			return m, nil
		case "f6":
			m.editor.InsertString("- list item\n")
			m.message = "Bullet inserted"
			return m, nil
		case "f7":
			m.preview = !m.preview
			m.resize()
			m.message = toggle("Preview", m.preview)
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}
func (m model) updateMenu(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc", "f10", "alt+m":
		m.menuOpen = false
		m.resize()
		return m, nil
	case "left":
		m.menuIndex = (m.menuIndex + len(signalMenus) - 1) % len(signalMenus)
		m.itemIndex = 0
	case "right":
		m.menuIndex = (m.menuIndex + 1) % len(signalMenus)
		m.itemIndex = 0
	case "up":
		items := signalMenus[m.menuIndex].items
		m.itemIndex = (m.itemIndex + len(items) - 1) % len(items)
	case "down":
		items := signalMenus[m.menuIndex].items
		m.itemIndex = (m.itemIndex + 1) % len(items)
	case "enter":
		return m.executeMenu(signalMenus[m.menuIndex].items[m.itemIndex])
	}
	return m, nil
}
func (m model) executeMenu(item menuItem) (tea.Model, tea.Cmd) {
	m.menuOpen = false
	m.resize()
	switch item.action {
	case "new":
		m.editor.SetValue("")
		m.path, m.message = "", "New document created"
	case "open":
		return m.ask("Open path", "")
	case "save":
		if m.path == "" {
			return m.ask("Save path", "signal.md")
		}
		return m.save()
	case "save-as":
		return m.ask("Save path", m.path)
	case "up":
		m.editor.CursorUp()
		m.message = "Cursor up"
	case "down":
		m.editor.CursorDown()
		m.message = "Cursor down"
	case "preview":
		m.preview = !m.preview
		m.resize()
		m.message = toggle("Preview", m.preview)
	case "heading":
		m.editor.InsertString("## Heading\n")
		m.message = "Heading inserted"
	case "bullet":
		m.editor.InsertString("- list item\n")
		m.message = "Bullet inserted"
	case "bold", "italic", "underline":
		m.editor.InsertString(editor.InsertFormatting(item.action))
		m.message = strings.Title(item.action) + " inserted"
	case "controls":
		m.message = "Alt+M menus · arrows select · Enter runs"
	case "about":
		m.message = "Signal — high-clarity editor"
	}
	return m, nil
}
func (m *model) resize() {
	width := max(34, m.width-8)
	if m.preview {
		width = max(28, (m.width-10)/2)
	}
	m.editor.SetWidth(width)
	menuRows := 0
	if m.menuOpen {
		menuRows = 7
	}
	// The permanent bordered menu row occupies three terminal rows.
	m.editor.SetHeight(max(8, m.height-13-menuRows))
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
			body, err := editor.ReadDocument(value)
			if err != nil {
				m.message = "Open failed: " + err.Error()
				return m, nil
			}
			m.editor.SetValue(body)
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
		return "\n[INPUT REQUIRED] " + m.prompt + ": " + m.input.View() + "\n"
	}
	head := lipgloss.NewStyle().Bold(true).Border(lipgloss.DoubleBorder(), false, false, true, false).Render("SIGNAL — HIGH CLARITY EDITOR")
	menu := m.menuBar()
	if m.menuOpen {
		menu += "\n" + m.menuPopup()
	}
	body := panel("[EDITING] Document input", m.documentView(), max(34, m.editor.Width()))
	if m.preview {
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, "  ", panel("[READING] Markdown preview", m.previewView(), max(30, m.editor.Width())))
	}
	name := "untitled.md"
	if m.path != "" {
		name = filepath.Base(m.path)
	}
	footer := fmt.Sprintf("[STATUS] %s | %s | Ln %d, Col %d\n[Alt+M] Menus  [F2] New  [F3] Save  [F4] Open  [F5] Heading  [F6] Bullet  [F7] Preview  [Ctrl+C] Quit", name, m.message, m.editor.Line()+1, m.editor.LineInfo().StartColumn+m.editor.LineInfo().ColumnOffset+1)
	return "\n" + head + "\n" + menu + "\n\n" + body + "\n\n" + footer + "\n"
}
func (m model) menuBar() string {
	parts := make([]string, len(signalMenus))
	for i, item := range signalMenus {
		style := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
		if m.menuOpen && i == m.menuIndex {
			style = style.Reverse(true)
		}
		parts[i] = style.Render(item.name)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
func (m model) menuPopup() string {
	items := signalMenus[m.menuIndex].items
	rows := make([]string, len(items))
	for i, item := range items {
		prefix := "  "
		if i == m.itemIndex {
			prefix = "► "
		}
		rows[i] = prefix + item.label
	}
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).Padding(0, 1).Render(strings.Join(rows, "\n"))
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
			marker = ">"
		}
		fmt.Fprintf(&out, "%s %3d ", marker, line+1)
		if line >= len(lines) {
			continue
		}
		text := []rune(lines[line])
		if line == active {
			col := min(position, len(text))
			text = append(text, 0)
			copy(text[col+1:], text[col:])
			text[col] = '▌'
		}
		out.WriteString(string(text))
	}
	return out.String()
}

const underlineStart = "SIGNALUNDERLINESTART"
const underlineEnd = "SIGNALUNDERLINEEND"

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
		value := "\x1b[4m" + strings.ReplaceAll(rendered[start+len(underlineStart):end], "\x1b[0m", "\x1b[0m\x1b[4m") + "\x1b[24m"
		rendered = rendered[:start] + value + rendered[end+len(underlineEnd):]
	}
}
func panel(title, body string, width int) string {
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).Padding(0, 1).Width(width).Render(lipgloss.NewStyle().Bold(true).Render(title) + "\n" + body)
}
func toggle(label string, on bool) string {
	if on {
		return label + " on"
	}
	return label + " off"
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
