package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThinkingLevelCycle(t *testing.T) {
	require.Equal(t, ThinkingMinimal, NextThinkingLevel(ThinkingOff))
	require.Equal(t, ThinkingOff, PrevThinkingLevel(ThinkingMinimal))
	require.Equal(t, ThinkingOff, NextThinkingLevel(ThinkingXHigh))
	require.Equal(t, ThinkingXHigh, PrevThinkingLevel(ThinkingOff))
}
