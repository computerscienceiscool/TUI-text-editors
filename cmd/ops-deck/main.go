// Command ops-deck demonstrates a dense, operational editor layout.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"tui-text-editors/internal/editor"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	editor                textarea.Model
	input                 textinput.Model
	width, height         int
	path, message, prompt string
	preview, focusMode    bool
}

func newModel() model {
	ed := textarea.New()
	ed.Placeholder = "Operational update…"
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
		switch key.String() {
		case "ctrl+n":
			m.editor.SetValue("")
			m.path = ""
			m.message = "New document"
			return m, nil
		case "ctrl+o":
			return m.ask("Open path", "")
		case "ctrl+s":
			if m.path == "" {
				return m.ask("Save path", "ops-update.md")
			}
			return m.save()
		case "ctrl+p":
			m.preview = !m.preview
			m.resize()
			return m, nil
		case "ctrl+f":
			m.focusMode = !m.focusMode
			m.resize()
			return m, nil
		case "ctrl+h":
			m.editor.InsertString("## Heading\n")
			m.message = "Heading inserted"
			return m, nil
		case "ctrl+b":
			m.editor.InsertString("**bold text**")
			m.message = "Bold inserted"
			return m, nil
		case "ctrl+i":
			m.editor.InsertString("*italic text*")
			m.message = "Italic inserted"
			return m, nil
		case "ctrl+u":
			m.editor.InsertString("<u>underlined text</u>")
			m.message = "Underline inserted"
			return m, nil
		case "ctrl+l":
			m.editor.InsertString("- list item\n")
			m.message = "List inserted"
			return m, nil
		case "ctrl+t":
			m.editor.InsertString("- [ ] action\n")
			m.message = "Task inserted"
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m *model) resize() {
	width := max(30, m.width-10)
	if m.preview && !m.focusMode {
		width = max(28, (m.width-44)/2)
	}
	m.editor.SetWidth(width)
	m.editor.SetHeight(max(8, m.height-9))
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
		return "\n  " + m.prompt + ": " + m.input.View() + "\n"
	}
	head := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24")).Padding(0, 1).Render("OPS DECK") + "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("DENSE MODE · SCAN / EDIT / VERIFY")
	if m.focusMode {
		return "\n" + head + "\n\n" + panel("LIVE DRAFT", m.documentView(), max(50, m.editor.Width()), "39") + "\n\n" + m.footer() + "\n"
	}
	outline := panel("OUTLINE", m.outlineView(), 25, "214")
	draft := panel("LIVE DRAFT", m.documentView(), max(30, m.editor.Width()), "39")
	content := lipgloss.JoinHorizontal(lipgloss.Top, outline, " ", draft)
	if m.preview {
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, " ", panel("READING", m.previewView(), max(26, m.editor.Width()), "86"))
	}
	return "\n" + head + "\n\n" + content + "\n\n" + m.footer() + "\n"
}
func (m model) footer() string {
	name := "untitled.md"
	if m.path != "" {
		name = filepath.Base(m.path)
	}
	li := m.editor.LineInfo()
	metrics := fmt.Sprintf("%d words · %d tasks · %d headings", wordCount(m.editor.Value()), strings.Count(m.editor.Value(), "- ["), len(headings(m.editor.Value())))
	return fmt.Sprintf("%s · %s · %s · Ln %d Col %d\n^N new  ^O open  ^S save  ^H heading  ^B bold  ^I italic  ^U underline  ^L list  ^T task  ^P preview  ^F focus  ^C quit", name, m.message, metrics, m.editor.Line()+1, li.StartColumn+li.ColumnOffset+1)
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
			mark = "▶"
		}
		fmt.Fprintf(&out, "%s%03d ", mark, line+1)
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
func headings(source string) []string {
	var out []string
	for index, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimLeft(line, " ")
		level := len(trimmed) - len(strings.TrimLeft(trimmed, "#"))
		if level > 0 && level <= 6 && len(trimmed) > level && trimmed[level] == ' ' {
			out = append(out, fmt.Sprintf("%03d %s%s", index+1, strings.Repeat("· ", level-1), strings.TrimSpace(trimmed[level:])))
		}
	}
	return out
}
func (m model) outlineView() string {
	hs := headings(m.editor.Value())
	if len(hs) == 0 {
		return "No headings yet\n\n^H adds one."
	}
	return strings.Join(hs, "\n")
}

const underlineStart = "OPSDECKUNDERLINESTART"
const underlineEnd = "OPSDECKUNDERLINEEND"

var underlineTag = regexp.MustCompile(`(?s)<u>(.*?)</u>`)

func (m model) previewView() string {
	source := underlineTag.ReplaceAllString(m.editor.Value(), underlineStart+"${1}"+underlineEnd)
	if !strings.HasSuffix(source, "\n") {
		source += "\n"
	}
	style := styles.DarkStyleConfig
	style.H1.Prefix, style.H2.Prefix, style.H3.Prefix = "", "", ""
	style.H4.Prefix, style.H5.Prefix, style.H6.Prefix = "", "", ""
	r, err := glamour.NewTermRenderer(glamour.WithStyles(style), glamour.WithWordWrap(max(24, m.editor.Width()-4)), glamour.WithPreservedNewLines())
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
func wordCount(text string) int {
	return len(strings.FieldsFunc(text, func(r rune) bool { return unicode.IsSpace(r) }))
}
func panel(title, body string, width int, color string) string {
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(color)).Padding(0, 1).Width(width).Render(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color)).Render(title) + "\n" + body)
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
