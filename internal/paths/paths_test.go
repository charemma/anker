package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	t.Run("uses IKNO_HOME env var when set", func(t *testing.T) {
		customPath := "/custom/ikno/path"
		t.Setenv("IKNO_HOME", customPath)
		t.Setenv("XDG_CONFIG_HOME", "")

		dir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dir != customPath {
			t.Errorf("expected %s, got %s", customPath, dir)
		}
	})

	t.Run("uses XDG_CONFIG_HOME when set", func(t *testing.T) {
		xdgBase := t.TempDir()
		t.Setenv("IKNO_HOME", "")
		t.Setenv("XDG_CONFIG_HOME", xdgBase)

		dir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := filepath.Join(xdgBase, "ikno")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})

	t.Run("falls back to ~/.config/ikno when no env vars set", func(t *testing.T) {
		t.Setenv("IKNO_HOME", "")
		t.Setenv("XDG_CONFIG_HOME", "")

		dir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		userHome, _ := os.UserHomeDir()
		expected := filepath.Join(userHome, ".config", "ikno")

		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})
}

func TestMaybeMigrate(t *testing.T) {
	t.Run("migrates ~/.anker to target when target does not exist", func(t *testing.T) {
		tmp := t.TempDir()
		oldPath := filepath.Join(tmp, ".anker")
		target := filepath.Join(tmp, ".config", "ikno")

		if err := os.MkdirAll(oldPath, 0755); err != nil {
			t.Fatal(err)
		}
		// Write a file to confirm migration
		testFile := filepath.Join(oldPath, "sources.yaml")
		if err := os.WriteFile(testFile, []byte("sources: []\n"), 0644); err != nil {
			t.Fatal(err)
		}

		maybeMigrate(tmp, target)

		if _, err := os.Stat(oldPath); err == nil {
			t.Error("old path should have been removed after migration")
		}
		if _, err := os.Stat(filepath.Join(target, "sources.yaml")); err != nil {
			t.Error("migrated file should exist at new location")
		}
	})

	t.Run("does not migrate when target already exists", func(t *testing.T) {
		tmp := t.TempDir()
		oldPath := filepath.Join(tmp, ".anker")
		target := filepath.Join(tmp, ".config", "ikno")

		if err := os.MkdirAll(oldPath, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(target, 0755); err != nil {
			t.Fatal(err)
		}

		maybeMigrate(tmp, target)

		// old path should still exist since target was already there
		if _, err := os.Stat(oldPath); err != nil {
			t.Error("old path should still exist when target already exists")
		}
	})
}
