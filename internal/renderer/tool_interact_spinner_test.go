package renderer

import (
	"testing"

	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/stretchr/testify/require"
)

func TestSpinnerTicksDuringToolInteractDialog(t *testing.T) {
	m := testInputModel(t)
	m = m.beginAgentTurn()
	m.agent.SpinnerFrame = 0
	m.toolInteractForm = newToolApprovalForm(agent.ToolInteractRequest{
		Kind: agent.ToolInteractApproval,
		Name: "Bash",
	}, 60)
	m.toolInteractPending = toolInteractOffer{
		Req: agent.ToolInteractRequest{Kind: agent.ToolInteractApproval, Name: "Bash"},
	}

	updated, cmd := m.Update(spinnerTickMsg{})
	m = updated.(Model)

	require.Equal(t, 1, m.agent.SpinnerFrame)
	require.NotNil(t, cmd)
}

func TestSpinnerRestartsAfterToolInteractApproval(t *testing.T) {
	m := testInputModel(t)
	m = m.beginAgentTurn()
	m.agent.SpinnerFrame = 0
	respCh := make(chan agent.ToolInteractResponse, 1)
	m.toolInteractForm = newToolApprovalForm(agent.ToolInteractRequest{
		Kind: agent.ToolInteractApproval,
		Name: "Bash",
	}, 60)
	m.toolInteractPending = toolInteractOffer{
		Req:    agent.ToolInteractRequest{Kind: agent.ToolInteractApproval, Name: "Bash"},
		RespCh: respCh,
	}

	m, cmd := m.completeToolInteractWith(agent.ToolInteractResponse{Approved: true})
	require.NotNil(t, cmd)
	require.True(t, (<-respCh).Approved)

	updated, tickCmd := m.Update(spinnerTickMsg{})
	m = updated.(Model)
	require.Equal(t, 1, m.agent.SpinnerFrame)
	require.NotNil(t, tickCmd)
}

func TestOfferToolInteractStartsSpinner(t *testing.T) {
	m := testInputModel(t)
	m = m.beginAgentTurn()
	m.agent.SpinnerFrame = 0

	offer := toolInteractOffer{
		Req: agent.ToolInteractRequest{
			Kind: agent.ToolInteractApproval,
			Name: "Bash",
			Args: map[string]any{"command": "echo hi"},
		},
		RespCh: make(chan agent.ToolInteractResponse, 1),
	}

	m, cmd := m.offerToolInteract(toolInteractOfferMsg{offer: offer})
	require.True(t, m.toolInteractDialogActive())
	require.NotNil(t, cmd)

	updated, tickCmd := m.Update(spinnerTickMsg{})
	m = updated.(Model)
	require.Equal(t, 1, m.agent.SpinnerFrame)
	require.NotNil(t, tickCmd)
}
