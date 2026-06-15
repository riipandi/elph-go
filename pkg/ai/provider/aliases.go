package provider

import "github.com/riipandi/elph/pkg/ai/protocol"

type (
	TurnRequest     = protocol.TurnRequest
	TurnResult      = protocol.TurnResult
	TurnStream      = protocol.TurnStream
	TurnUsage       = protocol.TurnUsage
	Provider        = protocol.Provider
	ChatMessage     = protocol.ChatMessage
	ImageAttachment = protocol.ImageAttachment
	ToolDefinition  = protocol.ToolDefinition
	ToolCall        = protocol.ToolCall
	StopReason      = protocol.StopReason
	ThinkingLevel   = protocol.ThinkingLevel
	ThinkingConfig  = protocol.ThinkingConfig
	ThinkingFormat  = protocol.ThinkingFormat
	Compat          = protocol.Compat
	ProviderError   = protocol.ProviderError
)

const (
	StopReasonEndTurn   = protocol.StopReasonEndTurn
	StopReasonToolUse   = protocol.StopReasonToolUse
	StopReasonMaxTokens = protocol.StopReasonMaxTokens

	ThinkingOff     = protocol.ThinkingOff
	ThinkingMinimal = protocol.ThinkingMinimal
	ThinkingLow     = protocol.ThinkingLow
	ThinkingMedium  = protocol.ThinkingMedium
	ThinkingHigh    = protocol.ThinkingHigh
	ThinkingXHigh   = protocol.ThinkingXHigh

	ThinkingFormatReasoningEffort = protocol.ThinkingFormatReasoningEffort
	ThinkingFormatOpenRouter      = protocol.ThinkingFormatOpenRouter
	ThinkingFormatQwen            = protocol.ThinkingFormatQwen
	ThinkingFormatDeepSeek        = protocol.ThinkingFormatDeepSeek

	DefaultMaxTokens = protocol.DefaultMaxTokens
)

var ErrMissingAPIKey = protocol.ErrMissingAPIKey

var (
	NextThinkingLevel = protocol.NextThinkingLevel
	PrevThinkingLevel = protocol.PrevThinkingLevel
)

var (
	ParseContextTooLarge             = protocol.ParseContextTooLarge
	HeaderMap                        = protocol.HeaderMap
	ErrorTitleForStatus              = protocol.ErrorTitleForStatus
	WrapUnexpectedEOF                = protocol.WrapUnexpectedEOF
	IsStreamJSONError                = protocol.IsStreamJSONError
	ShouldStreamNonStreamingFallback = protocol.ShouldStreamNonStreamingFallback
	NormalizeToolArguments           = protocol.NormalizeToolArguments
	ProviderErrorSummary             = protocol.ProviderErrorSummary
	FormatProviderErrorDetail        = protocol.FormatProviderErrorDetail
	MaxTokensOrDefault               = protocol.MaxTokensOrDefault
	BuildMessages                    = protocol.BuildMessages
)
