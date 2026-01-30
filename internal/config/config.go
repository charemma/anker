package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charemma/anker/internal/git"
	"github.com/charemma/anker/internal/paths"
	"github.com/charemma/anker/internal/timerange"
	"gopkg.in/yaml.v3"
)

// Config holds user configuration for anker.
type Config struct {
	WeekStart   string `yaml:"week_start"`   // "monday" or "sunday"
	AuthorEmail string `yaml:"author_email"` // default git author email for filtering
}

// DefaultConfig returns the default configuration.
// Attempts to read author email from git config.
func DefaultConfig() *Config {
	return &Config{
		WeekStart:   "monday",
		AuthorEmail: "", // will be set from git config if available
	}
}

// Load reads the configuration from $ANKER_HOME/config.yaml or ~/.anker/config.yaml.
// If the file doesn't exist, returns the default configuration.
// If author_email is not set, attempts to read from git config.
func Load() (*Config, error) {
	baseDir, err := paths.GetAnkerHome()
	if err != nil {
		return nil, fmt.Errorf("failed to get anker home directory: %w", err)
	}

	configPath := filepath.Join(baseDir, "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			// Try to get author email from git config
			if email, err := git.GetAuthorEmail(); err == nil && email != "" {
				cfg.AuthorEmail = email
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate week_start
	ws := strings.ToLower(config.WeekStart)
	if ws != "monday" && ws != "sunday" {
		return nil, fmt.Errorf("invalid week_start: %s (must be 'monday' or 'sunday')", config.WeekStart)
	}
	config.WeekStart = ws

	// If author_email is not set, try to get from git config
	if config.AuthorEmail == "" {
		if email, err := git.GetAuthorEmail(); err == nil && email != "" {
			config.AuthorEmail = email
		}
	}

	return &config, nil
}

// Save writes the configuration to $ANKER_HOME/config.yaml or ~/.anker/config.yaml.
func Save(config *Config) error {
	baseDir, err := paths.GetAnkerHome()
	if err != nil {
		return fmt.Errorf("failed to get anker home directory: %w", err)
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(baseDir, "config.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetTimerangeConfig converts the config to a timerange.Config.
func (c *Config) GetTimerangeConfig() *timerange.Config {
	weekStart := timerange.Monday
	if c.WeekStart == "sunday" {
		weekStart = timerange.Sunday
	}

	return &timerange.Config{
		WeekStart: weekStart,
	}
}
