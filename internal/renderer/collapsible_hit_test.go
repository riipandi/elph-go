package renderer

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/riipandi/elph/internal/constants"
	"github.com/stretchr/testify/require"
)

func TestMouseClickTogglesDetailBlockViaHintOnly(t *testing.T) {
	m := testModelWithLayout(t)
	m.height = 40
	m.messages = []message{
		{kind: constants.MessageDetail, detailLabel: "First", text: "alpha\nbeta"},
		{kind: constants.MessageDetail, detailLabel: "Second", text: "gamma\ndelta"},
	}
	m = m.syncLayout(false)

	headerY, ok := m.collapsibleHeaderViewportY(0)
	require.True(t, ok)
	updated, _ := m.Update(mouseClick(2, headerY, tea.MouseLeft, 0))
	m = updated.(Model)
	require.False(t, m.messages[0].detailExpanded)
	require.True(t, m.selectingText)

	m.selectingText = false
	m.mouseEnabled = true

	footerY, ok := m.collapsibleFooterViewportY(0)
	require.True(t, ok)
	updated, cmds := m.Update(mouseClick(2, footerY, tea.MouseLeft, 0))
	m = updated.(Model)
	require.Empty(t, cmds)
	require.True(t, m.messages[0].detailExpanded)
	require.False(t, m.messages[1].detailExpanded)
	require.True(t, m.mouseEnabled)
	require.False(t, m.selectingText)
}

func TestMouseClickOnThinkingHeaderStillToggles(t *testing.T) {
	m := testModelWithLayout(t)
	m.messages = []message{
		{kind: constants.MessageThinking, detailLabel: "Thinking", text: "reasoning"},
		{kind: constants.MessageAI, text: "answer"},
	}
	m = m.syncLayout(false)

	y, ok := m.collapsibleHeaderViewportY(0)
	require.True(t, ok)

	updated, _ := m.Update(mouseClick(2, y, tea.MouseLeft, 0))
	m = updated.(Model)
	require.True(t, m.messages[0].detailExpanded)
}

func TestCtrlOTogglesNewestCollapsibleBlock(t *testing.T) {
	m := testModelWithLayout(t)
	m.height = 40
	m.messages = []message{
		{kind: constants.MessageDetail, detailLabel: "First", text: "alpha"},
		{kind: constants.MessageDetail, detailLabel: "Second", text: "beta"},
	}
	m = m.syncLayout(false)

	footerY, ok := m.collapsibleFooterViewportY(0)
	require.True(t, ok)
	updated, _ := m.Update(mouseClick(2, footerY, tea.MouseLeft, 0))
	m = updated.(Model)
	require.True(t, m.messages[0].detailExpanded)

	updated, _ = m.Update(keyCtrl('o'))
	m = updated.(Model)
	require.True(t, m.messages[0].detailExpanded)
	require.True(t, m.messages[1].detailExpanded)
}
