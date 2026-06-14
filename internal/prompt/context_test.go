package prompt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatProjectContextSection(t *testing.T) {
	got := formatProjectContextSection("Use Go 1.26.", "/tmp/AGENTS.md")
	require.Contains(t, got, "<project_context>")
	require.Contains(t, got, "Project-specific instructions and guidelines:")
	require.Contains(t, got, `<project_instructions path="/tmp/AGENTS.md">`)
	require.Contains(t, got, "Use Go 1.26.")
	require.Contains(t, got, "</project_instructions>")
	require.Contains(t, got, "</project_context>")
}

func TestFormatRuntimeContextSection(t *testing.T) {
	got := formatRuntimeContextSection("2026-06-15", "/tmp/work")
	require.Equal(t, "Current date: 2026-06-15\nCurrent working directory: /tmp/work", got)
}

func TestFormatSessionStateSection(t *testing.T) {
	got := formatSessionStateSection("plan")
	require.Contains(t, got, "<session_state>")
	require.Contains(t, got, "<session_mode>plan</session_mode>")
	require.Contains(t, got, "</session_state>")
}
