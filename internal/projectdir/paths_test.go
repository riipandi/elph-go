package projectdir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootAndSessionPaths(t *testing.T) {
	workDir := "/tmp/repo"
	require.Equal(t, filepath.Join(workDir, ".agents", "elph"), Root(workDir))
	require.Equal(t, filepath.Join(workDir, ".agents", "elph", "prompts"), PromptsDir(workDir))
	require.Equal(t, filepath.Join(workDir, ".agents", "elph", "skills"), SkillsDir(workDir))
	require.Equal(t, filepath.Join(workDir, ".agents", "elph", "metadata"), MetadataDir(workDir))
	require.Equal(t, filepath.Join(workDir, ".agents", "elph", "metadata", "sess_01"), SessionMetadataDir(workDir, "sess_01"))
	require.Equal(t, filepath.Join(workDir, ".agents", "elph", "metadata", "sess_01", "todos.jsonl"), SessionTodosPath(workDir, "sess_01"))
}

func TestEnsureRootWritesGitignore(t *testing.T) {
	workDir := t.TempDir()
	require.NoError(t, EnsureRoot(workDir))

	gitignore := filepath.Join(Root(workDir), ".gitignore")
	raw, err := os.ReadFile(gitignore)
	require.NoError(t, err)
	for _, entry := range gitignoreRequiredEntries {
		require.Contains(t, string(raw), entry)
	}
	require.Contains(t, string(raw), "metadata/")

	require.NoError(t, EnsureRoot(workDir))
	mod, err := os.Stat(gitignore)
	require.NoError(t, err)
	require.NoError(t, EnsureRoot(workDir))
	mod2, err := os.Stat(gitignore)
	require.NoError(t, err)
	require.Equal(t, mod.ModTime(), mod2.ModTime())
}

func TestEnsureRootUpgradesStaleGitignore(t *testing.T) {
	workDir := t.TempDir()
	root := Root(workDir)
	require.NoError(t, os.MkdirAll(root, 0o755))
	gitignore := filepath.Join(root, ".gitignore")
	require.NoError(t, os.WriteFile(gitignore, []byte("logs/\n"), 0o644))

	require.NoError(t, EnsureRoot(workDir))

	raw, err := os.ReadFile(gitignore)
	require.NoError(t, err)
	for _, entry := range gitignoreRequiredEntries {
		require.Contains(t, string(raw), entry)
	}
}

func TestEnsureSessionMetadataDir(t *testing.T) {
	workDir := t.TempDir()
	require.NoError(t, EnsureSessionMetadataDir(workDir, "sess_01"))
	require.DirExists(t, SessionMetadataDir(workDir, "sess_01"))
}
