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
	WeekStart    string `yaml:"week_start"`             // "monday" or "sunday"
	AuthorEmail  string `yaml:"author_email,omitempty"` // default git author email for filtering
	AIBaseURL    string `yaml:"ai_base_url"`            // OpenAI-compatible API base URL
	AIModel      string `yaml:"ai_model"`               // model name for AI summaries
	AIAPIKey     string `yaml:"ai_api_key"`             // API key (prefer env var AI_API_KEY)
	AIPrompt     string `yaml:"ai_prompt"`              // custom prompt for AI summaries
	AIBackend    string `yaml:"ai_backend"`             // "api" or "cli"
	AICLICommand string `yaml:"ai_cli_command"`         // CLI tool for ai_backend: cli
}

// DefaultConfig returns the default configuration.
// Attempts to read author email from git config.
func DefaultConfig() *Config {
	return &Config{
		WeekStart:    "monday",
		AuthorEmail:  "", // will be set from git config if available
		AIBaseURL:    "https://api.anthropic.com/v1/",
		AIModel:      "claude-sonnet-4-20250514",
		AIBackend:    "api",
		AICLICommand: "claude -p",
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

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate week_start
	ws := strings.ToLower(cfg.WeekStart)
	if ws != "monday" && ws != "sunday" {
		return nil, fmt.Errorf("invalid week_start: %s (must be 'monday' or 'sunday')", cfg.WeekStart)
	}
	cfg.WeekStart = ws

	// Validate ai_backend
	if cfg.AIBackend != "" {
		ab := strings.ToLower(cfg.AIBackend)
		if ab != "api" && ab != "cli" {
			return nil, fmt.Errorf("invalid ai_backend: %s (must be 'api' or 'cli')", cfg.AIBackend)
		}
		cfg.AIBackend = ab
	}

	// If author_email is not set, try to get from git config
	if cfg.AuthorEmail == "" {
		if email, err := git.GetAuthorEmail(); err == nil && email != "" {
			cfg.AuthorEmail = email
		}
	}

	return cfg, nil
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	baseDir, err := paths.GetAnkerHome()
	if err != nil {
		return "", fmt.Errorf("failed to get anker home directory: %w", err)
	}
	return filepath.Join(baseDir, "config.yaml"), nil
}

// configTemplate is written when creating a new config file, so the user
// can see all available options with commented-out examples.
const configTemplate = `# anker configuration
# See: https://github.com/charemma/anker

# Week start day for "thisweek"/"lastweek" timespecs
week_start: monday

# Default git author email for filtering commits
# author_email: you@example.com

# AI summary settings for "anker recap --format ai"
# Supports any OpenAI-compatible API endpoint.
#
# Providers:
#   Anthropic:  https://api.anthropic.com/v1/
#   OpenAI:     https://api.openai.com/v1/
#   ollama:     http://localhost:11434/v1/
#   vllm:       http://localhost:8000/v1/
#
# ai_base_url: https://api.anthropic.com/v1/
# ai_model: claude-sonnet-4-20250514
# ai_api_key: sk-...                  # or set AI_API_KEY env var
# ai_prompt: "Summarize my workday."  # override default summary prompt
#
# AI backend: "api" (default) calls an OpenAI-compatible API directly,
# "cli" pipes recap data into a CLI tool instead.
# ai_backend: api
# ai_cli_command: claude -p            # CLI tool for ai_backend: cli
`

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

// EnsureConfigFile creates the config file with commented defaults if it doesn't exist.
// Returns the path to the config file.
func EnsureConfigFile() (string, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	baseDir := filepath.Dir(configPath)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(configTemplate), 0644); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
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
