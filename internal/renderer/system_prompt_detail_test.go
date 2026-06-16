package renderer

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/stretchr/testify/require"
)

func TestContextCommandIncludesSystemPromptCollapsed(t *testing.T) {
	m := testInputModel(t)
	m.session.SystemPrompt = "You are an expert coding assistant.\n\n## Output\n- Be concise."
	m.contextWindow = 262144
	m.input.SetValue("/context")

	updated, cmd := m.Update(keyEnter())
	m = updated.(Model)

	require.Nil(t, cmd)
	// Expect: user msg, context detail, system prompt detail = 3 messages
	require.Len(t, m.messages, 3)
	require.Equal(t, uiconst.MessageUser, m.messages[0].kind)
	require.Equal(t, "/context", m.messages[0].text)
	require.Equal(t, uiconst.MessageDetail, m.messages[1].kind)
	require.Equal(t, "Context Usage", m.messages[1].detailLabel)
	require.True(t, m.messages[1].detailExpanded)
	require.Contains(t, m.messages[1].text, "tokens")
	// Third message is the system prompt, collapsed by default
	require.Equal(t, uiconst.MessageDetail, m.messages[2].kind)
	require.Equal(t, "System prompt", m.messages[2].detailLabel)
	require.False(t, m.messages[2].detailExpanded)

	rendered := stripANSI(m.renderMessageAt(2))
	require.Contains(t, rendered, "System prompt")
	require.Contains(t, rendered, "ctrl+o to expand")
	require.NotContains(t, rendered, "## Output")
}

func TestContextCommandSystemPromptExpands(t *testing.T) {
	m := testInputModel(t)
	m.session.SystemPrompt = "You are an expert coding assistant.\n\n## Output\n- Be concise."
	m.contextWindow = 262144
	m.input.SetValue("/context")

	updated, _ := m.Update(keyEnter())
	m = updated.(Model)

	// System prompt is the 3rd message (index 2)
	require.Equal(t, "System prompt", m.messages[2].detailLabel)
	collapsedH := lipgloss.Height(stripANSI(m.renderMessageAt(2)))
	m, _ = m.toggleDetailExpandAt(2)
	expanded := stripANSI(m.renderMessageAt(2))
	require.Contains(t, expanded, "## Output")
	require.Greater(t, lipgloss.Height(expanded), collapsedH)
}
