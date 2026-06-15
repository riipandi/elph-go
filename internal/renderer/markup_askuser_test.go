package renderer

import (
	"testing"

	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/stretchr/testify/require"
)

func TestTryQueueMarkupAskUserSkipsResolvedQuestion(t *testing.T) {
	m := testInputModel(t)
	call := agent.ParsedToolCall{
		Name: "AskUser",
		Parameters: map[string]string{
			"question": "Pick a language",
			"options":  `["English", "Indonesia"]`,
		},
	}
	sig := (Model{}).toolCallSignature(call)
	m.agent.ResolvedAskUsers = map[string]askUserResolution{
		sig: {Answer: "English"},
	}

	updated, queued := m.tryQueueMarkupAskUser(call)
	require.False(t, queued)
	require.Nil(t, updated.agent.MarkupAskUserPending)
}

func TestHandleMarkupAskUserCmdSkipsAlreadyResolved(t *testing.T) {
	m := testInputModel(t)
	m.agent.MarkupAskUserPending = &markupAskUserOffer{
		Name: "AskUser",
		Parameters: map[string]string{
			"question": "Pick a language",
			"options":  `["English", "Indonesia"]`,
		},
	}
	req := agent.ToolInteractRequest{
		Kind: agent.ToolInteractAskUser,
		Name: "AskUser",
		Args: parsedToolParamsToAny(m.agent.MarkupAskUserPending.Parameters),
	}
	m = m.recordAskUserResolution(req, agent.ToolInteractResponse{Answer: "English"})

	updated, cmd := m.handleMarkupAskUserCmd()
	require.Nil(t, cmd)
	require.False(t, updated.toolInteractDialogActive())
	require.Nil(t, updated.agent.MarkupAskUserPending)
}

func TestBridgeInteractReturnsResolvedAskUserWithoutDialog(t *testing.T) {
	bridge := newToolInteractBridge()
	store := map[string]askUserResolution{
		toolInteractAskUserSignature(agent.ToolInteractRequest{
			Kind: agent.ToolInteractAskUser,
			Name: "AskUser",
			Args: map[string]any{
				"question": "Pick one",
				"options":  []any{"A", "B"},
			},
		}): {Answer: "B"},
	}
	bridge.ResolvedAskUsers = &store

	resp, err := bridge.Interact(t.Context(), agent.ToolInteractRequest{
		Kind: agent.ToolInteractAskUser,
		Name: "AskUser",
		Args: map[string]any{
			"question": "Pick one",
			"options":  []any{"A", "B"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "B", resp.Answer)
}

func TestRecordAskUserResolutionClearsPendingMarkup(t *testing.T) {
	m := testInputModel(t)
	m.agent.MarkupAskUserPending = &markupAskUserOffer{
		Name:       "AskUser",
		Parameters: map[string]string{"question": "Pick one"},
	}
	req := agent.ToolInteractRequest{
		Kind: agent.ToolInteractAskUser,
		Name: "AskUser",
		Args: map[string]any{"question": "Pick one"},
	}

	m = m.recordAskUserResolution(req, agent.ToolInteractResponse{Cancelled: true})
	require.Nil(t, m.agent.MarkupAskUserPending)
}
