package openai

import (
	"encoding/json"
	"testing"

	provider "github.com/riipandi/elph/pkg/ai/protocol"
	"github.com/stretchr/testify/require"
)

func TestChatMessagesUsesSystemWhenDeveloperRoleUnsupported(t *testing.T) {
	msgs := chatMessages(
		"system instructions",
		nil,
		provider.ThinkingConfig{Enabled: true},
		provider.Compat{SupportsDeveloperRole: boolPtr(false)},
	)
	require.Len(t, msgs, 1)
	raw, err := json.Marshal(msgs[0])
	require.NoError(t, err)
	require.Contains(t, string(raw), `"role":"system"`)
	require.NotContains(t, string(raw), `"role":"developer"`)
}

func TestChatMessagesUsesDeveloperWhenSupported(t *testing.T) {
	msgs := chatMessages(
		"system instructions",
		nil,
		provider.ThinkingConfig{Enabled: true},
		provider.Compat{SupportsDeveloperRole: boolPtr(true)},
	)
	require.Len(t, msgs, 1)
	raw, err := json.Marshal(msgs[0])
	require.NoError(t, err)
	require.Contains(t, string(raw), `"role":"developer"`)
}

func boolPtr(v bool) *bool { return &v }
