package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoRoot(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string
		expectErr bool
		wantPath  string
	}{
		{
			name: "finds repo root in git directory",
			setup: func() string {
				repoRoot := filepath.Join(tmpDir, "myrepo")
				gitDir := filepath.Join(repoRoot, ".git")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				return repoRoot
			},
			expectErr: false,
		},
		{
			name: "finds repo root from subdirectory",
			setup: func() string {
				repoRoot := filepath.Join(tmpDir, "project")
				gitDir := filepath.Join(repoRoot, ".git")
				subDir := filepath.Join(repoRoot, "src", "internal", "deep")
				if err := os.MkdirAll(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatal(err)
				}
				return subDir
			},
			expectErr: false,
		},
		{
			name: "finds bare git repo",
			setup: func() string {
				bareRepo := filepath.Join(tmpDir, "bare.git")
				if err := os.MkdirAll(filepath.Join(bareRepo, "objects"), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(bareRepo, "HEAD"), []byte("ref: refs/heads/main\n"), 0644); err != nil {
					t.Fatal(err)
				}
				return bareRepo
			},
			expectErr: false,
		},
		{
			name: "returns error when not in git repo",
			setup: func() string {
				noGitDir := filepath.Join(tmpDir, "notarepo")
				if err := os.MkdirAll(noGitDir, 0755); err != nil {
					t.Fatal(err)
				}
				return noGitDir
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startPath := tt.setup()

			result, err := FindRepoRoot(startPath)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !errors.Is(err, ErrNotInRepo) {
					t.Errorf("expected ErrNotInRepo, got %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Error("expected non-empty result")
				return
			}

			// Verify result is a git repo: either has .git/ dir or is a bare repo (HEAD file + objects/ dir)
			gitDir := filepath.Join(result, ".git")
			info, statErr := os.Stat(gitDir)
			hasDotGit := statErr == nil && info.IsDir()

			headInfo, headErr := os.Stat(filepath.Join(result, "HEAD"))
			hasBare := headErr == nil && !headInfo.IsDir()

			if !hasDotGit && !hasBare {
				t.Errorf("result %s is neither a regular git repo (.git/) nor a bare repo (HEAD)", result)
			}
		})
	}
}
