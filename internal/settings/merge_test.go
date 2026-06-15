package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/riipandi/elph/internal/constants"
	"github.com/riipandi/elph/internal/projectdir"
	"github.com/stretchr/testify/require"
)

func TestActiveSettingsPathPrefersJSON(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "settings.jsonc"), []byte(`{"theme":"jsonc"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"theme":"json"}`), 0o644))

	path, ok := activeSettingsPath(dir)
	require.True(t, ok)
	require.Equal(t, filepath.Join(dir, "settings.json"), path)
}

func TestLoadForMergesProjectOverHome(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	showThinking := true
	require.NoError(t, Save(Settings{
		Theme:        "dark",
		ShowThinking: &showThinking,
		SyncInterval: "12h",
		Session: SessionSettings{
			AgentMode: string(constants.ModeBuild),
		},
	}))

	projectDir := projectdir.Root(work)
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(projectDir, "settings.json"),
		[]byte(`{
			"theme": "light",
			"stickyScroll": false,
			"session": { "agentMode": "plan" }
		}`),
		0o644,
	))

	cfg, err := LoadFor(work)
	require.NoError(t, err)
	require.Equal(t, "light", cfg.Theme)
	require.False(t, cfg.StickyScrollEnabled())
	require.True(t, cfg.ShowThinkingEnabled())
	require.Equal(t, "12h", cfg.SyncInterval)
	require.Equal(t, constants.ModePlan, cfg.AgentMode())
}

func TestLoadForProjectJSONCWhenJSONMissing(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, Ensure())

	projectDir := projectdir.Root(work)
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(projectDir, "settings.jsonc"),
		[]byte(`{"preferedResponseLanguage":"Indonesian"}`),
		0o644,
	))

	cfg, err := LoadFor(work)
	require.NoError(t, err)
	require.Equal(t, "Indonesian", cfg.ResponseLanguage())
}

func TestLoadForIgnoresEmptyProjectSettings(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, Save(Settings{Theme: "dark"}))

	projectDir := projectdir.Root(work)
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "settings.jsonc"), []byte(`{}`), 0o644))

	cfg, err := LoadFor(work)
	require.NoError(t, err)
	require.Equal(t, "dark", cfg.Theme)
}

func TestLoadForProjectDoesNotOverrideHomeModelSelection(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, Save(Settings{
		Session: SessionSettings{
			ProviderID: "home-provider",
			ModelID:    "home-model",
		},
	}))

	projectDir := projectdir.Root(work)
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(projectDir, "settings.json"),
		[]byte(`{
			"session": {
				"providerId": "project-provider",
				"modelId": "project-model",
				"agentMode": "plan"
			}
		}`),
		0o644,
	))

	cfg, err := LoadFor(work)
	require.NoError(t, err)
	require.Equal(t, "home-provider", cfg.ActiveProviderID())
	require.Equal(t, "home-model", cfg.ActiveModelID())
	require.Equal(t, constants.ModePlan, cfg.AgentMode())
}

func TestLoadForProjectSuppliesDefaultModelWhenHomeUnset(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, Ensure())

	projectDir := projectdir.Root(work)
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(projectDir, "settings.json"),
		[]byte(`{"session":{"providerId":"project-provider","modelId":"project-model"}}`),
		0o644,
	))

	cfg, err := LoadFor(work)
	require.NoError(t, err)
	require.Equal(t, "project-provider", cfg.ActiveProviderID())
	require.Equal(t, "project-model", cfg.ActiveModelID())
}

func TestLoadForMergesLegacyNestedSyncInterval(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, Ensure())

	projectDir := projectdir.Root(work)
	require.NoError(t, os.MkdirAll(projectDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(projectDir, "settings.json"),
		[]byte(`{"models":{"syncInterval":"8h"}}`),
		0o644,
	))

	cfg, err := LoadFor(work)
	require.NoError(t, err)
	require.Equal(t, "8h", cfg.SyncInterval)
}

func TestMergeSettingsThinkingBudgets(t *testing.T) {
	base := Settings{ThinkingBudgets: map[string]int{"high": 1000}}
	overlay := Settings{ThinkingBudgets: map[string]int{"low": 200, "high": 1500}}

	merged := mergeSettings(base, overlay)
	require.Equal(t, 1500, merged.ThinkingBudgets["high"])
	require.Equal(t, 200, merged.ThinkingBudgets["low"])
}
