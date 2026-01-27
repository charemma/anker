package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatal(err)
	}

	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmd := range commands {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = repoPath
		if err := c.Run(); err != nil {
			t.Fatalf("failed to run %v: %v", cmd, err)
		}
	}

	return repoPath
}

func addCommit(t *testing.T, repoPath, message string) {
	t.Helper()
	addCommitWithDate(t, repoPath, message, "")
}

func addCommitWithDate(t *testing.T, repoPath, message, date string) {
	t.Helper()

	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte(message), 0644); err != nil {
		t.Fatal(err)
	}

	commands := [][]string{
		{"git", "add", "."},
	}

	if date != "" {
		commands = append(commands, []string{"git", "commit", "-m", message, "--date", date})
	} else {
		commands = append(commands, []string{"git", "commit", "-m", message})
	}

	for _, cmd := range commands {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = repoPath
		if err := c.Run(); err != nil {
			t.Fatalf("failed to run %v: %v", cmd, err)
		}
	}
}

func TestGitSource_Type(t *testing.T) {
	source := NewGitSource("/path/to/repo")
	if source.Type() != "git" {
		t.Errorf("expected type 'git', got %s", source.Type())
	}
}

func TestGitSource_Location(t *testing.T) {
	path := "/path/to/repo"
	source := NewGitSource(path)
	if source.Location() != path {
		t.Errorf("expected location %s, got %s", path, source.Location())
	}
}

func TestGitSource_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() string
		expectErr bool
	}{
		{
			name: "valid git repository",
			setup: func() string {
				return setupTestRepo(t)
			},
			expectErr: false,
		},
		{
			name: "non-existent directory",
			setup: func() string {
				return "/does/not/exist"
			},
			expectErr: true,
		},
		{
			name: "directory without .git",
			setup: func() string {
				tmpDir := t.TempDir()
				return tmpDir
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := tt.setup()
			source := NewGitSource(repoPath)

			err := source.Validate()
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGitSource_GetEntries(t *testing.T) {
	repoPath := setupTestRepo(t)

	addCommit(t, repoPath, "First commit")
	time.Sleep(10 * time.Millisecond)
	addCommit(t, repoPath, "Second commit")
	time.Sleep(10 * time.Millisecond)
	addCommit(t, repoPath, "Third commit")

	source := NewGitSource(repoPath)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(yesterday, now)
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	expectedMessages := []string{"Third commit", "Second commit", "First commit"}
	for i, entry := range entries {
		if entry.Source != "git" {
			t.Errorf("entry %d: expected source 'git', got %s", i, entry.Source)
		}

		if entry.Location != repoPath {
			t.Errorf("entry %d: expected location %s, got %s", i, repoPath, entry.Location)
		}

		if entry.Content != expectedMessages[i] {
			t.Errorf("entry %d: expected content %s, got %s", i, expectedMessages[i], entry.Content)
		}

		if entry.Metadata["author"] != "Test User" {
			t.Errorf("entry %d: expected author 'Test User', got %s", i, entry.Metadata["author"])
		}

		if entry.Metadata["email"] != "test@example.com" {
			t.Errorf("entry %d: expected email 'test@example.com', got %s", i, entry.Metadata["email"])
		}

		if entry.Metadata["hash"] == "" {
			t.Errorf("entry %d: hash should not be empty", i)
		}
	}
}

func TestGitSource_GetEntries_TimeRange(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Add old commit with explicit date in the past
	addCommitWithDate(t, repoPath, "Old commit", "2020-01-01 10:00:00")

	// Add recent commit
	addCommit(t, repoPath, "Recent commit")

	source := NewGitSource(repoPath)

	// Query only recent commits (last 24 hours)
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(yesterday, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		for i, e := range entries {
			t.Logf("Entry %d: %s (timestamp: %s)", i, e.Content, e.Timestamp)
		}
		t.Fatalf("expected 1 entry in time range, got %d", len(entries))
	}

	if entries[0].Content != "Recent commit" {
		t.Errorf("expected 'Recent commit', got %s", entries[0].Content)
	}
}

func TestGitSource_GetEntries_NoCommits(t *testing.T) {
	repoPath := setupTestRepo(t)
	source := NewGitSource(repoPath)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(yesterday, now)
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}
