package renderer

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/stretchr/testify/require"
)

func TestSubmitDiagnosticSystemPromptUsesDetailBox(t *testing.T) {
	m := testInputModel(t)
	m.session.SystemPrompt = "You are an expert coding assistant.\n\n## Output\n- Be concise."
	m.input.SetValue("/diagnostic:system-prompt")

	updated, cmd := m.Update(keyEnter())
	m = updated.(Model)

	require.Nil(t, cmd)
	require.Len(t, m.messages, 2)
	require.Equal(t, uiconst.MessageUser, m.messages[0].kind)
	require.Equal(t, "/diagnostic:system-prompt", m.messages[0].text)
	require.Equal(t, uiconst.MessageDetail, m.messages[1].kind)
	require.Equal(t, "System prompt", m.messages[1].detailLabel)
	require.Contains(t, m.messages[1].text, "You are an expert coding assistant.")
	require.False(t, m.messages[1].detailExpanded)

	rendered := stripANSI(m.renderMessageAt(1))
	require.Contains(t, rendered, "System prompt")
	require.Contains(t, rendered, "ctrl+o to expand")
	require.NotContains(t, rendered, "## Output")
}

func TestDiagnosticSystemPromptDetailExpands(t *testing.T) {
	m := testInputModel(t)
	m.session.SystemPrompt = "You are an expert coding assistant.\n\n## Output\n- Be concise."
	m.input.SetValue("/diagnostic:system-prompt")

	updated, _ := m.Update(keyEnter())
	m = updated.(Model)

	collapsedH := lipgloss.Height(stripANSI(m.renderMessageAt(1)))
	m, _ = m.toggleLastDetailExpand()
	expanded := stripANSI(m.renderMessageAt(1))
	require.Contains(t, expanded, "## Output")
	require.Greater(t, lipgloss.Height(expanded), collapsedH)
}
