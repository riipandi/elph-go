package renderer

import (
	"encoding/json"
	"testing"

	"github.com/riipandi/elph/internal/constants"
	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/stretchr/testify/require"
)

func TestBashApprovalDrainsToolCallDone(t *testing.T) {
	m := testInputModel(t)
	m.height = 24
	m.width = 100
	m.ready = true
	m = m.beginAgentTurn()

	events := make(chan agent.Event, 4)
	m.agent.Events = events
	m.agent.ToolInteractBridge = newToolInteractBridge()

	call := provider.ToolCall{
		ID:        "call_ping",
		Name:      "Bash",
		Arguments: json.RawMessage(`{"command":"ping 1.1.1.1"}`),
	}

	m, startCmd := m.handleAgentEvent(agentEventMsg{event: agent.ToolCallStartEvent(call)})
	require.NotNil(t, startCmd)

	idx := m.agent.NativeToolMsgIDs["call_ping"]
	require.Contains(t, m.messages[idx].text, "(running...)")
	require.Equal(t, constants.DetailStatusRunning, m.messages[idx].detailStatus)

	respCh := make(chan agent.ToolInteractResponse, 1)
	offer := toolInteractOffer{
		Req: agent.ToolInteractRequest{
			Kind: agent.ToolInteractApproval,
			Name: "Bash",
			Args: map[string]any{"command": "ping 1.1.1.1"},
		},
		RespCh: respCh,
	}
	updated, _ := m.Update(toolInteractOfferMsg{offer: offer})
	m = updated.(Model)

	m, approveCmd := m.completeToolInteractWith(agent.ToolInteractResponse{Approved: true})
	require.NotNil(t, approveCmd)
	require.True(t, (<-respCh).Approved)

	done := agent.ToolCallDoneEvent(call, agent.ToolRunResult{Output: "PING 1.1.1.1"})
	events <- done

	m, doneCmd := m.handleAgentEvent(agentEventMsg{event: done})
	require.NotNil(t, doneCmd)
	require.Equal(t, "PING 1.1.1.1", m.messages[idx].text)
	require.NotEqual(t, constants.DetailStatusRunning, m.messages[idx].detailStatus)
}
