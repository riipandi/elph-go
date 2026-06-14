package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

func TestReadNonRepo(t *testing.T) {
	dir := t.TempDir()
	st := Read(dir)
	require.Equal(t, "—", st.Branch)
	require.False(t, st.IsRepo)
}

func TestReadBranchAndDiffStats(t *testing.T) {
	dir := t.TempDir()
	initRepoWithChanges(t, dir)

	st := Read(dir)
	require.True(t, st.IsRepo)
	require.Equal(t, "master", st.Branch)
	require.Positive(t, st.Added)
}

func initRepoWithChanges(t *testing.T, dir string) {
	t.Helper()

	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello\n"), 0o644))
	_, err = wt.Add("a.txt")
	require.NoError(t, err)

	_, err = wt.Commit("init", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello\nworld\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("new\n"), 0o644))
	_, err = wt.Add("b.txt")
	require.NoError(t, err)
}

func TestChangedPathCount(t *testing.T) {
	status := git.Status{
		"a.txt": {Staging: git.Modified},
		"b.txt": {Worktree: git.Modified},
		"c.txt": {Staging: git.Unmodified, Worktree: git.Unmodified},
	}
	require.Equal(t, 2, changedPathCount(status))
}

func TestCountTextDiff(t *testing.T) {
	added, deleted := countTextDiff("one\n", "one\ntwo\n")
	require.Equal(t, 1, added)
	require.Equal(t, 0, deleted)

	added, deleted = countTextDiff("one\ntwo\n", "one\n")
	require.Equal(t, 0, added)
	require.Equal(t, 1, deleted)
}
