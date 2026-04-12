package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charemma/anker/internal/config"
	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/storage"
	"github.com/spf13/cobra"
)

var initYes bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup wizard",
	Long: `Scan common locations and add sources interactively.

Scans:
  ~/code/         git repositories (depth 1)
  ~/.claude/      Claude Code sessions
  ~/Documents/    Obsidian vault or markdown directory
  ./              current directory

Examples:
  anker init
  anker init --yes`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		registered, err := store.GetSources()
		if err != nil {
			return fmt.Errorf("failed to load sources: %w", err)
		}

		fmt.Println("Scanning for sources...")

		discovered := scanDefaultLocations(registered)

		// Deduplicate by (type, absPath)
		discovered = dedup(discovered)

		if len(discovered) == 0 {
			fmt.Println()
			fmt.Println("No new sources found.")
			fmt.Println()
			fmt.Println("Add sources manually:")
			fmt.Println("  anker source add git .              track commits in the current repo")
			fmt.Println("  anker source add ~/code/charemma   scan a directory for git repos")
			return nil
		}

		fmt.Println()
		printGrouped(discovered)
		fmt.Println()

		if initYes {
			return addAll(store, discovered)
		}

		if !isTTY() {
			return fmt.Errorf("interactive confirmation required, use --yes to skip")
		}

		_, _ = fmt.Fprintf(os.Stdout, "Review each location? [Y/n]: ")
		answer := strings.TrimSpace(strings.ToLower(readLine()))

		if answer == "n" || answer == "no" {
			// Add everything without per-item prompts
			return addAll(store, discovered)
		}

		// Per-item confirmation
		added := 0
		for _, d := range discovered {
			_, _ = fmt.Fprintf(os.Stdout, "%-10s %s  add? [y/n/skip-all]: ", d.Type, d.Path)
			ans := strings.TrimSpace(strings.ToLower(readLine()))
			switch ans {
			case "y", "yes", "":
				if err := addSingleSource(store, d.Type, d.Path); err != nil {
					return err
				}
				added++
			case "skip-all", "s":
				fmt.Println("skipping remaining sources")
				goto done
			}
		}

	done:
		fmt.Printf("\nAdded %d source(s). Run `anker recap today` to get started.\n", added)

		// Write config file if it doesn't exist
		if _, err := config.EnsureConfigFile(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: could not write config file: %v\n", err)
		}

		return nil
	},
}

// scanDefaultLocations scans the standard anker user locations and returns candidates.
func scanDefaultLocations(registered []sources.Config) []sources.DetectedSource {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var all []sources.DetectedSource

	locations := []struct {
		path  string
		depth int
	}{
		{filepath.Join(home, "code"), 1},
		{filepath.Join(home, ".claude"), 0},   // DetectType directly
		{filepath.Join(home, "Documents"), 0}, // DetectType directly
		{".", 0},                              // cwd
	}

	for _, loc := range locations {
		absPath, err := filepath.Abs(loc.path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err != nil {
			continue // location doesn't exist
		}

		if loc.depth == 0 {
			// Check the location itself
			detected, err := sources.DetectType(absPath)
			if err != nil {
				continue
			}
			all = append(all, detected...)
		} else {
			// Scan children
			found, err := sources.DiscoverSources(absPath, loc.depth, registered)
			if err != nil {
				continue
			}
			all = append(all, found...)
		}
	}

	return all
}

// dedup removes duplicate (type, path) pairs from the slice.
func dedup(in []sources.DetectedSource) []sources.DetectedSource {
	seen := make(map[string]bool, len(in))
	out := make([]sources.DetectedSource, 0, len(in))
	for _, d := range in {
		key := d.Type + "\x00" + d.Path
		if !seen[key] {
			seen[key] = true
			out = append(out, d)
		}
	}
	return out
}

// printGrouped prints sources grouped by their parent directory for readability.
func printGrouped(discovered []sources.DetectedSource) {
	for _, d := range discovered {
		fmt.Printf("  %-10s %s\n", d.Type, d.Path)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "Skip interactive confirmation, add all discovered sources")
}
