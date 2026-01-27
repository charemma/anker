package git

import (
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
				if err != ErrNotInRepo {
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

			gitDir := filepath.Join(result, ".git")
			info, err := os.Stat(gitDir)
			if err != nil {
				t.Errorf(".git directory not found at %s: %v", gitDir, err)
				return
			}

			if !info.IsDir() {
				t.Errorf(".git exists but is not a directory")
			}
		})
	}
}
