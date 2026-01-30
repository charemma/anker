package git

import (
	"os/exec"
	"strings"
)

// GetGlobalConfig reads a value from global git config.
func GetGlobalConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--global", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetAuthorEmail returns the configured git user email.
func GetAuthorEmail() (string, error) {
	return GetGlobalConfig("user.email")
}

// GetAuthorName returns the configured git user name.
func GetAuthorName() (string, error) {
	return GetGlobalConfig("user.name")
}
