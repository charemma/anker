package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetConfigDir returns the ikno configuration directory.
// Resolution order:
//  1. IKNO_HOME environment variable (explicit override)
//  2. $XDG_CONFIG_HOME/ikno/ if XDG_CONFIG_HOME is set
//  3. ~/.config/ikno/ (default)
//
// On first use, if ~/.anker/ exists and the target directory does not,
// the old directory is migrated automatically.
func GetConfigDir() (string, error) {
	if home := os.Getenv("IKNO_HOME"); home != "" {
		return home, nil
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configBase := os.Getenv("XDG_CONFIG_HOME")
	if configBase == "" {
		configBase = filepath.Join(userHome, ".config")
	}

	target := filepath.Join(configBase, "ikno")
	maybeMigrate(userHome, target)

	return target, nil
}

// maybeMigrate moves ~/.anker to target if the old path exists and target does not.
func maybeMigrate(userHome, target string) {
	oldPath := filepath.Join(userHome, ".anker")
	if _, err := os.Stat(target); err == nil {
		// target already exists, nothing to do
		return
	}
	if _, err := os.Stat(oldPath); err != nil {
		// old path does not exist either, nothing to migrate
		return
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return
	}
	if err := os.Rename(oldPath, target); err != nil {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "Migrated config from ~/.anker/ to ~/.config/ikno/\n")
}
