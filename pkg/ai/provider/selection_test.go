package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildModelConfigMissingAPIKey(t *testing.T) {
	dir := t.TempDir()
	writeProviderFile(t, dir, "demo.json", `{
		"baseUrl": "https://example.com/v1",
		"api": "openai-completions",
		"apiKey": "",
		"models": [{"id": "m1", "name": "Demo"}]
	}`)
	t.Setenv("ELPH_PROVIDERS_DIR", dir)

	catalog, err := LoadCatalog(dir)
	require.NoError(t, err)
	reg, ok := catalog.Provider("demo")
	require.True(t, ok)
	model, ok := FirstEnabledModel(reg)
	require.True(t, ok)

	cfg, err := BuildModelConfig(catalog, reg, model)
	require.ErrorIs(t, err, ErrMissingAPIKey)
	require.Nil(t, cfg.Provider)
	require.Equal(t, "demo", cfg.ProviderID)
	require.Equal(t, "m1", cfg.ModelID)
	require.Equal(t, "Demo", cfg.ModelName)
}

func TestBuildModelConfigUnresolvedEnv(t *testing.T) {
	dir := t.TempDir()
	writeProviderFile(t, dir, "demo.json", `{
		"baseUrl": "https://example.com/v1",
		"api": "openai-completions",
		"apiKey": "$MISSING_DEMO_KEY",
		"models": [{"id": "m1", "name": "Demo"}]
	}`)
	t.Setenv("ELPH_PROVIDERS_DIR", dir)
	t.Setenv("MISSING_DEMO_KEY", "")

	catalog, err := LoadCatalog(dir)
	require.NoError(t, err)
	reg, ok := catalog.Provider("demo")
	require.True(t, ok)
	model, ok := FirstEnabledModel(reg)
	require.True(t, ok)

	cfg, err := BuildModelConfig(catalog, reg, model)
	require.True(t, IsCredentialError(err))
	require.Nil(t, cfg.Provider)
	require.Equal(t, "m1", cfg.ModelID)
}

func TestCredentialHint(t *testing.T) {
	require.Contains(t, CredentialHint(RegisteredProvider{
		ID:     "openai",
		Config: FileConfig{APIKey: ""},
	}), "openai.json")
	require.Contains(t, CredentialHint(RegisteredProvider{
		ID:     "openai",
		Config: FileConfig{APIKey: "$OPENAI_API_KEY"},
	}), "environment variable")
}
