// Command guided-brief demonstrates a purpose-first terminal editor.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tui-text-editors/internal/editor"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	briefScreen screen = iota
	editorScreen
)

type menu struct {
	name  string
	items []menuItem
}

type menuItem struct{ label, action string }

var guidedMenus = []menu{
	{"File", []menuItem{{"New brief", "brief"}, {"Open…", "open"}, {"Save", "save"}, {"Save As…", "save-as"}}},
	{"View", []menuItem{{"Split preview", "preview"}}},
	{"Format", []menuItem{{"Heading", "heading"}, {"Bold", "bold"}, {"Italic", "italic"}, {"Underline", "underline"}, {"Bullet list", "bullet"}}},
	{"Help", []menuItem{{"Keyboard controls", "controls"}, {"About", "about"}}},
}

type model struct {
	form                  *huh.Form
	screen                screen
	editor                textarea.Model
	input                 textinput.Model
	width, height         int
	brief                 *briefAnswers
	preview               bool
	prompt, message, path string
	menuIndex, itemIndex  int
	menuOpen              bool
}

// briefAnswers is shared by Huh and Bubble Tea. Bubble Tea models are copied
// on every update, so form values need stable storage outside a model copy.
type briefAnswers struct {
	documentType string
	audience     string
	tone         []string
	includes     []string
}

func newModel() model {
	m := model{brief: &briefAnswers{}, preview: true, message: "Answer the brief, then shape the words."}
	m.form = m.newBriefForm()
	m.editor = textarea.New()
	m.editor.Placeholder = "Your guided draft appears here…"
	m.editor.Focus()
	m.input = textinput.New()
	return m
}

func (m *model) newBriefForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What are you writing?").
				Options(huh.NewOption("Proposal", "proposal"), huh.NewOption("Release note", "release-note"), huh.NewOption("Incident update", "incident-update"), huh.NewOption("Meeting note", "meeting-note")).
				Value(&m.brief.documentType),
			huh.NewSelect[string]().
				Title("Who is it for?").
				Options(huh.NewOption("Team", "team"), huh.NewOption("Leadership", "leadership"), huh.NewOption("Customers", "customers")).
				Value(&m.brief.audience),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Choose a tone").
				Description("Space selects options; Enter continues.").
				Options(huh.NewOption("Concise", "concise"), huh.NewOption("Warm", "warm"), huh.NewOption("Technical", "technical"), huh.NewOption("Urgent", "urgent")).
				Value(&m.brief.tone),
			huh.NewMultiSelect[string]().
				Title("What should it include?").
				Description("Space selects options; Enter creates the draft.").
				Options(huh.NewOption("Summary", "summary"), huh.NewOption("Decisions", "decisions"), huh.NewOption("Action items", "action-items"), huh.NewOption("Risks", "risks")).
				Value(&m.brief.includes),
		),
	).WithTheme(huh.ThemeCharm()).WithShowHelp(false)
}

func (m model) Init() tea.Cmd { return m.form.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = size.Width, size.Height
		m.resize()
		m.form = m.form.WithWidth(max(40, size.Width-8))
		return m, nil
	}
	if m.screen == briefScreen {
		updatedForm, cmd := m.form.Update(msg)
		m.form = updatedForm.(*huh.Form)
		if m.form.State == huh.StateCompleted {
			m.screen = editorScreen
			m.editor.SetValue(m.template())
			m.editor.Focus()
			m.resize()
			return m, m.editor.Focus()
		}
		return m, cmd
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		if m.prompt != "" {
			return m.updatePrompt(key)
		}
		if m.menuOpen {
			return m.updateMenu(key)
		}
		switch key.String() {
		case "f10", "alt+m":
			m.menuOpen, m.itemIndex = true, 0
			m.resize()
			return m, nil
		case "alt+f", "alt+v", "alt+r", "alt+h":
			m.menuIndex = map[string]int{"alt+f": 0, "alt+v": 1, "alt+r": 2, "alt+h": 3}[key.String()]
			m.menuOpen, m.itemIndex = true, 0
			m.resize()
			return m, nil
		case "ctrl+p":
			m.preview = !m.preview
			m.resize()
			return m, nil
		case "ctrl+s":
			if m.path == "" {
				return m.ask("Save path", "guided-brief.md")
			}
			return m.save()
		case "ctrl+o":
			return m.ask("Open path", "")
		case "ctrl+g":
			m.screen = briefScreen
			m.form = m.newBriefForm()
			return m, m.form.Init()
		case "ctrl+b":
			m.editor.InsertString("**bold text**")
		case "ctrl+i":
			m.editor.InsertString("*italic text*")
		case "ctrl+u":
			m.editor.InsertString("<u>underlined text</u>")
		case "ctrl+h":
			m.editor.InsertString("## Heading\n")
		case "ctrl+l":
			m.editor.InsertString("- list item\n")
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
		m.menuIndex = (m.menuIndex + len(guidedMenus) - 1) % len(guidedMenus)
		m.itemIndex = 0
	case "right":
		m.menuIndex = (m.menuIndex + 1) % len(guidedMenus)
		m.itemIndex = 0
	case "up":
		items := guidedMenus[m.menuIndex].items
		m.itemIndex = (m.itemIndex + len(items) - 1) % len(items)
	case "down":
		items := guidedMenus[m.menuIndex].items
		m.itemIndex = (m.itemIndex + 1) % len(items)
	case "enter":
		return m.executeMenu(guidedMenus[m.menuIndex].items[m.itemIndex])
	}
	return m, nil
}

func (m model) executeMenu(item menuItem) (tea.Model, tea.Cmd) {
	m.menuOpen = false
	m.resize()
	switch item.action {
	case "brief":
		m.screen, m.form = briefScreen, m.newBriefForm()
		return m, m.form.Init()
	case "open":
		return m.ask("Open path", "")
	case "save":
		if m.path == "" {
			return m.ask("Save path", "guided-brief.md")
		}
		return m.save()
	case "save-as":
		return m.ask("Save path", m.path)
	case "preview":
		m.preview = !m.preview
		m.resize()
		m.message = "Preview toggled"
	case "heading":
		m.editor.InsertString("## Heading\n")
		m.message = "Heading inserted"
	case "bold", "italic", "underline":
		m.editor.InsertString(editor.InsertFormatting(item.action))
		m.message = strings.Title(item.action) + " inserted"
	case "bullet":
		m.editor.InsertString("- list item\n")
		m.message = "Bullet inserted"
	case "controls":
		m.message = "Alt+M opens menus · arrows select · Enter runs"
	case "about":
		m.message = "Guided Brief — purpose before tools"
	}
	return m, nil
}

func (m *model) resize() {
	width := max(30, m.width-8)
	if m.preview {
		width = max(28, (m.width-10)/2)
	}
	m.editor.SetWidth(width)
	menuRows := 0
	if m.menuOpen {
		menuRows = 7
	}
	m.editor.SetHeight(max(8, m.height-9-menuRows))
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
		value := m.input.Value()
		prompt := m.prompt
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
			m.path, m.message = value, "Opened "+filepath.Base(value)
			m.editor.SetValue(contents)
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

func (m model) template() string {
	name := strings.ReplaceAll(m.brief.documentType, "-", " ")
	if name == "" {
		name = "brief"
	}
	audience := m.brief.audience
	if audience == "" {
		audience = "your audience"
	}
	var sections []string
	for _, section := range m.brief.includes {
		sections = append(sections, "## "+strings.Title(strings.ReplaceAll(section, "-", " "))+"\n")
	}
	if len(sections) == 0 {
		sections = []string{"## Summary\n"}
	}
	tone := strings.Join(m.brief.tone, ", ")
	if tone == "" {
		tone = "neutral"
	}
	return "# " + strings.Title(name) + "\n\n" + "_For " + audience + ". Tone: " + tone + "._\n\n" + strings.Join(sections, "\n")
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.screen == briefScreen {
		return lipgloss.NewStyle().Padding(1, 3).Render("GUIDED BRIEF\nPurpose before tools\n\n" + m.form.View())
	}
	if m.prompt != "" {
		return "\n  " + m.prompt + ": " + m.input.View() + "\n"
	}
	left := panel("DRAFT · guided by purpose", m.editor.View(), max(30, m.editor.Width()))
	content := left
	if m.preview {
		content = lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", panel("READING VIEW · Glamour", m.previewView(), max(30, m.editor.Width())))
	}
	path := "untitled.md"
	if m.path != "" {
		path = filepath.Base(m.path)
	}
	menu := m.menuBar()
	if m.menuOpen {
		menu += "\n" + m.menuPopup()
	}
	footer := fmt.Sprintf("%s  ·  %s\nalt+m menus • ctrl+g revise brief • ctrl+s save • ctrl+o open • ctrl+p preview • ctrl+b/i/u/h/l format • ctrl+c quit", path, m.message)
	return "\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render("GUIDED BRIEF  /  PURPOSE → STRUCTURE → WORDS") + "\n" + menu + "\n\n" + content + "\n\n" + footer + "\n"
}

func (m model) menuBar() string {
	parts := make([]string, len(guidedMenus))
	for i, item := range guidedMenus {
		style := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("86"))
		if m.menuOpen && i == m.menuIndex {
			style = style.Reverse(true)
		}
		parts[i] = style.Render(item.name)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
func (m model) menuPopup() string {
	items := guidedMenus[m.menuIndex].items
	rows := make([]string, len(items))
	for i, item := range items {
		prefix := "  "
		if i == m.itemIndex {
			prefix = "› "
		}
		rows[i] = prefix + item.label
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("86")).Padding(0, 1).Render(strings.Join(rows, "\n"))
}

func (m model) previewView() string {
	previewStyle := styles.DarkStyleConfig
	// The source pane owns Markdown markers. The reading view must not repeat
	// them as decoration, especially for generated headings.
	previewStyle.H1.Prefix = ""
	previewStyle.H2.Prefix = ""
	previewStyle.H3.Prefix = ""
	previewStyle.H4.Prefix = ""
	previewStyle.H5.Prefix = ""
	previewStyle.H6.Prefix = ""
	r, err := glamour.NewTermRenderer(glamour.WithStyles(previewStyle), glamour.WithWordWrap(max(26, m.editor.Width()-4)), glamour.WithPreservedNewLines())
	if err != nil {
		return err.Error()
	}
	out, err := r.Render(m.editor.Value() + "\n")
	if err != nil {
		return err.Error()
	}
	return out
}

func panel(title, body string, width int) string {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("86")).Padding(0, 1).Width(width).Render(lipgloss.NewStyle().Bold(true).Render(title) + "\n" + body)
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
