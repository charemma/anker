package paths

import (
	"os"
	"path/filepath"
)

// GetAnkerHome returns the anker configuration directory.
// It checks the ANKER_HOME environment variable first,
// and falls back to ~/.anker if not set.
func GetAnkerHome() (string, error) {
	if home := os.Getenv("ANKER_HOME"); home != "" {
		return home, nil
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userHome, ".anker"), nil
}
