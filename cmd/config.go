package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charemma/anker/internal/config"
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
		fmt.Println(configPath)
		return nil
	},
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
}
