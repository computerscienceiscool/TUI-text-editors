// Command playground is an expressive, low-stakes text editor demo.
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

var moods = []string{"SPARK ✦", "FOCUS ◉", "PLAY ★"}
var prompts = []string{"Start with a surprising fact.", "What would make this useful tomorrow?", "Name the decision before explaining it.", "Write the smallest honest next step."}

type model struct {
	editor                textarea.Model
	input                 textinput.Model
	width, height         int
	path, message, prompt string
	preview               bool
	mood, promptIndex     int
}

func newModel() model {
	ed := textarea.New()
	ed.Placeholder = "Make a tiny mess. We can edit it."
	ed.ShowLineNumbers = false
	ed.Focus()
	return model{editor: ed, input: textinput.New(), preview: true, message: "A blank page is an invitation."}
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
		case "tab":
			m.mood = (m.mood + 1) % len(moods)
			m.message = "Mood: " + moods[m.mood]
			return m, nil
		case "f1":
			m.promptIndex = (m.promptIndex + 1) % len(prompts)
			m.message = prompts[m.promptIndex]
			return m, nil
		case "ctrl+n":
			m.editor.SetValue("")
			m.path = ""
			m.message = "Fresh canvas!"
			return m, nil
		case "ctrl+s":
			if m.path == "" {
				return m.ask("Save path", "playground.md")
			}
			return m.save()
		case "ctrl+o":
			return m.ask("Open path", "")
		case "f5":
			m.editor.InsertString("## A small idea\n")
			m.message = "Title sticker added"
			return m, nil
		case "f6":
			m.editor.InsertString("- tiny step\n")
			m.message = "Step sticker added"
			return m, nil
		case "f7":
			m.editor.InsertString("> a line worth keeping\n")
			m.message = "Quote sticker added"
			return m, nil
		case "f8":
			m.editor.InsertString("✨ ")
			m.message = "Sparkle added"
			return m, nil
		case "ctrl+p":
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
func (m *model) resize() {
	width := max(30, m.width-8)
	if m.preview {
		if m.width < 150 {
			// Keep the prompt panel above the editor on ordinary terminals rather
			// than squeezing three bordered panels into a broken row.
			width = max(28, (m.width-6)/2)
		} else {
			width = max(28, (m.width-34)/2)
		}
	}
	m.editor.SetWidth(width)
	m.editor.SetHeight(max(8, m.height-10))
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
		if kind == "Open path" {
			body, err := editor.ReadDocument(value)
			if err != nil {
				m.message = "Open failed: " + err.Error()
				return m, nil
			}
			m.editor.SetValue(body)
			m.path = value
			m.message = "Welcome back, " + filepath.Base(value)
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
	m.message = "Saved — nice work."
	return m, nil
}
func (m model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.prompt != "" {
		return "\n  " + m.prompt + ": " + m.input.View() + "\n"
	}
	head := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("99")).Padding(0, 1).Render("PLAYGROUND") + "  " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render(moods[m.mood])
	side := panel("WRITING SPARK", prompts[m.promptIndex]+"\n\n[Tab] mood\n[F1] next prompt\n[F5] title\n[F6] step\n[F7] quote\n[F8] sparkle", 25, "212")
	draft := panel("YOUR DRAFT", m.documentView(), max(30, m.editor.Width()), "141")
	content := lipgloss.JoinHorizontal(lipgloss.Top, side, " ", draft)
	if m.preview {
		reading := panel("HOW IT READS", m.previewView(), max(26, m.editor.Width()), "86")
		if m.width < 150 {
			content = lipgloss.JoinVertical(lipgloss.Left, side, lipgloss.JoinHorizontal(lipgloss.Top, draft, " ", reading))
		} else {
			content = lipgloss.JoinHorizontal(lipgloss.Top, content, " ", reading)
		}
	}
	name := "untitled.md"
	if m.path != "" {
		name = filepath.Base(m.path)
	}
	footer := fmt.Sprintf("%s · %s · Ln %d Col %d\n^N new  ^O open  ^S save  ^P preview  ^C quit", name, m.message, m.editor.Line()+1, m.editor.LineInfo().StartColumn+m.editor.LineInfo().ColumnOffset+1)
	return "\n" + head + "\n\n" + content + "\n\n" + footer + "\n"
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
			mark = "✦"
		}
		fmt.Fprintf(&out, "%s %3d ", mark, line+1)
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

const underlineStart = "PLAYGROUNDUNDERLINESTART"
const underlineEnd = "PLAYGROUNDUNDERLINEEND"

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
func panel(title, body string, width int, color string) string {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(color)).Padding(0, 1).Width(width).Render(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color)).Render(title) + "\n" + body)
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
