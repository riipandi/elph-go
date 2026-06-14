package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetProviderEnabledPersists(t *testing.T) {
	dir := t.TempDir()
	writeProviderFile(t, dir, "demo.json", `{
		"baseUrl": "https://example.com/v1",
		"api": "openai-completions",
		"apiKey": "$KEY",
		"models": [{"id": "m1"}]
	}`)
	t.Setenv("ELPH_PROVIDERS_DIR", dir)
	t.Setenv("KEY", "secret")

	require.NoError(t, SetProviderEnabled("demo", false))

	catalog, err := LoadCatalog(dir)
	require.NoError(t, err)
	reg, ok := catalog.Provider("demo")
	require.True(t, ok)
	require.False(t, ProviderConfigEnabled(reg.Config))
	require.False(t, reg.Models[0].Enabled)

	cfg := Resolve()
	require.NotEqual(t, "demo", cfg.ProviderID)
}

func TestSetModelEnabledPersists(t *testing.T) {
	dir := t.TempDir()
	writeProviderFile(t, dir, "demo.json", `{
		"baseUrl": "https://example.com/v1",
		"api": "openai-completions",
		"apiKey": "$KEY",
		"models": [
			{"id": "keep"},
			{"id": "drop"}
		]
	}`)
	t.Setenv("ELPH_PROVIDERS_DIR", dir)
	t.Setenv("KEY", "secret")

	require.NoError(t, SetModelEnabled("demo", "drop", false))

	raw, err := os.ReadFile(filepath.Join(dir, "demo.json"))
	require.NoError(t, err)

	var cfg FileConfig
	require.NoError(t, json.Unmarshal(raw, &cfg))
	require.False(t, ConfigEnabled(cfg.Models[1].Enabled))

	catalog, err := LoadCatalog(dir)
	require.NoError(t, err)
	reg, ok := catalog.Provider("demo")
	require.True(t, ok)
	require.True(t, reg.Models[0].Enabled)
	require.False(t, reg.Models[1].Enabled)
	require.Equal(t, 1, EnabledModelCount(reg))
}

func TestResolveSkipsDisabledSavedModel(t *testing.T) {
	dir := t.TempDir()
	writeProviderFile(t, dir, "demo.json", `{
		"baseUrl": "https://example.com/v1",
		"api": "openai-completions",
		"apiKey": "$KEY",
		"models": [
			{"id": "m1", "name": "One"},
			{"id": "m2", "name": "Two", "enabled": false}
		]
	}`)
	t.Setenv("ELPH_PROVIDERS_DIR", dir)
	t.Setenv("KEY", "secret")

	cfg := ResolveActive("demo", "m2")
	require.Equal(t, "m1", cfg.ModelID)
}
