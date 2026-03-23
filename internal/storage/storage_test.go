package storage

import (
	"os"
	"testing"
)

func TestNewStore(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("ANKER_HOME", tmpDir)

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
