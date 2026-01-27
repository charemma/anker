package git

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNotInRepo = errors.New("not in a git repository")

// FindRepoRoot walks up the directory tree from the given path
// to find the git repository root (directory containing .git).
func FindRepoRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	current := absPath
	for {
		gitDir := filepath.Join(current, ".git")
		info, err := os.Stat(gitDir)
		if err == nil && info.IsDir() {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", ErrNotInRepo
		}
		current = parent
	}
}
