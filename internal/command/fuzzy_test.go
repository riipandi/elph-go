package command

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuzzyScoreSubsequence(t *testing.T) {
	require.Positive(t, fuzzyScore("quit", "quit"))
	require.Positive(t, fuzzyScore("qt", "quit"))
	require.Positive(t, fuzzyScore("diag", "diagnostic:list-tools"))
	require.Equal(t, -1, fuzzyScore("zzz", "help"))
}

func TestCommandScoreUsesAliases(t *testing.T) {
	cmd, ok := Get("exit")
	require.True(t, ok)
	require.Positive(t, commandScore("quit", cmd))
	require.Positive(t, commandScore("qt", cmd))
}
