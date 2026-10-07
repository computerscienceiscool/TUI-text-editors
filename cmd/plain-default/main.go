// Command tui-text-editors is a small, presentation-focused Markdown editor.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tui-text-editors/internal/editor"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type menu struct {
	name  string
	items []string
}

var menus = []menu{
	{"File", []string{"New", "Open…", "Save", "Save As…", "Recent Files", "Quit"}},
	{"Edit", []string{"Undo", "Redo", "Cut", "Copy", "Paste", "Find…"}},
	{"Format", []string{"Heading 1", "Heading 2", "Bold", "Italic", "Underline", "Clear Formatting"}},
	{"Insert", []string{"Link", "Checklist", "Quote", "Date / time"}},
	{"View", []string{"Line numbers", "Word wrap", "Split preview", "Focus mode"}},
	{"Help", []string{"Keyboard shortcuts", "About"}},
}

type keyMap struct{ menu, preview, quit key.Binding }

func (k keyMap) ShortHelp() []key.Binding  { return []key.Binding{k.menu, k.preview, k.quit} }
func (k keyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

type model struct {
	editor                                    textarea.Model
	input                                     textinput.Model
	help                                      help.Model
	keys                                      keyMap
	width, height                             int
	path, message, prompt                     string
	menuIndex, itemIndex                      int
	menuOpen, preview, lineNumbers, focusMode bool
	recent                                    []string
}

func newModel() model {
	editor := textarea.New()
	editor.Placeholder = "Start writing…"
	editor.ShowLineNumbers = true
	editor.Focus()
	return model{editor: editor, input: textinput.New(), help: help.New(), preview: true, lineNumbers: true,
		keys:    keyMap{menu: key.NewBinding(key.WithKeys("alt+m"), key.WithHelp("alt+m", "menu")), preview: key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "preview")), quit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
		message: "New Markdown document"}
}

func (m model) Init() tea.Cmd { return tea.Batch(m.editor.Focus(), textarea.Blink) }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
	case tea.KeyMsg:
		if m.prompt != "" {
			return m.updatePrompt(msg)
		}
		if key.Matches(msg, m.keys.quit) {
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.menu) {
			m.menuOpen = !m.menuOpen
			return m, nil
		}
		if key.Matches(msg, m.keys.preview) {
			m.preview = !m.preview
			m.resize()
			return m, nil
		}
		if m.menuOpen {
			return m.updateMenu(msg)
		}
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m *model) resize() {
	width := m.width - 6
	if m.preview && !m.focusMode {
		width = width / 2
	}
	if width < 28 {
		width = 28
	}
	m.editor.SetWidth(width)
	m.editor.SetHeight(max(8, m.height-9))
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.menuOpen = false
	case "left", "h":
		m.menuIndex = (m.menuIndex + len(menus) - 1) % len(menus)
		m.itemIndex = 0
	case "right", "l":
		m.menuIndex = (m.menuIndex + 1) % len(menus)
		m.itemIndex = 0
	case "up", "k":
		m.itemIndex = (m.itemIndex + len(menus[m.menuIndex].items) - 1) % len(menus[m.menuIndex].items)
	case "down", "j":
		m.itemIndex = (m.itemIndex + 1) % len(menus[m.menuIndex].items)
	case "enter":
		return m.execute(menus[m.menuIndex].items[m.itemIndex])
	}
	return m, nil
}

func (m model) execute(action string) (tea.Model, tea.Cmd) {
	m.menuOpen = false
	switch action {
	case "New":
		m.editor.SetValue("")
		m.path = ""
		m.message = "New Markdown document"
	case "Open…":
		return m.ask("Open path", "")
	case "Save":
		if m.path == "" {
			return m.ask("Save path", "untitled.md")
		}
		return m.save()
	case "Save As…":
		return m.ask("Save path", m.path)
	case "Recent Files":
		m.message = "Recent: " + strings.Join(m.recent, ", ")
	case "Quit":
		return m, tea.Quit
	case "Find…":
		return m.ask("Find", "")
	case "Line numbers":
		m.lineNumbers = !m.lineNumbers
		m.editor.ShowLineNumbers = m.lineNumbers
		m.message = toggleMessage("Line numbers", m.lineNumbers)
	case "Word wrap":
		m.message = "Word wrap is enabled in the text area"
	case "Split preview":
		m.preview = !m.preview
		m.resize()
		m.message = toggleMessage("Split preview", m.preview)
	case "Focus mode":
		m.focusMode = !m.focusMode
		m.resize()
		m.message = toggleMessage("Focus mode", m.focusMode)
	case "Keyboard shortcuts":
		m.message = "Alt+M menu · Ctrl+P preview · Ctrl+C quit"
	case "About":
		m.message = "Plain Default — a Charm learning editor"
	case "Undo", "Redo", "Cut", "Copy", "Paste", "Clear Formatting":
		m.message = action + " is intentionally a demo command"
	default:
		kind := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(action, " ", "-"), "/", ""))
		if action == "Date / time" {
			kind = "date"
		}
		m.editor.InsertString(editor.InsertFormatting(kind))
		m.message = action + " inserted"
	}
	return m, nil
}

func (m model) ask(prompt, value string) (tea.Model, tea.Cmd) {
	m.prompt = prompt
	m.input.SetValue(value)
	m.input.Focus()
	return m, nil
}

func (m model) updatePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.prompt = ""
		m.editor.Focus()
		return m, nil
	}
	if msg.String() == "enter" {
		value := m.input.Value()
		prompt := m.prompt
		m.prompt = ""
		m.editor.Focus()
		if prompt == "Find" {
			m.message = fmt.Sprintf("Found %d match(es) for %q", strings.Count(strings.ToLower(m.editor.Value()), strings.ToLower(value)), value)
			return m, nil
		}
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
			m.addRecent(value)
			m.message = "Opened " + filepath.Base(value)
			return m, nil
		}
		m.path = value
		return m.save()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) save() (tea.Model, tea.Cmd) {
	if err := editor.WriteDocument(m.path, m.editor.Value()); err != nil {
		m.message = "Save failed: " + err.Error()
		return m, nil
	}
	m.addRecent(m.path)
	m.message = "Saved " + filepath.Base(m.path)
	return m, nil
}

func (m *model) addRecent(path string) {
	for _, old := range m.recent {
		if old == path {
			return
		}
	}
	m.recent = append([]string{path}, m.recent...)
	if len(m.recent) > 5 {
		m.recent = m.recent[:5]
	}
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading editor…"
	}
	bar := m.menuBar()
	if m.menuOpen {
		bar += "\n" + m.menuPopup()
	}
	if m.prompt != "" {
		return bar + "\n\n  " + m.prompt + ": " + m.input.View() + "\n"
	}
	editor := panel("DOCUMENT", m.editor.View(), max(30, m.editor.Width()))
	content := editor
	if m.preview && !m.focusMode {
		content = lipgloss.JoinHorizontal(lipgloss.Top, editor, "  ", panel("PREVIEW · Glamour", m.renderPreview(), max(30, m.editor.Width())))
	}
	footer := status(m.path, m.message) + "\n" + m.help.View(m.keys)
	return "\n" + bar + "\n\n" + content + "\n\n" + footer + "\n"
}

func (m model) menuBar() string {
	parts := make([]string, len(menus))
	for i, item := range menus {
		style := lipgloss.NewStyle().Padding(0, 1)
		if m.menuOpen && i == m.menuIndex {
			style = style.Reverse(true)
		}
		parts[i] = style.Render(item.name)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
func (m model) menuPopup() string {
	items := menus[m.menuIndex].items
	lines := make([]string, len(items))
	for i, item := range items {
		prefix := "  "
		if i == m.itemIndex {
			prefix = "› "
		}
		lines[i] = prefix + item
	}
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).Render(strings.Join(lines, "\n"))
}
func (m model) renderPreview() string {
	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(max(26, m.editor.Width()-4)))
	if err != nil {
		return err.Error()
	}
	out, err := r.Render(m.editor.Value())
	if err != nil {
		return err.Error()
	}
	return out
}
func panel(title, body string, width int) string {
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).Width(width).Render(lipgloss.NewStyle().Bold(true).Render(title) + "\n" + body)
}
func status(path, message string) string {
	name := "untitled.md"
	if path != "" {
		name = filepath.Base(path)
	}
	return lipgloss.NewStyle().Faint(true).Render(name + "  ·  " + message)
}
func toggleMessage(label string, on bool) string {
	if on {
		return label + " on"
	}
	return label + " off"
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
