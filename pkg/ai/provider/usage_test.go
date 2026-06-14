package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTurnCostUSD(t *testing.T) {
	cost := Cost{Input: 3, Output: 15}
	got := cost.TurnCostUSD(TurnUsage{InputTokens: 1_000_000, OutputTokens: 1_000_000})
	require.InDelta(t, 18, got, 0.001)
}

func TestSupportsImageInput(t *testing.T) {
	require.True(t, SupportsImageInput([]string{"text", "image"}))
	require.False(t, SupportsImageInput([]string{"text"}))
}
