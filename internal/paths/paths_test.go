package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetAnkerHome(t *testing.T) {
	t.Run("uses ANKER_HOME env var when set", func(t *testing.T) {
		customPath := "/custom/anker/path"
		os.Setenv("ANKER_HOME", customPath)
		defer os.Unsetenv("ANKER_HOME")

		home, err := GetAnkerHome()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if home != customPath {
			t.Errorf("expected %s, got %s", customPath, home)
		}
	})

	t.Run("falls back to ~/.anker when ANKER_HOME not set", func(t *testing.T) {
		os.Unsetenv("ANKER_HOME")

		home, err := GetAnkerHome()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		userHome, _ := os.UserHomeDir()
		expected := filepath.Join(userHome, ".anker")

		if home != expected {
			t.Errorf("expected %s, got %s", expected, home)
		}
	})
}
