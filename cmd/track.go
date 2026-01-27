package cmd

import (
	"fmt"
	"os"

	"github.com/charemma/anker/internal/git"
	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/storage"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Mark current repository for later analysis",
	Long:  `Detects the git repository root and registers it as a tracked source.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		repoRoot, err := git.FindRepoRoot(cwd)
		if err != nil {
			return fmt.Errorf("not in a git repository: %w", err)
		}

		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		config := sources.Config{
			Type: "git",
			Path: repoRoot,
		}

		if err := store.AddSource(config); err != nil {
			return fmt.Errorf("failed to track repository: %w", err)
		}

		fmt.Printf("tracked: %s\n", repoRoot)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
}
