package git

import (
	"io"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	gitdiff "github.com/go-git/go-git/v5/utils/diff"
	dmp "github.com/sergi/go-diff/diffmatchpatch"
)

// Status summarizes the current git branch and unstaged/staged line changes.
type Status struct {
	Branch  string
	Added   int
	Deleted int
	IsRepo  bool
}

// Read returns git metadata for workDir. Non-git directories use Branch "—".
func Read(workDir string) Status {
	repo, err := git.PlainOpen(workDir)
	if err != nil {
		return Status{Branch: "—"}
	}

	wt, err := repo.Worktree()
	if err != nil {
		return Status{Branch: "—"}
	}

	branch := resolveBranch(repo)
	added, deleted := lineStats(repo, wt)
	return Status{
		Branch:  branch,
		Added:   added,
		Deleted: deleted,
		IsRepo:  true,
	}
}

func resolveBranch(repo *git.Repository) string {
	ref, err := repo.Head()
	if err != nil {
		return "—"
	}
	if ref.Name() == plumbing.HEAD {
		return "detached"
	}
	branch := ref.Name().Short()
	if branch == "" {
		return "—"
	}
	return branch
}

func lineStats(repo *git.Repository, wt *git.Worktree) (added, deleted int) {
	status, err := wt.Status()
	if err != nil {
		return 0, 0
	}

	for path, fs := range status {
		if fs.Staging != git.Unmodified && fs.Staging != git.Untracked {
			a, d := stagedPathStats(repo, path, fs.Staging)
			added += a
			deleted += d
		}
		if fs.Worktree != git.Unmodified && fs.Worktree != git.Untracked {
			a, d := unstagedPathStats(repo, wt, path, fs.Worktree)
			added += a
			deleted += d
		}
	}
	return added, deleted
}

func stagedPathStats(repo *git.Repository, path string, code git.StatusCode) (added, deleted int) {
	switch code {
	case git.Added:
		content, ok := readIndexBytes(repo, path)
		if ok {
			return countTextLines(string(content)), 0
		}
	case git.Deleted:
		content, ok := readHeadBytes(repo, path)
		if ok {
			return 0, countTextLines(content)
		}
	case git.Modified:
		from, fromOK := readHeadBytes(repo, path)
		to, toOK := readIndexBytes(repo, path)
		return diffText(from, fromOK, string(to), toOK)
	}
	return 0, 0
}

func unstagedPathStats(repo *git.Repository, wt *git.Worktree, path string, code git.StatusCode) (added, deleted int) {
	switch code {
	case git.Deleted:
		content, ok := readIndexBytes(repo, path)
		if ok {
			return 0, countTextLines(string(content))
		}
	case git.Modified:
		from, fromOK := readIndexBytes(repo, path)
		to, toOK := readWorktreeBytes(wt, path)
		return diffText(string(from), fromOK, string(to), toOK)
	}
	return 0, 0
}

func diffText(from string, fromOK bool, to string, toOK bool) (added, deleted int) {
	switch {
	case fromOK && toOK:
		return countTextDiff(from, to)
	case !fromOK && toOK:
		return countTextLines(to), 0
	case fromOK && !toOK:
		return 0, countTextLines(from)
	default:
		return 0, 0
	}
}

func readHeadBytes(repo *git.Repository, path string) (string, bool) {
	tree, err := headTree(repo)
	if err != nil || tree == nil {
		return "", false
	}
	file, err := tree.File(path)
	if err != nil {
		return "", false
	}
	content, err := file.Contents()
	if err != nil {
		return "", false
	}
	return content, true
}

func headTree(repo *git.Repository) (*object.Tree, error) {
	ref, err := repo.Head()
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			return nil, nil
		}
		return nil, err
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}
	return commit.Tree()
}

func readIndexBytes(repo *git.Repository, path string) ([]byte, bool) {
	idx, err := repo.Storer.Index()
	if err != nil {
		return nil, false
	}
	entry, err := idx.Entry(path)
	if err != nil {
		return nil, false
	}
	return readBlobBytes(repo.Storer, entry.Hash)
}

func readWorktreeBytes(wt *git.Worktree, path string) ([]byte, bool) {
	f, err := wt.Filesystem.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, false
	}
	return data, true
}

func readBlobBytes(store storer.EncodedObjectStorer, hash plumbing.Hash) ([]byte, bool) {
	if hash.IsZero() {
		return nil, false
	}
	blob, err := object.GetBlob(store, hash)
	if err != nil {
		return nil, false
	}
	reader, err := blob.Reader()
	if err != nil {
		return nil, false
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, false
	}
	return data, true
}

func countTextDiff(from, to string) (added, deleted int) {
	for _, d := range gitdiff.Do(from, to) {
		lines := countTextLines(d.Text)
		switch d.Type {
		case dmp.DiffInsert:
			added += lines
		case dmp.DiffDelete:
			deleted += lines
		}
	}
	return added, deleted
}

func countTextLines(text string) int {
	if text == "" {
		return 0
	}
	lines := strings.Count(text, "\n")
	if !strings.HasSuffix(text, "\n") {
		lines++
	}
	return lines
}
