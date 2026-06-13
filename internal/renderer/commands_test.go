package renderer

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/command"
	"github.com/stretchr/testify/require"
)

func TestCommandPaletteAppearsForSlashInput(t *testing.T) {
	m := testInputModel(t)
	m.input.SetValue("/diagnostic:")

	m = m.syncCommandSuggestions()
	require.True(t, m.commandPaletteActive())
	require.NotEmpty(t, m.commandPaletteView())
	for _, cmd := range m.cmdSuggestions {
		require.Contains(t, cmd.Name, "diagnostic:")
	}
}

func TestCommandPaletteHiddenForNormalInput(t *testing.T) {
	m := testInputModel(t)
	m.input.SetValue("hello")

	m = m.syncCommandSuggestions()
	require.False(t, m.commandPaletteActive())
}

func TestCommandPaletteTwoColumnLayout(t *testing.T) {
	m := testInputModel(t)
	m.input.SetValue("/")
	m = m.syncCommandSuggestions()
	require.GreaterOrEqual(t, len(m.cmdSuggestions), 2)

	got := command.FormatList(m.cmdSuggestions)
	lines := strings.Split(got, "\n")
	require.GreaterOrEqual(t, len(lines), 2)

	firstSummary := strings.Index(lines[0], m.cmdSuggestions[0].Description)
	secondSummary := strings.Index(lines[1], m.cmdSuggestions[1].Description)
	require.Equal(t, firstSummary, secondSummary)

	view := stripANSI(m.commandPaletteView())
	require.Contains(t, view, "/"+m.cmdSuggestions[0].Name)
	require.Contains(t, view, m.cmdSuggestions[0].Description)
}

func TestTabCompletesSelectedCommand(t *testing.T) {
	m := testInputModel(t)
	m.input.SetValue("/help")
	m = m.syncCommandSuggestions()

	updated, consumed := m.handleCommandPaletteKey(keyTab())
	require.True(t, consumed)
	require.Equal(t, "/help", updated.input.Value())
}

func TestDownCyclesSuggestionSelection(t *testing.T) {
	m := testInputModel(t)
	m.input.SetValue("/")
	m = m.syncCommandSuggestions()
	require.GreaterOrEqual(t, len(m.cmdSuggestions), 2)

	updated, consumed := m.handleCommandPaletteKey(keyDown())
	require.True(t, consumed)
	require.Equal(t, 1, updated.cmdSuggestIndex)
}

func TestPaletteSitsFlushAboveInput(t *testing.T) {
	m := testInputModel(t)
	m.input.SetValue("/help")
	m = m.syncCommandSuggestions()

	chromeH := lipgloss.Height(m.inputChromeView())
	paletteH := lipgloss.Height(m.commandPaletteView())
	inputH := lipgloss.Height(m.inputBoxView(true))
	require.Equal(t, paletteH+inputH, chromeH)
}

func TestCompleteInputUsesCatalogName(t *testing.T) {
	cmd, ok := command.Get(command.DiagnosticListTools)
	require.True(t, ok)
	require.Equal(t, "/diagnostic:list-tools", command.CompleteInput(cmd))
}
