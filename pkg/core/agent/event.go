package agent

import "github.com/riipandi/elph/pkg/ai/protocol"

// EventKind identifies agent runtime events emitted during a turn.
type EventKind int

const (
	EventActivity EventKind = iota
	EventThinkingDelta
	EventResponseDelta
	EventToolCallStart
	EventToolCallOutputDelta
	EventToolCallDone
	EventTurnDone
)

// Event is a framework-neutral agent runtime message.
type Event struct {
	Kind        EventKind
	Activity    Activity
	Delta       string
	Thinking    string
	Response    string
	Usage       protocol.TurnUsage
	ToolCall    protocol.ToolCall
	ToolResult  ToolRunResult
	History     []protocol.ChatMessage
	ProviderErr error
}

// ActivityEvent returns an activity phase update.
func ActivityEvent(activity Activity) Event {
	return Event{Kind: EventActivity, Activity: activity}
}

// ThinkingDeltaEvent returns an incremental reasoning chunk.
func ThinkingDeltaEvent(delta string) Event {
	return Event{Kind: EventThinkingDelta, Delta: delta}
}

// ResponseDeltaEvent returns an incremental assistant response chunk.
func ResponseDeltaEvent(delta string) Event {
	return Event{Kind: EventResponseDelta, Delta: delta}
}

// ToolCallStartEvent announces a provider-native tool invocation.
func ToolCallStartEvent(call protocol.ToolCall) Event {
	return Event{Kind: EventToolCallStart, ToolCall: call}
}

// ToolCallOutputDeltaEvent streams incremental tool output to the host UI.
func ToolCallOutputDeltaEvent(call protocol.ToolCall, delta string) Event {
	return Event{Kind: EventToolCallOutputDelta, ToolCall: call, Delta: delta}
}

// ToolCallDoneEvent reports a completed tool invocation.
func ToolCallDoneEvent(call protocol.ToolCall, result ToolRunResult) Event {
	return Event{Kind: EventToolCallDone, ToolCall: call, ToolResult: result}
}

// TurnDoneEvent returns a completed turn with the final assistant response.
func TurnDoneEvent(result protocol.TurnResult) Event {
	return Event{
		Kind:     EventTurnDone,
		Thinking: result.Thinking,
		Response: result.Content,
		Usage:    result.Usage,
	}
}

// TurnDoneWithHistoryEvent returns a completed turn and updated conversation history.
func TurnDoneWithHistoryEvent(result protocol.TurnResult, history []protocol.ChatMessage) Event {
	return Event{
		Kind:     EventTurnDone,
		Thinking: result.Thinking,
		Response: result.Content,
		Usage:    result.Usage,
		History:  history,
	}
}

// TurnDoneProviderErrorEvent returns a failed turn with provider error details.
func TurnDoneProviderErrorEvent(err error, history []protocol.ChatMessage) Event {
	return Event{
		Kind:        EventTurnDone,
		Response:    protocol.ProviderErrorSummary(err),
		History:     history,
		ProviderErr: err,
	}
}
