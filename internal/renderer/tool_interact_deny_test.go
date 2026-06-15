package renderer

import (
	"encoding/json"
	"testing"

	"charm.land/huh/v2"
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/stretchr/testify/require"
)

func TestDenyApprovalUpdatesToolDetailImmediately(t *testing.T) {
	m := testInputModel(t)
	m = m.beginAgentTurn()

	call := provider.ToolCall{
		ID:        "call_bash",
		Name:      "Bash",
		Arguments: json.RawMessage(`{"command":"ping 1.1.1.1"}`),
	}
	m = m.beginNativeToolCall(call)
	idx := m.agent.NativeToolMsgIDs["call_bash"]
	require.Equal(t, uiconst.DetailStatusRunning, m.messages[idx].detailStatus)

	respCh := make(chan agent.ToolInteractResponse, 1)
	m.toolInteractForm = newToolApprovalForm(agent.ToolInteractRequest{
		Kind: agent.ToolInteractApproval,
		Name: "Bash",
	}, 60)
	m.toolInteractPending = toolInteractOffer{
		Req: agent.ToolInteractRequest{
			Kind:     agent.ToolInteractApproval,
			Name:     "Bash",
			ToolCall: call,
		},
		RespCh: respCh,
	}

	m, _ = m.completeToolInteractWith(agent.ToolInteractResponse{Approved: false})
	require.False(t, (<-respCh).Approved)
	require.Equal(t, agent.ToolDeniedMessage, m.messages[idx].text)
	require.NotEqual(t, uiconst.DetailStatusRunning, m.messages[idx].detailStatus)
	require.Equal(t, agent.ActivityThinking, m.agent.Activity)
}

func TestDeniedApprovalNotPromptedAgainSameTurn(t *testing.T) {
	bridge := newToolInteractBridge()
	req := agent.ToolInteractRequest{
		Kind: agent.ToolInteractApproval,
		Name: "Bash",
		Args: map[string]any{"command": "ping 1.1.1.1"},
	}
	bridge.DeniedApprovals = map[string]struct{}{
		toolApprovalSignature(req): {},
	}

	resp, err := bridge.Interact(t.Context(), req)
	require.NoError(t, err)
	require.False(t, resp.Approved)
}

func TestCompleteToolInteractDenyBlocksRepeatPrompt(t *testing.T) {
	m := testInputModel(t)
	bridge := newToolInteractBridge()
	m.agent.ToolInteractBridge = bridge
	req := agent.ToolInteractRequest{
		Kind: agent.ToolInteractApproval,
		Name: "Bash",
		Args: map[string]any{"command": "ping 1.1.1.1"},
	}
	m.toolInteractPending = toolInteractOffer{Req: req, RespCh: make(chan agent.ToolInteractResponse, 1)}

	m, _ = m.completeToolInteractWith(agent.ToolInteractResponse{Approved: false})
	require.Contains(t, bridge.DeniedApprovals, toolApprovalSignature(req))

	resp, err := bridge.Interact(t.Context(), req)
	require.NoError(t, err)
	require.False(t, resp.Approved)
}

func TestRecordToolApprovalDenialCachesSignature(t *testing.T) {
	m := testInputModel(t)
	bridge := newToolInteractBridge()
	m.agent.ToolInteractBridge = bridge
	req := agent.ToolInteractRequest{
		Kind: agent.ToolInteractApproval,
		Name: "Bash",
		Args: map[string]any{"command": "rm -rf /"},
	}

	m = m.recordToolApprovalDenial(agent.ToolInteractResponse{Approved: false}, req)
	require.Contains(t, bridge.DeniedApprovals, toolApprovalSignature(req))
}

func TestApprovalFormAbortCountsAsCancel(t *testing.T) {
	form := newToolApprovalForm(agent.ToolInteractRequest{
		Kind: agent.ToolInteractApproval,
		Name: "Bash",
	}, 60)
	form.State = huh.StateAborted

	m := testInputModel(t)
	resp := m.approvalFormResponse(form)
	require.False(t, resp.Approved)
	require.True(t, resp.Cancelled)
}
