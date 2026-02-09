package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("ANKER_HOME", tmpDir)
	defer os.Unsetenv("ANKER_HOME")

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() failed: %v", err)
	}

	if store.baseDir == "" {
		t.Error("baseDir should not be empty")
	}

	if store.baseDir != tmpDir {
		t.Errorf("expected baseDir %s, got %s", tmpDir, store.baseDir)
	}

	info, err := os.Stat(tmpDir)
	if err != nil {
		t.Errorf("base directory not accessible: %v", err)
	}
	if !info.IsDir() {
		t.Error("baseDir is not a directory")
	}
}

func TestTrackRepo(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{baseDir: tmpDir}

	repoPath := "/path/to/test/repo"

	err := store.TrackRepo(repoPath)
	if err != nil {
		t.Fatalf("TrackRepo() failed: %v", err)
	}

	registryPath := filepath.Join(tmpDir, "repos.yaml")
	if _, err := os.Stat(registryPath); err != nil {
		t.Errorf("repos.yaml not created: %v", err)
	}

	registry, err := store.loadRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	if len(registry.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(registry.Repos))
	}

	if registry.Repos[0].Path != repoPath {
		t.Errorf("expected path %s, got %s", repoPath, registry.Repos[0].Path)
	}

	if registry.Repos[0].LastSeen.IsZero() {
		t.Error("LastSeen should not be zero")
	}
}

func TestTrackRepoUpdatesTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{baseDir: tmpDir}

	repoPath := "/path/to/test/repo"

	err := store.TrackRepo(repoPath)
	if err != nil {
		t.Fatalf("first TrackRepo() failed: %v", err)
	}

	registryPath := filepath.Join(tmpDir, "repos.yaml")
	registry1, err := store.loadRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	firstTimestamp := registry1.Repos[0].LastSeen

	time.Sleep(10 * time.Millisecond)

	err = store.TrackRepo(repoPath)
	if err != nil {
		t.Fatalf("second TrackRepo() failed: %v", err)
	}

	registry2, err := store.loadRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry after update: %v", err)
	}

	if len(registry2.Repos) != 1 {
		t.Fatalf("expected 1 repo after update, got %d", len(registry2.Repos))
	}

	secondTimestamp := registry2.Repos[0].LastSeen

	if !secondTimestamp.After(firstTimestamp) {
		t.Errorf("timestamp should be updated, first: %v, second: %v", firstTimestamp, secondTimestamp)
	}
}

func TestTrackMultipleRepos(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{baseDir: tmpDir}

	repos := []string{
		"/path/to/repo1",
		"/path/to/repo2",
		"/path/to/repo3",
	}

	for _, repo := range repos {
		if err := store.TrackRepo(repo); err != nil {
			t.Fatalf("TrackRepo(%s) failed: %v", repo, err)
		}
	}

	registryPath := filepath.Join(tmpDir, "repos.yaml")
	registry, err := store.loadRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	if len(registry.Repos) != len(repos) {
		t.Fatalf("expected %d repos, got %d", len(repos), len(registry.Repos))
	}

	for i, repo := range repos {
		if registry.Repos[i].Path != repo {
			t.Errorf("repo %d: expected path %s, got %s", i, repo, registry.Repos[i].Path)
		}
	}
}

func TestLoadRegistryNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{baseDir: tmpDir}

	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.yaml")
	registry, err := store.loadRegistry(nonExistentPath)

	if err == nil {
		t.Error("expected error for non-existent file")
	}

	if registry == nil {
		t.Error("registry should not be nil even on error")
		return
	}

	if len(registry.Repos) != 0 {
		t.Errorf("expected empty repos on error, got %d", len(registry.Repos))
	}
}
