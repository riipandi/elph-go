package settings

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadVersionDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	v, err := LoadVersion()
	require.NoError(t, err)
	require.Empty(t, v.LastSyncProviders)
	require.Equal(t, dummyStableVersion, v.StableVersion)
	require.Equal(t, dummyStableVersion, v.Version)
	require.Equal(t, dummyReleaseCheckedAt, v.ReleaseCheckedAt)
}

func TestMarkProvidersSyncedWritesVersionFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	now := time.Date(2026, 6, 13, 15, 30, 0, 0, time.UTC)
	require.NoError(t, MarkProvidersSynced(now))

	path, err := VersionPath()
	require.NoError(t, err)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"lastSyncProviders": "2026-06-13T15:30:00Z"`)
	require.Contains(t, string(raw), `"stableVersion": "0.2.16"`)
	require.Contains(t, string(raw), `"relaseCheckedAt"`)
}

func TestMigrateLastSyncFromSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	last := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	require.NoError(t, Save(Settings{
		SyncInterval: "24h",
		Models: &ModelsSettings{
			LastSync: last.Format(time.RFC3339),
		},
	}))

	v, err := LoadVersion()
	require.NoError(t, err)
	require.Equal(t, last.Format(time.RFC3339), v.LastSyncProviders)

	path, err := VersionPath()
	require.NoError(t, err)
	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestVersionPathUsesElphHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := VersionPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".elph", "version.json"), path)
}
