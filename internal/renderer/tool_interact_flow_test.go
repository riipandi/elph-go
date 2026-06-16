package renderer

import (
	"testing"

	"github.com/riipandi/elph/internal/uiconst"
	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/stretchr/testify/require"
)

func TestToolInteractOfferShowsAskUserDialog(t *testing.T) {
	m := testInputModel(t)
	m.height = 24
	m.width = 100
	m.ready = true
	m = m.beginAgentTurn()
	bridge := newToolInteractBridge()
	m.agent.ToolInteractBridge = bridge

	offer := toolInteractOffer{
		Req: agent.ToolInteractRequest{
			Kind: agent.ToolInteractAskUser,
			Args: map[string]any{"question": "Pick one"},
		},
		RespCh: make(chan agent.ToolInteractResponse, 1),
	}
	bridge.Inbox <- offer

	updated, _ := m.Update(toolInteractOfferMsg{Offer: offer})
	m = updated.(Model)

	require.True(t, m.toolInteractDialogActive())
	dialog := stripANSI(m.toolInteractChromeView())
	require.Contains(t, dialog, "Question")
	require.Contains(t, dialog, "Pick one")
	require.NotContains(t, dialog, "↑ up")
	input := stripANSI(m.inputChromeView())
	require.Contains(t, input, ">")
	require.NotContains(t, input, "Pick one")
	content := stripANSI(m.contentView())
	require.NotContains(t, content, "Pick one")
	require.False(t, m.input.Focused())
}

func TestToolInteractOfferMsgReturnsWithoutFallingThrough(t *testing.T) {
	m := testInputModel(t)
	m.height = 24
	m.width = 100
	m.ready = true
	m.messages = []message{{text: "hi", kind: uiconst.MessageUser}}

	offer := toolInteractOffer{
		Req: agent.ToolInteractRequest{
			Kind: agent.ToolInteractAskUser,
			Args: map[string]any{"question": "Q"},
		},
		RespCh: make(chan agent.ToolInteractResponse, 1),
	}

	updated, cmd := m.Update(toolInteractOfferMsg{Offer: offer})
	m = updated.(Model)
	require.True(t, m.toolInteractDialogActive())
	require.NotNil(t, cmd)
}
