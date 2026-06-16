package settings

import (
	"os"
	"testing"

	"github.com/riipandi/elph/internal/appconst"
	"github.com/stretchr/testify/require"
)

func TestSessionDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, appconst.ModeBuild, cfg.AgentMode())
	require.Equal(t, appconst.ThinkingHigh, cfg.ThinkingLevel())
	require.Empty(t, cfg.ActiveProviderID())
	require.Empty(t, cfg.ActiveModelID())
}

func TestSetActiveModelPersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, SetActiveModel("anthropic", "claude-sonnet-4"))

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "anthropic", cfg.ActiveProviderID())
	require.Equal(t, "claude-sonnet-4", cfg.ActiveModelID())
}

func TestSetAgentModePersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, SetAgentMode(appconst.ModePlan))

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, appconst.ModePlan, cfg.AgentMode())
}

func TestSetThinkingLevelPersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, SetThinkingLevel(appconst.ThinkingLow))

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, appconst.ThinkingLow, cfg.ThinkingLevel())
}

func TestSessionNormalizesInvalidValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, Save(Settings{
		Session: SessionSettings{
			AgentMode:     "nope",
			ThinkingLevel: "nope",
		},
	}))

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, appconst.ModeBuild, cfg.AgentMode())
	require.Equal(t, appconst.ThinkingHigh, cfg.ThinkingLevel())
}

func TestSessionRoundTripPreservesOtherFields(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	disabled := false
	require.NoError(t, Save(Settings{
		SyncInterval: "6h",
		ShowThinking: &disabled,
	}))

	require.NoError(t, SetActiveModel("openai", "gpt-4o"))
	require.NoError(t, SetAgentMode(appconst.ModeAsk))
	require.NoError(t, SetThinkingLevel(appconst.ThinkingMedium))

	cfg, err := Load()
	require.NoError(t, err)
	require.False(t, cfg.ShowThinkingEnabled())
	require.Equal(t, "6h", cfg.SyncInterval)
	require.Equal(t, "openai", cfg.ActiveProviderID())
	require.Equal(t, "gpt-4o", cfg.ActiveModelID())
	require.Equal(t, appconst.ModeAsk, cfg.AgentMode())
	require.Equal(t, appconst.ThinkingMedium, cfg.ThinkingLevel())

	raw, err := os.ReadFile(mustPath(t))
	require.NoError(t, err)
	require.Contains(t, string(raw), `"providerId": "openai"`)
	require.Contains(t, string(raw), `"agentMode": "ask"`)
}

func mustPath(t *testing.T) string {
	t.Helper()
	path, err := Path()
	require.NoError(t, err)
	return path
}
