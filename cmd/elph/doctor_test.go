package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeDoctorProvider(t *testing.T, dir, name, body string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
}

func setupDoctorHome(t *testing.T) (home, providersDir string) {
	t.Helper()
	home = t.TempDir()
	providersDir = filepath.Join(home, ".elph", "providers")
	require.NoError(t, os.MkdirAll(providersDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".elph", "settings.json"), []byte(`{
		"syncInterval": "24h",
		"session": {"providerId": "demo", "modelId": "m1"}
	}`), 0o644))
	t.Setenv("HOME", home)
	t.Setenv("ELPH_PROVIDERS_DIR", providersDir)
	return home, providersDir
}

func TestDoctorHealthyConfig(t *testing.T) {
	_, providersDir := setupDoctorHome(t)
	writeDoctorProvider(t, providersDir, "demo.json", `{
		"name": "Demo",
		"baseUrl": "https://example.com/v1",
		"api": "openai-completions",
		"apiKey": "secret",
		"models": [{"id": "m1", "name": "Demo Model"}]
	}`)
	t.Setenv("ELPH_PROVIDER", "demo")
	t.Setenv("ELPH_MODEL", "m1")

	report, err := runDoctorChecks(t.TempDir())
	require.NoError(t, err)
	require.False(t, report.hasFailures())

	var buf bytes.Buffer
	report.write(&buf)
	out := buf.String()
	require.Contains(t, out, "credentials verified for runtime model")
	require.Contains(t, out, "ELPH_PROVIDER=demo")
}

func TestDoctorMissingProvidersDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	missing := filepath.Join(home, ".elph", "providers")
	t.Setenv("ELPH_PROVIDERS_DIR", missing)

	report, err := runDoctorChecks(t.TempDir())
	require.NoError(t, err)
	require.True(t, report.hasFailures())

	var buf bytes.Buffer
	report.write(&buf)
	require.Contains(t, buf.String(), "provider connect")
}

func TestDoctorWarnsOnMissingAPIKey(t *testing.T) {
	_, providersDir := setupDoctorHome(t)
	writeDoctorProvider(t, providersDir, "demo.json", `{
		"name": "Demo",
		"baseUrl": "https://example.com/v1",
		"api": "openai-completions",
		"apiKey": "",
		"models": [{"id": "m1", "name": "Demo Model"}]
	}`)

	report, err := runDoctorChecks(t.TempDir())
	require.NoError(t, err)
	require.False(t, report.hasFailures())

	var buf bytes.Buffer
	report.write(&buf)
	require.Contains(t, buf.String(), "add apiKey")
}

func TestDoctorFailsOnInvalidProviderFile(t *testing.T) {
	_, providersDir := setupDoctorHome(t)
	writeDoctorProvider(t, providersDir, "broken.json", `{not json`)

	report, err := runDoctorChecks(t.TempDir())
	require.NoError(t, err)
	require.True(t, report.hasFailures())
}

func TestDoctorWarnsOnInvalidSyncInterval(t *testing.T) {
	home := t.TempDir()
	providersDir := filepath.Join(home, ".elph", "providers")
	require.NoError(t, os.MkdirAll(providersDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".elph", "settings.json"), []byte(`{
		"syncInterval": "not-a-duration"
	}`), 0o644))
	t.Setenv("HOME", home)
	t.Setenv("ELPH_PROVIDERS_DIR", providersDir)
	writeDoctorProvider(t, providersDir, "demo.json", `{
		"baseUrl": "https://example.com/v1",
		"api": "openai-completions",
		"apiKey": "secret",
		"models": [{"id": "m1", "name": "Demo"}]
	}`)

	report, err := runDoctorChecks(t.TempDir())
	require.NoError(t, err)
	require.False(t, report.hasFailures())

	var buf bytes.Buffer
	report.write(&buf)
	require.Contains(t, buf.String(), "syncInterval")
	require.Contains(t, buf.String(), "warn")
}
