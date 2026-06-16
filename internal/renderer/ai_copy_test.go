package renderer

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/stretchr/testify/require"
)

func TestAICopyHintGapAfterContent(t *testing.T) {
	m := testModel()
	rendered := stripANSI(m.renderMessage(message{
		text: "Assistant reply body.",
		kind: uiconst.MessageAI,
	}))
	hintIdx := strings.Index(rendered, aiCopyHintText)
	bodyIdx := strings.Index(rendered, "Assistant reply body.")
	require.Greater(t, hintIdx, bodyIdx)
	require.Greater(t, strings.Count(rendered[bodyIdx:hintIdx], "\n"), 1)
}

func TestAIMessageShowsCopyHint(t *testing.T) {
	m := testModel()
	rendered := stripANSI(m.renderMessage(message{
		text: "Hello from the assistant.",
		kind: uiconst.MessageAI,
	}))
	require.Contains(t, rendered, aiCopyHintText)
}

func TestMarkdownPendingShowsCopyHint(t *testing.T) {
	m := testModel()
	m.messages = []message{{text: "**hello**", kind: uiconst.MessageAI, markdownPending: true}}
	rendered := stripANSI(m.renderMessageAt(0))
	require.Contains(t, rendered, aiCopyHintText)
}

func TestStreamingAIMessageHidesCopyHint(t *testing.T) {
	m := testModel()
	m.agent.Busy = true
	m.agent.ResponseMsgID = 0
	m.messages = []message{{text: "partial", kind: uiconst.MessageAI}}
	rendered := stripANSI(m.renderMessageAt(0))
	require.NotContains(t, rendered, aiCopyHintText)
}

func TestMouseClickOnAICopyHintCopiesMessage(t *testing.T) {
	m := testModelWithLayout(t)
	m.height = 40
	m.messages = []message{
		{kind: uiconst.MessageAI, text: "first answer"},
		{kind: uiconst.MessageAI, text: "second answer"},
	}
	m = m.syncLayout(false)
	m.content.GotoTop()

	y, ok := m.aiCopyFooterViewportY(0)
	require.True(t, ok)
	updated, _ := m.Update(mouseClick(2, y, tea.MouseLeft, 0))
	m = updated.(Model)
	require.GreaterOrEqual(t, len(m.messages), 3)
	require.Contains(t, stripANSI(m.messages[len(m.messages)-1].text), "Copied to clipboard")

	// Clicking body should start text selection, not copy.
	bodyY := y - 1
	require.Greater(t, bodyY, 0)
	updated, _ = m.Update(mouseClick(2, bodyY, tea.MouseLeft, 0))
	m = updated.(Model)
	require.True(t, m.selectingText)
}

func TestCtrlYCopiesLastAIMessage(t *testing.T) {
	m := testModel()
	m.messages = []message{
		{kind: uiconst.MessageUser, text: "question"},
		{kind: uiconst.MessageAI, text: "answer text"},
	}
	updated, _ := m.Update(keyCtrl('y'))
	m = updated.(Model)
	require.GreaterOrEqual(t, len(m.messages), 3)
	require.Contains(t, stripANSI(m.messages[len(m.messages)-1].text), "Copied to clipboard")
}
