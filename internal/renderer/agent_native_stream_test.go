package renderer

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/stretchr/testify/require"
)

func TestBeginNativeToolCallBashShowsCommandLabel(t *testing.T) {
	m := testInputModel(t)
	call := provider.ToolCall{
		ID:        "call_bash",
		Name:      "Bash",
		Arguments: json.RawMessage(`{"command":"ping 1.1.1.1"}`),
	}
	m = m.beginNativeToolCall(call)
	idx := m.agent.NativeToolMsgIDs["call_bash"]
	require.Equal(t, "$ ping 1.1.1.1", m.messages[idx].detailLabel)
}

func TestAppendNativeToolOutputStreamsIntoDetailBox(t *testing.T) {
	m := testInputModel(t)
	call := provider.ToolCall{
		ID:        "call_stream",
		Name:      "Bash",
		Arguments: json.RawMessage(`{"command":"echo hi"}`),
	}
	m = m.beginNativeToolCall(call)

	m = m.appendNativeToolOutput(call, "he")
	m = m.appendNativeToolOutput(call, "llo\n")

	idx := m.agent.NativeToolMsgIDs["call_stream"]
	require.Equal(t, "hello", strings.TrimRight(m.messages[idx].text, "\n"))

	view := stripANSI(m.renderMessageAt(idx))
	require.Contains(t, view, "hello")
	require.NotContains(t, view, "Running...")
}

func TestAppendNativeToolOutputHonorsCarriageReturn(t *testing.T) {
	m := testInputModel(t)
	call := provider.ToolCall{
		ID:        "call_ping",
		Name:      "Bash",
		Arguments: json.RawMessage(`{"command":"ping 1.1.1.1"}`),
	}
	m = m.beginNativeToolCall(call)

	m = m.appendNativeToolOutput(call, "PING 1.1.1.1\n64 bytes")
	m = m.appendNativeToolOutput(call, "\r128 bytes")

	idx := m.agent.NativeToolMsgIDs["call_ping"]
	require.Equal(t, "PING 1.1.1.1\n128 bytes", m.messages[idx].text)
}

func TestAppendNativeToolOutputPreservesChunkBoundaryNewlines(t *testing.T) {
	m := testInputModel(t)
	call := provider.ToolCall{
		ID:        "call_ping_chunks",
		Name:      "Bash",
		Arguments: json.RawMessage(`{"command":"ping 1.1.1.1"}`),
	}
	m = m.beginNativeToolCall(call)

	m = m.appendNativeToolOutput(call, "PING 1.1.1.1 (1.1.1.1): 56 data bytes\n64 bytes from 1.1.1.1: icmp_seq=0 ttl=58 time=9.046 ms\n")
	m = m.appendNativeToolOutput(call, "64 bytes from 1.1.1.1: icmp_seq=1 ttl=58 time=9.158 ms\n")

	idx := m.agent.NativeToolMsgIDs["call_ping_chunks"]
	require.Contains(t, m.messages[idx].text, "ms\n64 bytes")
	require.NotContains(t, m.messages[idx].text, "ms64 bytes")
}

func TestFinishNativeToolCallBashKeepsRawStreamedOutput(t *testing.T) {
	m := testInputModel(t)
	call := provider.ToolCall{
		ID:        "call_bash_done",
		Name:      "Bash",
		Arguments: json.RawMessage(`{"command":"false"}`),
	}
	m = m.beginNativeToolCall(call)
	m = m.appendNativeToolOutput(call, "failed\n")

	m = m.finishNativeToolCall(call, agent.ToolRunResult{Output: "failed\n\n(exit 1)"})

	idx := m.agent.NativeToolMsgIDs["call_bash_done"]
	require.Equal(t, "$ false", m.messages[idx].detailLabel)
	require.Equal(t, "failed\n(exit 1)", m.messages[idx].text)
	require.NotContains(t, m.messages[idx].text, "Tool failed")
}
