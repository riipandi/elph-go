package renderer

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/riipandi/elph/internal/constants"
	"github.com/riipandi/elph/pkg/tools/todolist"
)

var (
	todoPanelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(constants.DimText).
			Padding(0, 1)
	todoPanelTitleStyle = lipgloss.NewStyle().Foreground(constants.BrightText).Bold(true)
	todoDoneStyle       = lipgloss.NewStyle().Foreground(constants.Green)
	todoActiveStyle     = lipgloss.NewStyle().Foreground(constants.Yellow).Bold(true)
	todoPendingStyle    = lipgloss.NewStyle().Foreground(constants.DimText)
)

func (m Model) showsTodoPanel() bool {
	return todolist.HasActive(m.session.Todos())
}

func formatTodosCompletedMessage(todos []todolist.Todo) string {
	if len(todos) == 0 {
		return "All tasks completed."
	}
	var b strings.Builder
	b.WriteString("All tasks completed.")
	for _, item := range todos {
		b.WriteString("\n✓ ")
		b.WriteString(item.Title)
	}
	return b.String()
}

func (m Model) addTodoCompletionMessage(text string) Model {
	m.messages = append(m.messages, message{text: text, kind: constants.MessageSystem})
	m.session.AppendLog("system", text)
	m.layout.ContentDirty = true
	return m
}

func (m Model) todoPanelHeight() int {
	if !m.showsTodoPanel() {
		return 0
	}
	return lipgloss.Height(m.todoPanelView())
}

func (m Model) todoPanelView() string {
	if !m.showsTodoPanel() {
		return ""
	}

	boxW := borderedChromeWidth(m.chromeOuterWidth())
	innerW := inputContentWidth(m.chromeOuterWidth())

	var lines []string
	title := "Tasks"
	if m.agent.TodoListUpdating {
		frame := spinnerFrames[m.agent.SpinnerFrame%len(spinnerFrames)]
		title = frame + " " + title
	}
	lines = append(lines, todoPanelTitleStyle.Render(title))

	for _, item := range m.session.Todos() {
		lines = append(lines, m.todoPanelLine(item, innerW))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return todoPanelBorder.Width(boxW).Render(body)
}

func (m Model) todoPanelLine(item todolist.Todo, maxWidth int) string {
	marker, textStyle := m.todoPanelMarker(item.Status)
	line := marker + " " + item.Title
	if maxWidth > 0 {
		line = ansi.Truncate(line, maxWidth, "…")
	}
	return textStyle.Render(line)
}

func (m Model) todoPanelMarker(status todolist.Status) (string, lipgloss.Style) {
	switch status {
	case todolist.StatusDone:
		return "✓", todoDoneStyle
	case todolist.StatusInProgress:
		if m.agent.Busy || m.agent.TodoListUpdating {
			frame := spinnerFrames[m.agent.SpinnerFrame%len(spinnerFrames)]
			return frame, todoActiveStyle
		}
		return "◐", todoActiveStyle
	default:
		return "○", todoPendingStyle
	}
}
