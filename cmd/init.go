package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charemma/anker/internal/config"
	"github.com/charemma/anker/internal/git"
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
  ~/code/               git repositories (depth 1)
  ~/.claude/            Claude Code sessions
  ~/Documents/Notes/    Obsidian vault or markdown directory
  ./                    current directory

Examples:
  anker init
  anker init --yes`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !initIsTTY() && !initYes {
			return fmt.Errorf("interactive confirmation required, use --yes to skip")
		}

		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		registered, err := store.GetSources()
		if err != nil {
			return fmt.Errorf("failed to load sources: %w", err)
		}

		fmt.Println("Scanning for sources...")

		discovered := initScanLocations(registered)
		discovered = initDedup(discovered)

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
		initPrintGrouped(discovered)
		fmt.Println()

		added := 0

		if initYes {
			added, err = initAddAll(store, discovered)
			if err != nil {
				return err
			}
		} else {
			_, _ = fmt.Fprint(os.Stdout, "Review each location? [Y/n]: ")
			answer := strings.TrimSpace(strings.ToLower(initReadLine()))

			if answer == "n" || answer == "no" {
				added, err = initAddAll(store, discovered)
				if err != nil {
					return err
				}
			} else {
				// Per-item confirmation
			loop:
				for _, d := range discovered {
					_, _ = fmt.Fprintf(os.Stdout, "%-10s %s  add? [y/n/skip-all]: ", d.Type, d.Path)
					ans := strings.TrimSpace(strings.ToLower(initReadLine()))
					switch ans {
					case "y", "yes", "":
						if addErr := initAddSource(store, d.Type, d.Path); addErr != nil {
							return addErr
						}
						added++
					case "skip-all", "s":
						fmt.Println("skipping remaining sources")
						break loop
					}
				}
			}
		}

		// Write config file if it doesn't exist
		if _, cfgErr := config.EnsureConfigFile(); cfgErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: could not write config file: %v\n", cfgErr)
		}

		fmt.Printf("\nAdded %d source(s). Run `anker recap today` to get started.\n", added)
		return nil
	},
}

// initScanLocations scans the standard anker user locations and returns candidates.
func initScanLocations(registered []sources.Config) []sources.DetectedSource {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var all []sources.DetectedSource

	type scanEntry struct {
		path  string
		depth int
	}

	entries := []scanEntry{
		{filepath.Join(home, "code"), 1},
		{filepath.Join(home, ".claude"), 0},
		{filepath.Join(home, "Documents", "Notes"), 0},
	}

	// Add cwd only when it is not the home directory itself -- scanning home
	// as a source candidate makes no sense.
	if cwd, cwdErr := os.Getwd(); cwdErr == nil {
		cwdAbs, _ := filepath.Abs(cwd)
		homeAbs, _ := filepath.Abs(home)
		if cwdAbs != homeAbs {
			entries = append(entries, scanEntry{cwd, 0})
		}
	}

	for _, e := range entries {
		absPath, absErr := filepath.Abs(e.path)
		if absErr != nil {
			continue
		}
		if _, statErr := os.Stat(absPath); statErr != nil {
			continue
		}

		if e.depth == 0 {
			detected, detectErr := sources.DetectType(absPath)
			if detectErr != nil {
				continue
			}
			all = append(all, detected...)
		} else {
			found, discoverErr := sources.DiscoverSources(absPath, e.depth, registered)
			if discoverErr != nil {
				continue
			}
			all = append(all, found...)
		}
	}

	return all
}

// initDedup removes duplicate (type, path) pairs.
func initDedup(in []sources.DetectedSource) []sources.DetectedSource {
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

// initPrintGrouped prints sources with type and path.
func initPrintGrouped(discovered []sources.DetectedSource) {
	for _, d := range discovered {
		fmt.Printf("  %-10s %s\n", d.Type, d.Path)
	}
}

// initAddAll adds all discovered sources and returns the count added.
func initAddAll(store *storage.Store, discovered []sources.DetectedSource) (int, error) {
	added := 0
	for _, d := range discovered {
		if err := initAddSource(store, d.Type, d.Path); err != nil {
			return added, err
		}
		added++
	}
	return added, nil
}

// initAddSource adds a single source to the store, resolving author email for git sources.
func initAddSource(store *storage.Store, sourceType, path string) error {
	cfg := sources.Config{
		Type:     sourceType,
		Path:     path,
		Metadata: make(map[string]string),
	}

	if sourceType == "git" {
		if email, err := git.GetAuthorEmail(); err == nil && email != "" {
			cfg.Metadata["author"] = email
		}
	}

	return store.AddSource(cfg)
}

// initIsTTY reports whether stdin is an interactive terminal.
func initIsTTY() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && (fi.Mode()&os.ModeCharDevice) != 0
}

// initReadLine reads one line from stdin.
func initReadLine() string {
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "Skip interactive confirmation, add all discovered sources")
}
