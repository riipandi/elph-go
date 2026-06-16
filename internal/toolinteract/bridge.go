package toolinteract

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/riipandi/elph/pkg/core/agent"
)

// Offer is a pending tool-interact dialog delivered to the TUI.
type Offer struct {
	Req        agent.ToolInteractRequest
	RespCh     chan<- agent.ToolInteractResponse
	FromMarkup bool
}

// OfferMsg is a Bubble Tea message carrying a tool-interact offer.
type OfferMsg struct {
	Offer Offer
}

// Bridge connects the agent turn loop to the TUI approval/ask-user dialogs.
type Bridge struct {
	Inbox               chan Offer
	SkipSessionApproval bool
	DeniedApprovals     map[string]struct{}
	ResolvedAskUsers    *map[string]AskUserResolution
}

// NewBridge allocates a bridge with a buffered offer inbox.
func NewBridge() *Bridge {
	return &Bridge{Inbox: make(chan Offer, 1)}
}

// Interact blocks until the TUI delivers a response or ctx is cancelled.
func (b *Bridge) Interact(ctx context.Context, req agent.ToolInteractRequest) (agent.ToolInteractResponse, error) {
	if req.Kind == agent.ToolInteractAskUser {
		if resp, ok := LookupResolvedAskUser(b.ResolvedAskUsers, req); ok {
			return resp, nil
		}
	}
	if b.SkipSessionApproval && req.Kind == agent.ToolInteractApproval {
		return agent.ToolInteractResponse{Approved: true}, nil
	}
	if req.Kind == agent.ToolInteractApproval && b.DeniedApprovals != nil {
		if _, denied := b.DeniedApprovals[ApprovalSignature(req)]; denied {
			return agent.ToolInteractResponse{Approved: false}, nil
		}
	}
	respCh := make(chan agent.ToolInteractResponse, 1)
	select {
	case b.Inbox <- Offer{Req: req, RespCh: respCh}:
	case <-ctx.Done():
		return agent.ToolInteractResponse{}, ctx.Err()
	}
	select {
	case resp := <-respCh:
		return resp, nil
	case <-ctx.Done():
		return agent.ToolInteractResponse{}, ctx.Err()
	}
}

// WaitOffer returns a command that reads the next offer from the bridge inbox.
func WaitOffer(b *Bridge) tea.Cmd {
	if b == nil {
		return nil
	}
	return func() tea.Msg {
		offer, ok := <-b.Inbox
		if !ok {
			return nil
		}
		return OfferMsg{Offer: offer}
	}
}
