package sources

import (
	"os"
	"path/filepath"
	"testing"
)

// mkDir creates a directory relative to base and returns the full path.
func mkDir(t *testing.T, base, rel string) string {
	t.Helper()
	p := filepath.Join(base, rel)
	if err := os.MkdirAll(p, 0755); err != nil {
		t.Fatalf("mkDir(%s): %v", rel, err)
	}
	return p
}

// mkFile creates an empty file relative to base.
func mkFile(t *testing.T, base, rel string) {
	t.Helper()
	p := filepath.Join(base, rel)
	if err := os.WriteFile(p, nil, 0644); err != nil {
		t.Fatalf("mkFile(%s): %v", rel, err)
	}
}

func TestDetectType(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, dir string)
		wantTypes     []string
		wantNoResults bool
	}{
		{
			name: "git repo",
			setup: func(t *testing.T, dir string) {
				mkDir(t, dir, ".git")
			},
			wantTypes: []string{"git"},
		},
		{
			name: "obsidian vault",
			setup: func(t *testing.T, dir string) {
				mkDir(t, dir, ".obsidian")
			},
			wantTypes: []string{"obsidian"},
		},
		{
			name: "markdown directory (no .obsidian)",
			setup: func(t *testing.T, dir string) {
				mkFile(t, dir, "notes.md")
			},
			wantTypes: []string{"markdown"},
		},
		{
			name: "git repo with markdown files",
			setup: func(t *testing.T, dir string) {
				mkDir(t, dir, ".git")
				mkFile(t, dir, "README.md")
			},
			wantTypes: []string{"git", "markdown"},
		},
		{
			name: "obsidian vault with markdown files -- no markdown type",
			setup: func(t *testing.T, dir string) {
				mkDir(t, dir, ".obsidian")
				mkFile(t, dir, "notes.md")
			},
			wantTypes: []string{"obsidian"},
		},
		{
			name: "git repo and obsidian vault (both)",
			setup: func(t *testing.T, dir string) {
				mkDir(t, dir, ".git")
				mkDir(t, dir, ".obsidian")
			},
			wantTypes: []string{"git", "obsidian"},
		},
		{
			name:          "empty directory -- no match",
			setup:         func(t *testing.T, dir string) {},
			wantNoResults: true,
		},
		{
			name: "directory with only non-md files -- no markdown",
			setup: func(t *testing.T, dir string) {
				mkFile(t, dir, "main.go")
				mkFile(t, dir, "main.py")
			},
			wantNoResults: true,
		},
		{
			name: "md file uppercase extension",
			setup: func(t *testing.T, dir string) {
				mkFile(t, dir, "README.MD")
			},
			wantTypes: []string{"markdown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			got, err := DetectType(dir)
			if err != nil {
				t.Fatalf("DetectType error: %v", err)
			}

			if tt.wantNoResults {
				if len(got) != 0 {
					t.Errorf("want no results, got %v", got)
				}
				return
			}

			if len(got) != len(tt.wantTypes) {
				t.Errorf("want %d results, got %d: %v", len(tt.wantTypes), len(got), got)
				return
			}

			gotTypes := make(map[string]bool, len(got))
			for _, d := range got {
				gotTypes[d.Type] = true
				if d.Path == "" {
					t.Errorf("DetectedSource.Path is empty for type %s", d.Type)
				}
				if d.Reason == "" {
					t.Errorf("DetectedSource.Reason is empty for type %s", d.Type)
				}
			}

			for _, want := range tt.wantTypes {
				if !gotTypes[want] {
					t.Errorf("missing type %q in results %v", want, got)
				}
			}
		})
	}
}

func TestDetectType_ClaudePath(t *testing.T) {
	t.Run("directory with .claude/projects child", func(t *testing.T) {
		dir := t.TempDir()
		mkDir(t, dir, filepath.Join(".claude", "projects"))

		got, err := DetectType(dir)
		if err != nil {
			t.Fatalf("DetectType error: %v", err)
		}

		found := false
		for _, d := range got {
			if d.Type == "claude" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected claude type, got %v", got)
		}
	})
}

func TestDetectType_NonExistentPath(t *testing.T) {
	// Should return an error for non-existent paths because filepath.Abs succeeds
	// but the directory itself won't yield any matches (isDir returns false).
	dir := t.TempDir()
	nonExistent := filepath.Join(dir, "does-not-exist")

	got, err := DetectType(nonExistent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No results expected for a non-existent directory.
	if len(got) != 0 {
		t.Errorf("expected no results for non-existent path, got %v", got)
	}
}

func TestDiscoverSources(t *testing.T) {
	t.Run("finds git repos in direct children", func(t *testing.T) {
		root := t.TempDir()
		mkDir(t, root, filepath.Join("repo1", ".git"))
		mkDir(t, root, filepath.Join("repo2", ".git"))
		mkDir(t, root, "empty")

		got, err := DiscoverSources(root, 1, nil)
		if err != nil {
			t.Fatalf("DiscoverSources error: %v", err)
		}

		gitCount := 0
		for _, d := range got {
			if d.Type == "git" {
				gitCount++
			}
		}
		if gitCount != 2 {
			t.Errorf("expected 2 git repos, got %d (all: %v)", gitCount, got)
		}
	})

	t.Run("skips already registered paths", func(t *testing.T) {
		root := t.TempDir()
		repo := mkDir(t, root, "repo1")
		mkDir(t, root, filepath.Join("repo1", ".git"))
		mkDir(t, root, filepath.Join("repo2", ".git"))

		registered := []Config{{Type: "git", Path: repo}}

		got, err := DiscoverSources(root, 1, registered)
		if err != nil {
			t.Fatalf("DiscoverSources error: %v", err)
		}

		for _, d := range got {
			absGot, _ := filepath.Abs(d.Path)
			absReg, _ := filepath.Abs(repo)
			if absGot == absReg {
				t.Errorf("returned already-registered path: %s", d.Path)
			}
		}
	})

	t.Run("depth 2 finds nested repos", func(t *testing.T) {
		root := t.TempDir()
		mkDir(t, root, filepath.Join("org", "repo1", ".git"))
		mkDir(t, root, filepath.Join("org", "repo2", ".git"))

		got, err := DiscoverSources(root, 2, nil)
		if err != nil {
			t.Fatalf("DiscoverSources error: %v", err)
		}

		gitCount := 0
		for _, d := range got {
			if d.Type == "git" {
				gitCount++
			}
		}
		if gitCount != 2 {
			t.Errorf("expected 2 git repos at depth 2, got %d (all: %v)", gitCount, got)
		}
	})

	t.Run("depth 1 does not find nested repos", func(t *testing.T) {
		root := t.TempDir()
		mkDir(t, root, filepath.Join("org", "repo1", ".git"))

		got, err := DiscoverSources(root, 1, nil)
		if err != nil {
			t.Fatalf("DiscoverSources error: %v", err)
		}

		for _, d := range got {
			if d.Type == "git" {
				t.Errorf("depth 1 should not find nested repo, found: %s", d.Path)
			}
		}
	})

	t.Run("returns error for unreadable root dir", func(t *testing.T) {
		_, err := DiscoverSources("/nonexistent-path-that-cannot-exist", 1, nil)
		if err == nil {
			t.Error("expected error for unreadable root dir")
		}
	})

	t.Run("skips non-directories", func(t *testing.T) {
		root := t.TempDir()
		mkFile(t, root, "somefile.txt")
		mkDir(t, root, filepath.Join("repo", ".git"))

		got, err := DiscoverSources(root, 1, nil)
		if err != nil {
			t.Fatalf("DiscoverSources error: %v", err)
		}

		gitCount := 0
		for _, d := range got {
			if d.Type == "git" {
				gitCount++
			}
		}
		if gitCount != 1 {
			t.Errorf("expected 1 git repo, got %d", gitCount)
		}
	})

	t.Run("empty dir returns no results", func(t *testing.T) {
		root := t.TempDir()

		got, err := DiscoverSources(root, 1, nil)
		if err != nil {
			t.Fatalf("DiscoverSources error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected no results, got %v", got)
		}
	})
}

func TestBuildRegisteredSet(t *testing.T) {
	dir := t.TempDir()
	abs := dir // t.TempDir() already returns an absolute path

	configs := []Config{
		{Type: "git", Path: abs},
		{Type: "markdown", Path: filepath.Join(abs, "notes")},
	}

	set := buildRegisteredSet(configs)

	if !set[abs] {
		t.Errorf("expected %q in registered set", abs)
	}
	notesAbs, _ := filepath.Abs(filepath.Join(abs, "notes"))
	if !set[notesAbs] {
		t.Errorf("expected %q in registered set", notesAbs)
	}
}
