package toolinteract

import (
	"strings"

	"github.com/riipandi/elph/pkg/core/agent"
)

// AskUserResolution records a completed AskUser gate so native and markup paths
// do not reopen the same question in one user-initiated turn chain.
type AskUserResolution struct {
	Answer    string
	Cancelled bool
}

// LookupResolvedAskUser returns a cached answer when the same ask-user was already resolved.
func LookupResolvedAskUser(store *map[string]AskUserResolution, req agent.ToolInteractRequest) (agent.ToolInteractResponse, bool) {
	if store == nil || *store == nil || req.Kind != agent.ToolInteractAskUser {
		return agent.ToolInteractResponse{}, false
	}
	res, ok := (*store)[AskUserSignature(req)]
	if !ok {
		return agent.ToolInteractResponse{}, false
	}
	if res.Cancelled {
		return agent.ToolInteractResponse{Cancelled: true}, true
	}
	return agent.ToolInteractResponse{Answer: res.Answer}, true
}

// RecordAskUserResolution stores the response for deduplication within a turn chain.
func RecordAskUserResolution(store *map[string]AskUserResolution, req agent.ToolInteractRequest, resp agent.ToolInteractResponse) {
	if store == nil || *store == nil || req.Kind != agent.ToolInteractAskUser {
		return
	}
	(*store)[AskUserSignature(req)] = AskUserResolution{
		Answer:    strings.TrimSpace(resp.Answer),
		Cancelled: resp.Cancelled,
	}
}

// AskUserSignature builds a stable key for ask-user deduplication.
func AskUserSignature(req agent.ToolInteractRequest) string {
	params := make(map[string]string)
	for key := range req.Args {
		if val, ok := StringArg(req.Args, key); ok {
			params[key] = val
		}
	}
	return ToolCallSignature(agent.ParsedToolCall{
		Name:       req.Name,
		Parameters: params,
	})
}

// ParsedAskUserResolved reports whether a parsed tool call was already answered.
func ParsedAskUserResolved(store map[string]AskUserResolution, call agent.ParsedToolCall) bool {
	if store == nil {
		return false
	}
	_, ok := store[ToolCallSignature(call)]
	return ok
}
