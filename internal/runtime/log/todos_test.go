package log

import (
	"os"
	"strings"
	"testing"

	"github.com/riipandi/elph/internal/projectdir"
	"github.com/riipandi/elph/pkg/tools/todolist"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadTodosJSONL(t *testing.T) {
	workDir := t.TempDir()
	require.NoError(t, SaveTodosSnapshot(workDir, "sess_a", []todolist.Todo{
		{Title: "alpha", Status: todolist.StatusPending},
	}))
	require.NoError(t, SaveTodosSnapshot(workDir, "sess_a", []todolist.Todo{
		{Title: "alpha", Status: todolist.StatusDone},
	}))
	require.NoError(t, SaveTodosSnapshot(workDir, "sess_b", []todolist.Todo{
		{Title: "beta", Status: todolist.StatusInProgress},
		{Title: "gamma", Status: todolist.StatusDone},
	}))

	loadedA, err := LoadTodos(workDir, "sess_a")
	require.NoError(t, err)
	require.Len(t, loadedA, 1)
	require.Equal(t, todolist.StatusDone, loadedA[0].Status)

	loadedB, err := LoadTodos(workDir, "sess_b")
	require.NoError(t, err)
	require.Len(t, loadedB, 2)
	require.Equal(t, "beta", loadedB[0].Title)
	require.Equal(t, todolist.StatusDone, loadedB[1].Status)

	rawA, err := os.ReadFile(projectdir.SessionTodosPath(workDir, "sess_a"))
	require.NoError(t, err)
	require.Contains(t, string(rawA), `"title":"alpha"`)
	require.NotContains(t, string(rawA), `"session"`)

	rawB, err := os.ReadFile(projectdir.SessionTodosPath(workDir, "sess_b"))
	require.NoError(t, err)
	require.Contains(t, string(rawB), `"title":"gamma"`)
}

func TestSaveTodosSnapshotReplacesInsteadOfAppending(t *testing.T) {
	workDir := t.TempDir()
	sessionID := "sess_compact"
	for range 12 {
		require.NoError(t, SaveTodosSnapshot(workDir, sessionID, []todolist.Todo{
			{Title: "task", Status: todolist.StatusInProgress},
		}))
	}

	raw, err := os.ReadFile(projectdir.SessionTodosPath(workDir, sessionID))
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(strings.TrimSpace(string(raw)), "\n")+1)
}

func TestSaveTodosSnapshotClearsFileWhenEmpty(t *testing.T) {
	workDir := t.TempDir()
	sessionID := "sess_clear"
	require.NoError(t, SaveTodosSnapshot(workDir, sessionID, []todolist.Todo{
		{Title: "task", Status: todolist.StatusPending},
	}))
	path := projectdir.SessionTodosPath(workDir, sessionID)
	require.FileExists(t, path)

	require.NoError(t, SaveTodosSnapshot(workDir, sessionID, nil))
	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err))
}

func TestLoadTodosReadsLegacyAppendOnlyFile(t *testing.T) {
	workDir := t.TempDir()
	path := projectdir.SessionTodosPath(workDir, "sess_legacy")
	require.NoError(t, projectdir.EnsureSessionMetadataDir(workDir, "sess_legacy"))
	legacy := "" +
		`{"time":"2026-01-01T00:00:00Z","session":"sess_legacy","todos":[{"title":"old","status":"pending"}]}` + "\n" +
		`{"time":"2026-01-02T00:00:00Z","session":"sess_legacy","todos":[{"title":"new","status":"done"}]}` + "\n"
	require.NoError(t, os.WriteFile(path, []byte(legacy), 0o644))

	loaded, err := LoadTodos(workDir, "sess_legacy")
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	require.Equal(t, "new", loaded[0].Title)
	require.Equal(t, todolist.StatusDone, loaded[0].Status)
}
