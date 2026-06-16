package protocol

// ThinkingLevel is the UI/runtime reasoning intensity for a model turn.
type ThinkingLevel string

const (
	ThinkingOff     ThinkingLevel = "off"
	ThinkingMinimal ThinkingLevel = "minimal"
	ThinkingLow     ThinkingLevel = "low"
	ThinkingMedium  ThinkingLevel = "medium"
	ThinkingHigh    ThinkingLevel = "high"
	ThinkingXHigh   ThinkingLevel = "xhigh"
)

var thinkingLevels = []ThinkingLevel{
	ThinkingOff,
	ThinkingMinimal,
	ThinkingLow,
	ThinkingMedium,
	ThinkingHigh,
	ThinkingXHigh,
}

// NextThinkingLevel cycles forward through supported levels.
func NextThinkingLevel(lvl ThinkingLevel) ThinkingLevel {
	for i, l := range thinkingLevels {
		if l == lvl {
			return thinkingLevels[(i+1)%len(thinkingLevels)]
		}
	}
	return thinkingLevels[0]
}

// PrevThinkingLevel cycles backward through supported levels.
func PrevThinkingLevel(lvl ThinkingLevel) ThinkingLevel {
	for i, l := range thinkingLevels {
		if l == lvl {
			p := i - 1
			if p < 0 {
				p = len(thinkingLevels) - 1
			}
			return thinkingLevels[p]
		}
	}
	return thinkingLevels[len(thinkingLevels)-1]
}

// ThinkingFormat selects how reasoning parameters are sent to OpenAI-compatible APIs.
type ThinkingFormat string

const (
	ThinkingFormatReasoningEffort ThinkingFormat = "reasoning_effort"
	ThinkingFormatOpenRouter      ThinkingFormat = "openrouter"
	ThinkingFormatQwen            ThinkingFormat = "qwen"
	ThinkingFormatDeepSeek        ThinkingFormat = "deepseek"
)

// ThinkingConfig is the resolved thinking payload for one turn.
type ThinkingConfig struct {
	Enabled bool

	BudgetTokens int

	Adaptive       bool
	AdaptiveEffort string

	ReasoningEffort string
	ThinkingFormat  ThinkingFormat
	EnableThinking  bool
}
