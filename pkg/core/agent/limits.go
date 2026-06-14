package agent

// Tool and conversation limits to bound memory and provider payload size.
const (
	// MaxProviderToolBytes is the max tool result sent back to the model API.
	MaxProviderToolBytes = 32 << 10

	// MaxDisplayToolBytes is the max tool output kept for TUI detail boxes and events.
	MaxDisplayToolBytes = 40 << 10

	// MaxHistoryMessages is the max provider messages retained across turns.
	MaxHistoryMessages = 32

	// MaxHistoryBytes is the approximate total UTF-8 size of retained history.
	MaxHistoryBytes = 512 << 10

	// MaxAssistantHistoryBytes caps a single assistant message body in history.
	MaxAssistantHistoryBytes = 64 << 10

	// MaxUIMessageBytes caps stored bubble text for AI and detail messages.
	MaxUIMessageBytes = 48 << 10
)
