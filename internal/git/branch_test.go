package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadBranchNonRepo(t *testing.T) {
	dir := t.TempDir()
	st := ReadBranch(dir)
	require.Equal(t, "—", st.Branch)
	require.False(t, st.IsRepo)
	require.Zero(t, st.Added)
	require.Zero(t, st.Deleted)
}

func TestReadBranchFromRepo(t *testing.T) {
	dir := t.TempDir()
	initRepoWithChanges(t, dir)

	st := ReadBranch(dir)
	require.True(t, st.IsRepo)
	require.Equal(t, "master", st.Branch)
	require.Zero(t, st.Added)
}

func TestReadBranchWorktreeGitFile(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, "actual-git")
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: "+gitDir+"\n"), 0o644))

	st := ReadBranch(dir)
	require.True(t, st.IsRepo)
	require.Equal(t, "main", st.Branch)
}
