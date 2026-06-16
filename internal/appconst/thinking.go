package appconst

import "github.com/riipandi/elph/pkg/ai/protocol"

type ThinkingLevel = protocol.ThinkingLevel

const (
	ThinkingOff     = protocol.ThinkingOff
	ThinkingMinimal = protocol.ThinkingMinimal
	ThinkingLow     = protocol.ThinkingLow
	ThinkingMedium  = protocol.ThinkingMedium
	ThinkingHigh    = protocol.ThinkingHigh
	ThinkingXHigh   = protocol.ThinkingXHigh
)

var (
	NextThinkingLevel = protocol.NextThinkingLevel
	PrevThinkingLevel = protocol.PrevThinkingLevel
)
