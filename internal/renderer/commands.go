package renderer

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/command"
	"github.com/riipandi/elph/internal/constants"
)

var (
	cmdPaletteSelected = lipgloss.NewStyle().Foreground(constants.Highlight).Bold(true)
	cmdPaletteName     = lipgloss.NewStyle().Foreground(constants.White)
)

func cmdPaletteBorder(mode constants.AgentMode) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(constants.ModeBorderColor(mode)).
		BorderBottom(false).
		Padding(0, 1)
}

func (m Model) commandPaletteActive() bool {
	return len(m.cmdSuggestions) > 0 && m.slashQueryActive()
}

func (m Model) slashQueryActive() bool {
	return strings.HasPrefix(strings.TrimLeft(m.input.Value(), " \t"), "/")
}

func (m Model) slashQuery() string {
	val := strings.TrimLeft(m.input.Value(), " \t")
	if !strings.HasPrefix(val, "/") {
		return ""
	}
	query := strings.TrimPrefix(val, "/")
	if idx := strings.Index(query, " "); idx >= 0 {
		query = query[:idx]
	}
	return strings.ToLower(strings.TrimSpace(query))
}

func (m Model) syncCommandSuggestions() Model {
	if m.busy || !m.input.Focused() {
		m.cmdSuggestions = nil
		m.cmdSuggestIndex = 0
		return m
	}

	if !m.slashQueryActive() {
		m.cmdSuggestions = nil
		m.cmdSuggestIndex = 0
		return m
	}

	m.cmdSuggestions = command.Suggest(m.slashQuery())
	if m.cmdSuggestIndex >= len(m.cmdSuggestions) {
		m.cmdSuggestIndex = 0
	}
	return m
}

func (m Model) applyCommandCompletion() Model {
	if len(m.cmdSuggestions) == 0 {
		return m
	}
	selected := m.cmdSuggestions[m.cmdSuggestIndex]
	m.input.SetValue(command.CompleteInput(selected))
	m = m.syncPromptPrefix()
	m = m.syncInputWidth()
	m = m.syncCommandSuggestions()
	return m
}

func (m Model) handleCommandPaletteKey(msg tea.KeyPressMsg) (Model, bool) {
	if !m.commandPaletteActive() {
		return m, false
	}

	switch msg.String() {
	case "tab", "right":
		return m.applyCommandCompletion(), true
	case "up":
		if len(m.cmdSuggestions) == 0 {
			return m, false
		}
		m.cmdSuggestIndex = (m.cmdSuggestIndex - 1 + len(m.cmdSuggestions)) % len(m.cmdSuggestions)
		return m, true
	case "down":
		if len(m.cmdSuggestions) == 0 {
			return m, false
		}
		m.cmdSuggestIndex = (m.cmdSuggestIndex + 1) % len(m.cmdSuggestions)
		return m, true
	}
	return m, false
}

func (m Model) commandPaletteView() string {
	if !m.commandPaletteActive() {
		return ""
	}

	nameColW := command.NameColumnWidth(m.cmdSuggestions, false)
	lines := make([]string, len(m.cmdSuggestions))
	for i, cmd := range m.cmdSuggestions {
		name, gap, summary := command.AlignedRow(cmd, nameColW, false)
		if i == m.cmdSuggestIndex {
			name = cmdPaletteSelected.Render(name)
		} else {
			name = cmdPaletteName.Render(name)
		}
		lines[i] = name + gap + dimStyle.Render(summary)
	}

	inner := strings.Join(lines, "\n")
	boxW := borderedChromeWidth(m.chromeOuterWidth())
	return cmdPaletteBorder(m.mode).Width(boxW).Render(inner)
}

func (m Model) commandPaletteHeight() int {
	if view := m.commandPaletteView(); view != "" {
		return lipgloss.Height(view)
	}
	return 0
}
