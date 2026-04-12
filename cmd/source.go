package cmd

import (
	"fmt"
	"strings"

	"github.com/charemma/anker/internal/config"
	"github.com/charemma/anker/internal/git"
	"github.com/charemma/anker/internal/sources"
	claudesource "github.com/charemma/anker/internal/sources/claude"
	"github.com/charemma/anker/internal/storage"
	"github.com/charemma/anker/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	gitAuthors       []string
	markdownTags     []string
	markdownHeadings []string
)

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

Examples:
  anker source add git .
  anker source add git ~/code/my-project
  anker source add git . --author user@example.com
  anker source add git . --author foo@work.com --author bar@personal.com
  anker source add markdown ~/Obsidian/Daily
  anker source add markdown ~/notes --tags work,done
  anker source add obsidian ~/Documents/Obsidian
  anker source add claude`,
	Args: cobra.RangeArgs(1, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"git", "markdown", "obsidian", "claude"}, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveFilterDirs
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceType := args[0]
		path := ""
		if len(args) > 1 {
			path = args[1]
		}

		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

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
			// Use --author flags if provided, otherwise use config, otherwise use git config
			if len(gitAuthors) > 0 {
				srcCfg.Metadata["author"] = strings.Join(gitAuthors, ",")
			} else {
				// Fallback to anker config if available
				cfg, err := config.Load()
				if err == nil && cfg.AuthorEmail != "" {
					srcCfg.Metadata["author"] = cfg.AuthorEmail
				} else {
					// Final fallback: git config --global user.email
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
			// Obsidian source has no additional metadata for now
		case "claude":
			if path == "" {
				path = claudesource.DefaultClaudeHome()
			}
			srcCfg.Path = path
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
	},
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
			// Only path provided - find source by path
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
			// Type and path provided
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
}
