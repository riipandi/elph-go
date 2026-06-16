package renderer

import (
	"context"
	"testing"

	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/stretchr/testify/require"
)

func TestToolInteractBridgeSkipsApprovalOnlyAfterSessionAllow(t *testing.T) {
	bridge := newToolInteractBridge()
	req := agent.ToolInteractRequest{Kind: agent.ToolInteractApproval, Name: "Bash"}

	go func() {
		resp, err := bridge.Interact(t.Context(), req)
		require.NoError(t, err)
		require.True(t, resp.Approved)
		require.False(t, resp.AllowSession)
		msg := waitToolInteractOffer(bridge)().(toolInteractOfferMsg)
		msg.Offer.RespCh <- agent.ToolInteractResponse{Approved: true}
	}()

	bridge.SkipSessionApproval = true
	resp, err := bridge.Interact(t.Context(), req)
	require.NoError(t, err)
	require.True(t, resp.Approved)
	require.False(t, resp.AllowSession)
}

func TestToolInteractBridgeSessionAllowDoesNotStickWithoutFlag(t *testing.T) {
	bridge := newToolInteractBridge()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go func() {
		_, _ = bridge.Interact(ctx, agent.ToolInteractRequest{Kind: agent.ToolInteractApproval, Name: "Bash"})
	}()

	msg := waitToolInteractOffer(bridge)().(toolInteractOfferMsg)
	require.NotNil(t, msg.Offer.RespCh)
}
