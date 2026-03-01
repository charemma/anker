package obsidian

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestObsidianSource_Validate(t *testing.T) {
	// Create temp vault
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test-vault")
	if err := os.Mkdir(vaultPath, 0755); err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	source := NewObsidianSource(vaultPath)

	// Should fail without .obsidian folder
	if err := source.Validate(); err == nil {
		t.Error("expected validation to fail without .obsidian folder")
	}

	// Create .obsidian folder
	if err := os.Mkdir(filepath.Join(vaultPath, ".obsidian"), 0755); err != nil {
		t.Fatalf("failed to create .obsidian: %v", err)
	}

	// Should succeed now
	if err := source.Validate(); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}

func TestObsidianSource_GetEntries(t *testing.T) {
	// Create temp vault
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test-vault")
	if err := os.Mkdir(vaultPath, 0755); err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}
	if err := os.Mkdir(filepath.Join(vaultPath, ".obsidian"), 0755); err != nil {
		t.Fatalf("failed to create .obsidian: %v", err)
	}
	if err := os.Mkdir(filepath.Join(vaultPath, "Daily Notes"), 0755); err != nil {
		t.Fatalf("failed to create Daily Notes: %v", err)
	}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)

	// Create test files with different timestamps
	files := map[string]time.Time{
		"note1.md":                  now,
		"note2.md":                  yesterday,
		"Daily Notes/2025-01-29.md": now,
		".obsidian/workspace.json":  now, // should be skipped
		"old-note.md":               lastWeek,
	}

	for name, modTime := range files {
		path := filepath.Join(vaultPath, name)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("failed to change times for %s: %v", name, err)
		}
	}

	source := NewObsidianSource(vaultPath)

	// Get entries from yesterday to now
	from := yesterday.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	entries, err := source.GetEntries(from, to)
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	// Should find 3 files (note1, note2, Daily Notes/2025-01-29)
	// old-note is outside range, workspace.json is in .obsidian
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
		for _, e := range entries {
			t.Logf("  - %s", e.Content)
		}
	}

	// Check that entries have correct metadata
	for _, entry := range entries {
		if entry.Source != "obsidian" {
			t.Errorf("expected source 'obsidian', got '%s'", entry.Source)
		}
		if entry.Location != vaultPath {
			t.Errorf("expected location '%s', got '%s'", vaultPath, entry.Location)
		}
		if entry.Metadata["path"] == "" {
			t.Error("entry missing 'path' metadata")
		}
		if entry.Metadata["action"] != "modified" && entry.Metadata["action"] != "created" {
			t.Errorf("invalid action: %s", entry.Metadata["action"])
		}
	}
}

func TestObsidianSource_SkipHiddenFolders(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test-vault")
	if err := os.Mkdir(vaultPath, 0755); err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}
	if err := os.Mkdir(filepath.Join(vaultPath, ".obsidian"), 0755); err != nil {
		t.Fatalf("failed to create .obsidian: %v", err)
	}
	if err := os.Mkdir(filepath.Join(vaultPath, ".trash"), 0755); err != nil {
		t.Fatalf("failed to create .trash: %v", err)
	}

	now := time.Now()

	// Create files in hidden folders
	if err := os.WriteFile(filepath.Join(vaultPath, ".obsidian", "config.md"), []byte("config"), 0644); err != nil {
		t.Fatalf("failed to write config.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultPath, ".trash", "deleted.md"), []byte("deleted"), 0644); err != nil {
		t.Fatalf("failed to write deleted.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultPath, "visible.md"), []byte("visible"), 0644); err != nil {
		t.Fatalf("failed to write visible.md: %v", err)
	}

	if err := os.Chtimes(filepath.Join(vaultPath, ".obsidian", "config.md"), now, now); err != nil {
		t.Fatalf("failed to change times for config.md: %v", err)
	}
	if err := os.Chtimes(filepath.Join(vaultPath, ".trash", "deleted.md"), now, now); err != nil {
		t.Fatalf("failed to change times for deleted.md: %v", err)
	}
	if err := os.Chtimes(filepath.Join(vaultPath, "visible.md"), now, now); err != nil {
		t.Fatalf("failed to change times for visible.md: %v", err)
	}

	source := NewObsidianSource(vaultPath)

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	entries, err := source.GetEntries(from, to)
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	// Should only find visible.md
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	if len(entries) > 0 && entries[0].Metadata["file"] != "visible.md" {
		t.Errorf("expected visible.md, got %s", entries[0].Metadata["file"])
	}
}
