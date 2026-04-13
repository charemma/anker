package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charemma/anker/internal/ai"
	"github.com/charemma/anker/internal/config"
	"github.com/charemma/anker/internal/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit configuration",
	Long: `Open the anker configuration file in your editor.

Creates the config file with commented defaults if it doesn't exist yet.
Editor is resolved from $VISUAL, $EDITOR, or falls back to vi.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.EnsureConfigFile()
		if err != nil {
			return err
		}

		editor := resolveEditor()
		editorCmd := exec.Command(editor, configPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		return editorCmd.Run()
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print path to config file",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.ConfigPath()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, configPath)
		return nil
	},
}

// validConfigKeys lists all settable config keys with their YAML names.
var validConfigKeys = []string{
	"week_start", "author_email",
	"ai_backend", "ai_cli_command", "ai_base_url", "ai_model", "ai_api_key", "ai_prompt",
	"ai_default_style",
}

var configSetCmd = &cobra.Command{
	Use:       "set <key> <value>",
	Short:     "Set a configuration value",
	Args:      cobra.ExactArgs(2),
	ValidArgs: validConfigKeys,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := applyConfigKey(cfg, key, value); err != nil {
			return err
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		_, _ = fmt.Fprintln(os.Stdout, ui.StyleSuccess.Render("Set "+key+" = "+value))
		return nil
	},
}

// applyConfigKey sets the named config field to value, returning an error for
// unknown keys or invalid values.
func applyConfigKey(cfg *config.Config, key, value string) error {
	switch key {
	case "week_start":
		v := strings.ToLower(value)
		if v != "monday" && v != "sunday" {
			return fmt.Errorf("week_start must be 'monday' or 'sunday'")
		}
		cfg.WeekStart = v
	case "author_email":
		cfg.AuthorEmail = value
	case "ai_backend":
		v := strings.ToLower(value)
		if v != "api" && v != "cli" {
			return fmt.Errorf("ai_backend must be 'api' or 'cli'")
		}
		cfg.AIBackend = v
	case "ai_cli_command":
		cfg.AICLICommand = value
	case "ai_base_url":
		cfg.AIBaseURL = value
	case "ai_model":
		cfg.AIModel = value
	case "ai_api_key":
		cfg.AIAPIKey = value
	case "ai_prompt":
		cfg.AIPrompt = value
	case "ai_default_style":
		v := strings.ToLower(value)
		if !ai.IsValidStyle(v) {
			return fmt.Errorf("ai_default_style must be one of: %s", strings.Join(ai.ValidStyleNames(), ", "))
		}
		cfg.AIDefaultStyle = v
	default:
		return fmt.Errorf("unknown key %q\n\nValid keys: %s", key, strings.Join(validConfigKeys, ", "))
	}
	return nil
}

func resolveEditor() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vi"
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configSetCmd)
}
