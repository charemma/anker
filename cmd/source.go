package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charemma/ikno/internal/config"
	"github.com/charemma/ikno/internal/git"
	"github.com/charemma/ikno/internal/sources"
	claudesource "github.com/charemma/ikno/internal/sources/claude"
	"github.com/charemma/ikno/internal/storage"
	"github.com/charemma/ikno/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	gitAuthors       []string
	markdownTags     []string
	markdownHeadings []string
	addType          string
	addYes           bool
)

// knownTypes is the set of built-in source type identifiers.
var knownTypes = map[string]bool{
	"git":      true,
	"markdown": true,
	"obsidian": true,
	"claude":   true,
}

var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage data sources",
	Long:  `Add, list, or remove data sources for activity tracking.`,
}

var sourceAddCmd = &cobra.Command{
	Use:   "add [type] [path]",
	Short: "Add a new data source",
	Long: `Add a new data source for tracking.

Supported types:
  git      - Track git repository commits
  markdown - Track markdown files (notes, journals, etc.)
  obsidian - Track Obsidian vault file changes
  claude   - Track Claude Code session interactions

With auto-detection:
  ikno source add                      detect and add cwd
  ikno source add ~/path               detect ~/path (or scan children)
  ikno source add git ~/path           explicit type (unchanged)
  ikno source add ~/code --type git    force type on a path

Examples:
  ikno source add
  ikno source add ~/code/my-project
  ikno source add ~/code/charemma      (scans directory children)
  ikno source add git .
  ikno source add git ~/code/my-project
  ikno source add git . --author user@example.com
  ikno source add markdown ~/Obsidian/Daily
  ikno source add markdown ~/notes --tags work,done
  ikno source add obsidian ~/Documents/Obsidian
  ikno source add claude`,
	Args: cobra.RangeArgs(0, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// Complete both known types and directories
			types := []string{"git", "markdown", "obsidian", "claude"}
			return types, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveFilterDirs
		}
		if len(args) == 1 && knownTypes[args[0]] {
			return nil, cobra.ShellCompDirectiveFilterDirs
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		sourceType, path, explicit := parseAddArgs(args)

		// Validate ambiguous 2-arg case where first arg is not a known type
		if len(args) == 2 && !knownTypes[args[0]] {
			return fmt.Errorf("unrecognized source type %q -- use: ikno source add <type> <path> or ikno source add <path>", args[0])
		}

		// --type flag overrides auto-detection (only when not already explicit)
		if !explicit && addType != "" {
			sourceType = addType
			explicit = true
		}

		if explicit {
			// Legacy path: type is known, add single source directly
			return addSingleSource(store, sourceType, path)
		}

		// Auto-detection path
		detected, err := sources.DetectType(path)
		if err != nil {
			return fmt.Errorf("failed to detect source type: %w", err)
		}

		if len(detected) > 0 {
			// Single-path flow
			return handleDetectedSources(store, detected)
		}

		// No match at path itself -- scan children
		registered, err := store.GetSources()
		if err != nil {
			return fmt.Errorf("failed to load sources: %w", err)
		}

		discovered, err := sources.DiscoverSources(path, 1, registered)
		if err != nil {
			return fmt.Errorf("failed to scan directory: %w", err)
		}

		if len(discovered) == 0 {
			return fmt.Errorf("could not detect source type for %s, use --type to specify", path)
		}

		return handleDiscoveredBatch(store, discovered)
	},
}

// parseAddArgs disambiguates the argument list and returns (sourceType, path, explicit).
// explicit=true means the user provided the type directly.
func parseAddArgs(args []string) (sourceType, path string, explicit bool) {
	switch len(args) {
	case 0:
		cwd, _ := os.Getwd()
		path = cwd
	case 1:
		if knownTypes[args[0]] {
			// legacy: "ikno source add claude" or "ikno source add git"
			sourceType = args[0]
			explicit = true
		} else {
			// path with auto-detect
			path = args[0]
		}
	case 2:
		if knownTypes[args[0]] {
			// legacy: "ikno source add git ~/path"
			sourceType = args[0]
			path = args[1]
			explicit = true
		}
		// else: caller handles the error (ambiguous 2-arg case)
	}
	return
}

// handleDetectedSources handles the result of DetectType for a single path.
func handleDetectedSources(store *storage.Store, detected []sources.DetectedSource) error {
	if len(detected) == 1 {
		d := detected[0]
		fmt.Printf("detected %s: %s (%s)\n", d.Type, d.Path, d.Reason)
		return addSingleSource(store, d.Type, d.Path)
	}

	// Multiple matches: prompt unless --yes
	if addYes {
		for _, d := range detected {
			if err := addSingleSource(store, d.Type, d.Path); err != nil {
				return err
			}
		}
		return nil
	}

	if !isTTY() {
		return fmt.Errorf("ambiguous source type, use --type to specify")
	}

	_, _ = fmt.Fprintf(os.Stdout, "Detected multiple source types for %s:\n", detected[0].Path)
	for i, d := range detected {
		_, _ = fmt.Fprintf(os.Stdout, "  %d) %-10s (%s)\n", i+1, d.Type, d.Reason)
	}

	choices := make([]string, len(detected))
	for i := range detected {
		choices[i] = fmt.Sprintf("%d", i+1)
	}
	choices = append(choices, "all", "none")
	_, _ = fmt.Fprintf(os.Stdout, "Add which? [%s]: ", strings.Join(choices, "/"))

	answer := readLine()
	answer = strings.TrimSpace(strings.ToLower(answer))

	switch answer {
	case "none", "":
		fmt.Println("no sources added")
		return nil
	case "all":
		for _, d := range detected {
			if err := addSingleSource(store, d.Type, d.Path); err != nil {
				return err
			}
		}
		return nil
	default:
		// Try to match a number
		for i, d := range detected {
			if answer == fmt.Sprintf("%d", i+1) {
				return addSingleSource(store, d.Type, d.Path)
			}
		}
		return fmt.Errorf("invalid choice %q", answer)
	}
}

// handleDiscoveredBatch handles batch confirmation for directory scanning results.
func handleDiscoveredBatch(store *storage.Store, discovered []sources.DetectedSource) error {
	// Group by type for display
	fmt.Printf("Found %d source(s):\n", len(discovered))
	for _, d := range discovered {
		fmt.Printf("  %-10s %s\n", d.Type, d.Path)
	}

	if addYes {
		return addAll(store, discovered)
	}

	if !isTTY() {
		return fmt.Errorf("interactive confirmation required, use --yes to skip")
	}

	_, _ = fmt.Fprintf(os.Stdout, "Add all? [Y/n]: ")
	answer := readLine()
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "" || answer == "y" || answer == "yes" {
		return addAll(store, discovered)
	}

	// Item-by-item confirmation
	for _, d := range discovered {
		_, _ = fmt.Fprintf(os.Stdout, "Add %s %s? [y/n/quit]: ", d.Type, d.Path)
		ans := strings.TrimSpace(strings.ToLower(readLine()))
		switch ans {
		case "y", "yes":
			if err := addSingleSource(store, d.Type, d.Path); err != nil {
				return err
			}
		case "quit", "q":
			return nil
		}
	}

	return nil
}

// addAll adds every discovered source without prompting.
func addAll(store *storage.Store, discovered []sources.DetectedSource) error {
	for _, d := range discovered {
		if err := addSingleSource(store, d.Type, d.Path); err != nil {
			return err
		}
	}
	return nil
}

// addSingleSource validates and stores a single source config.
func addSingleSource(store *storage.Store, sourceType, path string) error {
	srcCfg := sources.Config{
		Type:     sourceType,
		Path:     path,
		Metadata: make(map[string]string),
	}

	if path == "" && sourceType != "claude" {
		return fmt.Errorf("path is required for source type: %s", sourceType)
	}

	switch sourceType {
	case "git":
		if len(gitAuthors) > 0 {
			srcCfg.Metadata["author"] = strings.Join(gitAuthors, ",")
		} else {
			cfg, err := config.Load()
			if err == nil && cfg.AuthorEmail != "" {
				srcCfg.Metadata["author"] = cfg.AuthorEmail
			} else {
				if email, err := git.GetAuthorEmail(); err == nil && email != "" {
					srcCfg.Metadata["author"] = email
					_, _ = fmt.Printf("using git user.email: %s\n", email)
				} else {
					_, _ = fmt.Println(ui.StyleMuted.Render("warning: no author email configured - will track ALL commits in this repo"))
					_, _ = fmt.Println(ui.StyleMuted.Render("  set author with: --author your@email.com"))
					_, _ = fmt.Println(ui.StyleMuted.Render("  or configure git: git config --global user.email your@email.com"))
				}
			}
		}
	case "markdown":
		if len(markdownTags) > 0 {
			srcCfg.Metadata["tags"] = strings.Join(markdownTags, ",")
		}
		if len(markdownHeadings) > 0 {
			srcCfg.Metadata["headings"] = strings.Join(markdownHeadings, ",")
		}
	case "obsidian":
		// no additional metadata
	case "claude":
		if path == "" {
			path = claudesource.DefaultClaudeHome()
			srcCfg.Path = path
		}
	default:
		return fmt.Errorf("unsupported source type: %s (supported: git, markdown, obsidian, claude)", sourceType)
	}

	if err := store.AddSource(srcCfg); err != nil {
		return fmt.Errorf("failed to add source: %w", err)
	}

	typeStyled := lipgloss.NewStyle().Foreground(ui.SourceColor(sourceType)).Render(sourceType)
	_, _ = fmt.Printf("%s %s source: %s\n",
		ui.StyleSuccess.Render("added"),
		typeStyled,
		srcCfg.Path)
	return nil
}

// isTTY reports whether stdout is a terminal.
func isTTY() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

// readLine reads a single line from stdin.
func readLine() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

var sourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		configs, err := store.GetSources()
		if err != nil {
			return fmt.Errorf("failed to get sources: %w", err)
		}

		if len(configs) == 0 {
			_, _ = fmt.Println(ui.StyleMuted.Render("no sources configured"))
			return nil
		}

		for _, cfg := range configs {
			typeStyled := lipgloss.NewStyle().Foreground(ui.SourceColor(cfg.Type)).Render(cfg.Type)
			_, _ = fmt.Printf("%s  %s\n", typeStyled, cfg.Path)
			if len(cfg.Metadata) > 0 {
				for k, v := range cfg.Metadata {
					_, _ = fmt.Printf("  %s\n", ui.StyleMuted.Render(k+": "+v))
				}
			}
		}

		return nil
	},
}

var sourceRemoveCmd = &cobra.Command{
	Use:   "remove [path] or remove [type] [path]",
	Short: "Remove a data source",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		var sourceType, path string

		if len(args) == 1 {
			path = args[0]
			removed, err := store.RemoveSourceByPath(path)
			if err != nil {
				return fmt.Errorf("failed to remove source: %w", err)
			}
			typeStyled := lipgloss.NewStyle().Foreground(ui.SourceColor(removed.Type)).Render(removed.Type)
			_, _ = fmt.Printf("%s %s source: %s\n",
				ui.StyleSuccess.Render("removed"),
				typeStyled,
				removed.Path)
		} else {
			sourceType = args[0]
			path = args[1]
			if err := store.RemoveSource(sourceType, path); err != nil {
				return fmt.Errorf("failed to remove source: %w", err)
			}
			typeStyled := lipgloss.NewStyle().Foreground(ui.SourceColor(sourceType)).Render(sourceType)
			_, _ = fmt.Printf("%s %s source: %s\n",
				ui.StyleSuccess.Render("removed"),
				typeStyled,
				path)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(sourceCmd)
	sourceCmd.AddCommand(sourceAddCmd)
	sourceCmd.AddCommand(sourceListCmd)
	sourceCmd.AddCommand(sourceRemoveCmd)

	sourceAddCmd.Flags().StringSliceVar(&gitAuthors, "author", nil, "Git author email(s) to filter commits (can be specified multiple times)")
	sourceAddCmd.Flags().StringSliceVar(&markdownTags, "tags", nil, "Filter markdown by tags (comma-separated)")
	sourceAddCmd.Flags().StringSliceVar(&markdownHeadings, "headings", nil, "Filter markdown by headings (comma-separated)")
	sourceAddCmd.Flags().StringVarP(&addType, "type", "t", "", "Force source type (overrides auto-detection)")
	sourceAddCmd.Flags().BoolVarP(&addYes, "yes", "y", false, "Skip interactive confirmation, add all discovered sources")
}
