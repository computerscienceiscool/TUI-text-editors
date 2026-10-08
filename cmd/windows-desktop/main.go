// Command windows-desktop is a 1990s desktop-inspired Markdown editor demo.
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
	"github.com/muesli/termenv"
)

const desktopTeal = "6"
const windowGray = "7"
const titleNavy = "4"

type menu struct {
	name  string
	items []menuItem
}

type menuItem struct {
	label  string
	action string
}

var desktopMenus = []menu{
	{"File", []menuItem{{"New", "new"}, {"Open…", "open"}, {"Save", "save"}, {"Save As…", "save-as"}, {"Quit", "quit"}}},
	{"Edit", []menuItem{{"Cursor up", "cursor-up"}, {"Cursor down", "cursor-down"}, {"Start of line", "line-start"}, {"End of line", "line-end"}}},
	{"View", []menuItem{{"Split preview", "preview"}, {"Scroll to top", "scroll-top"}, {"Scroll to bottom", "scroll-bottom"}}},
	{"Format", []menuItem{{"Heading", "heading"}, {"Bold", "bold"}, {"Italic", "italic"}, {"Underline", "underline"}, {"Bullet list", "bullet"}}},
	{"Help", []menuItem{{"Keyboard controls", "controls"}, {"About", "about"}}},
}

type model struct {
	editor                textarea.Model
	input                 textinput.Model
	width, height         int
	path, message, prompt string
	menuIndex, itemIndex  int
	menuOpen, preview     bool
}

func newModel() model {
	ed := textarea.New()
	ed.Placeholder = "Type a memo…"
	ed.ShowLineNumbers = false
	ed.Focus()
	return model{editor: ed, input: textinput.New(), preview: true, message: "Ready"}
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
		case "f10", "ctrl+g", "alt+m":
			m.menuOpen = true
			m.itemIndex = 0
			m.resize()
			return m, nil
		case "alt+f", "alt+e", "alt+v", "alt+r", "alt+h":
			m.menuIndex = map[string]int{"alt+f": 0, "alt+e": 1, "alt+v": 2, "alt+r": 3, "alt+h": 4}[key.String()]
			m.itemIndex, m.menuOpen = 0, true
			m.resize()
			return m, nil
		case "f2":
			m.editor.SetValue("")
			m.path = ""
			m.message = "New document"
			return m, nil
		case "f3":
			if m.path == "" {
				return m.ask("Save as", "memo.md")
			}
			return m.save()
		case "f4":
			return m.ask("Open", "")
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
		case "pgdown":
			m.moveCursor(m.editor.Height() - 2)
			m.message = "Scrolled down"
			return m, nil
		case "pgup":
			m.moveCursor(2 - m.editor.Height())
			m.message = "Scrolled up"
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m model) updateMenu(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc", "f10", "ctrl+g", "alt+m":
		m.menuOpen = false
		m.resize()
		return m, nil
	case "left":
		m.menuIndex = (m.menuIndex + len(desktopMenus) - 1) % len(desktopMenus)
		m.itemIndex = 0
	case "right":
		m.menuIndex = (m.menuIndex + 1) % len(desktopMenus)
		m.itemIndex = 0
	case "up":
		items := desktopMenus[m.menuIndex].items
		m.itemIndex = (m.itemIndex + len(items) - 1) % len(items)
	case "down":
		items := desktopMenus[m.menuIndex].items
		m.itemIndex = (m.itemIndex + 1) % len(items)
	case "enter":
		return m.execute(desktopMenus[m.menuIndex].items[m.itemIndex])
	}
	return m, nil
}

func (m model) execute(item menuItem) (tea.Model, tea.Cmd) {
	m.menuOpen = false
	m.resize()
	switch item.action {
	case "new":
		m.editor.SetValue("")
		m.path, m.message = "", "New document"
	case "open":
		return m.ask("Open", "")
	case "save":
		if m.path == "" {
			return m.ask("Save as", "memo.md")
		}
		return m.save()
	case "save-as":
		return m.ask("Save as", m.path)
	case "quit":
		return m, tea.Quit
	case "cursor-up":
		m.editor.CursorUp()
		m.message = "Cursor up"
	case "cursor-down":
		m.editor.CursorDown()
		m.message = "Cursor down"
	case "line-start":
		m.editor.CursorStart()
		m.message = "Start of line"
	case "line-end":
		m.editor.CursorEnd()
		m.message = "End of line"
	case "preview":
		m.preview = !m.preview
		m.resize()
		m.message = toggle("Preview", m.preview)
	case "scroll-top":
		m.moveCursor(-m.editor.Line())
		m.message = "Scrolled to top"
	case "scroll-bottom":
		m.moveCursor(len(strings.Split(m.editor.Value(), "\n")))
		m.message = "Scrolled to bottom"
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
		m.message = "F10 menus · arrows select · Enter runs · Esc closes"
	case "about":
		m.message = "Windows Desktop — same editor, different visual language"
	}
	return m, nil
}

func (m *model) moveCursor(lines int) {
	for range abs(lines) {
		if lines > 0 {
			m.editor.CursorDown()
		} else {
			m.editor.CursorUp()
		}
	}
}
func (m *model) resize() {
	width := max(30, m.width-10)
	if m.preview {
		width = max(28, (m.width-10)/2)
	}
	m.editor.SetWidth(width)
	// The desktop has a title bar, menu, raised toolbar, panels, and status
	// strip. Leave enough rows for that chrome so it never scrolls away.
	menuRows := 0
	if m.menuOpen {
		menuRows = 7
	}
	m.editor.SetHeight(max(8, m.height-15-menuRows))
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
		value, kind := m.input.Value(), m.prompt
		m.prompt = ""
		m.editor.Focus()
		if !editor.SupportedDocument(value) {
			m.message = "Use .md, .markdown, or .txt"
			return m, nil
		}
		if kind == "Open" {
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
		return desktopStyle().Render(window("Open / Save", "  "+m.prompt+": "+m.input.View()+"\n\n  Enter = confirm · Esc = cancel", max(42, m.width-12), 3))
	}
	name := "Untitled - Terminal Write"
	if m.path != "" {
		name = filepath.Base(m.path) + " - Terminal Write"
	}
	menu := m.menuBar()
	if m.menuOpen {
		menu += "\n" + m.menuPopup()
	}
	toolbar := lipgloss.JoinHorizontal(lipgloss.Top,
		button("F2 New"), button("F3 Save"), button("F4 Open"),
		button("F5 Head"), button("F6 List"), button("F7 Preview"),
	)
	draft := window("Document", m.documentView(), max(30, m.editor.Width()), m.editor.Height())
	content := draft
	if m.preview {
		content = lipgloss.JoinHorizontal(lipgloss.Top, draft, "  ", window("Preview", m.previewView(), max(30, m.editor.Width()), m.editor.Height()))
	}
	status := lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color(windowGray)).Render(" " + m.message + "  |  Ctrl+G / Alt+M menus  |  PgUp/PgDn scroll  |  Ctrl+C Exit ")
	body := titleBar(name, max(50, m.width-4)) + "\n" + menu + "\n" + toolbar + "\n\n" + content + "\n\n" + status
	return desktopStyle().Render(body)
}
func (m model) menuBar() string {
	parts := make([]string, len(desktopMenus))
	for i, item := range desktopMenus {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color(windowGray)).Padding(0, 1)
		if m.menuOpen && i == m.menuIndex {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color(titleNavy)).Padding(0, 1).Bold(true)
		}
		parts[i] = style.Render(item.name)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
func (m model) menuPopup() string {
	items := desktopMenus[m.menuIndex].items
	rows := make([]string, len(items))
	for i, item := range items {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color(windowGray)).Width(22)
		prefix := "  "
		if i == m.itemIndex {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color(titleNavy)).Width(22)
			prefix = "► "
		}
		rows[i] = style.Render(prefix + item.label)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color(windowGray)).Border(lipgloss.DoubleBorder()).Padding(0, 1).Render(strings.Join(rows, "\n"))
}
func desktopStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(lipgloss.Color(desktopTeal)).Padding(1, 2)
}
func titleBar(title string, width int) string {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color(titleNavy)).Width(width).Render(" " + title)
}
func button(label string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color(windowGray)).Border(lipgloss.NormalBorder()).Padding(0, 1).Render(label)
}
func window(title, body string, width, bodyHeight int) string {
	// A shared body height keeps split panes aligned. Without it, the shorter
	// preview left unpainted terminal background beneath its border.
	body = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color(windowGray)).Width(max(1, width-2)).Height(bodyHeight).Render(body)
	return lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color(windowGray)).Border(lipgloss.DoubleBorder()).Padding(0, 1).Width(width).Render(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color(titleNavy)).Render(" "+title+" ") + "\n" + body)
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
		mark := " "
		if line == active {
			mark = "►"
		}
		fmt.Fprintf(&out, "%s %3d ", mark, line+1)
		if line >= len(lines) {
			out.WriteString(m.scrollTrack(row, start, len(lines), height))
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
		out.WriteString(m.scrollTrack(row, start, len(lines), height))
	}
	return out.String()
}

func (m model) scrollTrack(row, start, total, height int) string {
	if total <= height {
		return "  "
	}
	thumb := (start * height) / max(1, total)
	thumbHeight := max(1, (height*height)/total)
	if row >= thumb && row < min(height, thumb+thumbHeight) {
		return " █"
	}
	return " │"
}

const underlineStart = "WINDOWSUNDERLINESTART"
const underlineEnd = "WINDOWSUNDERLINEEND"

var underlineTag = regexp.MustCompile(`(?s)<u>(.*?)</u>`)
var previewBackground = regexp.MustCompile(`\x1b\[(?:4[0-9]|48;5;[0-9]+)m`)

func (m model) previewView() string {
	source := underlineTag.ReplaceAllString(m.editor.Value(), underlineStart+"${1}"+underlineEnd)
	if !strings.HasSuffix(source, "\n") {
		source += "\n"
	}
	// The gray desktop window owns the background. Glamour contributes only
	// Markdown typography and color inside it.
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
	return keepPreviewSurface(applyUnderlines(out))
}

// Glamour resets styles after each span. In a terminal that reset restores the
// terminal's default background (often black), not this editor's gray paper.
// Keep Markdown foreground styling while restoring the preview surface after
// every reset and disallowing Glamour-owned background fills.
func keepPreviewSurface(rendered string) string {
	rendered = previewBackground.ReplaceAllString(rendered, "")
	rendered = strings.ReplaceAll(rendered, "\x1b[0m", "\x1b[0;30;47m")
	return "\x1b[30;47m" + rendered + "\x1b[0m"
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
func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
func main() {
	lipgloss.SetColorProfile(termenv.ANSI256)
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
