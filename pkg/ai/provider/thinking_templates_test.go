package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackfillProviderThinkingPreservesUserOverrides(t *testing.T) {
	existing := FileConfig{
		Models: []ModelConfig{
			{
				ID:        "claude-opus-4-20250514",
				Reasoning: true,
				ThinkingLevelMap: map[string]json.RawMessage{
					"high": json.RawMessage(`"custom"`),
				},
			},
		},
	}

	updated, changed := BackfillProviderThinking("anthropic", existing)
	require.True(t, changed)
	require.Equal(t, json.RawMessage(`"custom"`), updated.Models[0].ThinkingLevelMap["high"])
	require.True(t, updated.Models[0].Compat.ForceAdaptiveThinking)
}

func TestBackfillProviderThinkingFillsMissingMap(t *testing.T) {
	existing := FileConfig{
		Models: []ModelConfig{
			{
				ID:        "o3-mini",
				Reasoning: true,
			},
		},
	}

	updated, changed := BackfillProviderThinking("openai", existing)
	require.True(t, changed)
	require.NotEmpty(t, updated.Models[0].ThinkingLevelMap)
	require.Equal(t, json.RawMessage(`null`), updated.Models[0].ThinkingLevelMap["off"])
}

func TestBackfillAllProviderThinkingWritesFiles(t *testing.T) {
	dir := t.TempDir()
	initial := FileConfig{
		Name:    "Anthropic",
		BaseURL: "https://api.anthropic.com/v1",
		API:     APIAnthropicMessages,
		Models: []ModelConfig{
			{ID: "claude-opus-4-20250514", Reasoning: true},
		},
	}
	raw, err := json.MarshalIndent(initial, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "anthropic.json"), append(raw, '\n'), 0o644))

	result, err := BackfillAllProviderThinking(dir)
	require.NoError(t, err)
	require.Equal(t, []string{"anthropic.json"}, result.Backfilled)

	var saved FileConfig
	body, err := os.ReadFile(filepath.Join(dir, "anthropic.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &saved))
	require.True(t, saved.Models[0].Compat.ForceAdaptiveThinking)
	require.NotEmpty(t, saved.Models[0].ThinkingLevelMap)
}

func TestBootstrapProvidersBackfillsExistingFiles(t *testing.T) {
	dir := t.TempDir()
	initial := FileConfig{
		Name:    "OpenAI",
		BaseURL: "https://api.openai.com/v1",
		API:     APIOpenAICompletions,
		Models: []ModelConfig{
			{ID: "o3-mini", Reasoning: true},
		},
	}
	raw, err := json.MarshalIndent(initial, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openai.json"), append(raw, '\n'), 0o644))

	result, err := BootstrapProviders(dir, false)
	require.NoError(t, err)
	require.Contains(t, result.Backfilled, "openai.json")
	require.NotContains(t, result.Skipped, "openai.json")
}
