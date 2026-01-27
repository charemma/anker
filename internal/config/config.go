package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charemma/anker/internal/timerange"
	"gopkg.in/yaml.v3"
)

// Config holds user configuration for anker.
type Config struct {
	WeekStart string `yaml:"week_start"` // "monday" or "sunday"
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		WeekStart: "monday",
	}
}

// Load reads the configuration from ~/.anker/config.yaml.
// If the file doesn't exist, returns the default configuration.
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".anker", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
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

	return &config, nil
}

// Save writes the configuration to ~/.anker/config.yaml.
func Save(config *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(home, ".anker")
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
