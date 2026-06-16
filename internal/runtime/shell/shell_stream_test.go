package shell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyStreamChunkCarriageReturnOverwritesLine(t *testing.T) {
	t.Parallel()

	acc := ApplyStreamChunk("", "PING 1.1.1.1\n")
	acc = ApplyStreamChunk(acc, "64 bytes from 1.1.1.1")
	acc = ApplyStreamChunk(acc, "\r")
	acc = ApplyStreamChunk(acc, "128 bytes from 1.1.1.1")

	require.Equal(t, "PING 1.1.1.1\n128 bytes from 1.1.1.1", acc)
}

func TestApplyStreamChunkWindowsLineEnding(t *testing.T) {
	t.Parallel()
	require.Equal(t, "line1\nline2", ApplyStreamChunk("", "line1\r\nline2"))
}

func TestSanitizeStreamChunkStillNormalizesStandalone(t *testing.T) {
	t.Parallel()
	require.Equal(t, "a\nb", SanitizeStreamChunk("a\rb"))
}
