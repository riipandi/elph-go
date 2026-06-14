package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadBranch returns the current branch name without opening the repository via
// go-git. Added and Deleted are always zero; use Read for line stats.
func ReadBranch(workDir string) Status {
	gitDir, err := resolveGitDir(workDir)
	if err != nil {
		return Status{Branch: "—"}
	}
	return Status{Branch: parseHEAD(gitDir), IsRepo: true}
}

func resolveGitDir(workDir string) (string, error) {
	gitPath := filepath.Join(workDir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return gitPath, nil
	}

	data, err := os.ReadFile(gitPath)
	if err != nil {
		return "", err
	}
	const prefix = "gitdir: "
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf("invalid .git file")
	}
	dir := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(workDir, dir)
	}
	return dir, nil
}

func parseHEAD(gitDir string) string {
	data, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return "—"
	}
	line := strings.TrimSpace(string(data))
	if strings.HasPrefix(line, "ref: ") {
		ref := strings.TrimPrefix(line, "ref: ")
		if branch, ok := strings.CutPrefix(ref, "refs/heads/"); ok && branch != "" {
			return branch
		}
		return "—"
	}
	if len(line) >= 7 {
		return "detached"
	}
	return "—"
}
