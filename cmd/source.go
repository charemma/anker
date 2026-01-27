package cmd

import (
	"fmt"
	"strings"

	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/storage"
	"github.com/spf13/cobra"
)

var (
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
  markdown - Track markdown files (notes, journals, etc.)

Examples:
  anker source add markdown ~/Obsidian/Daily
  anker source add markdown ~/notes --tags work,done`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceType := args[0]
		path := args[1]

		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		config := sources.Config{
			Type:     sourceType,
			Path:     path,
			Metadata: make(map[string]string),
		}

		switch sourceType {
		case "markdown":
			if len(markdownTags) > 0 {
				config.Metadata["tags"] = strings.Join(markdownTags, ",")
			}
			if len(markdownHeadings) > 0 {
				config.Metadata["headings"] = strings.Join(markdownHeadings, ",")
			}
		default:
			return fmt.Errorf("unsupported source type: %s", sourceType)
		}

		if err := store.AddSource(config); err != nil {
			return fmt.Errorf("failed to add source: %w", err)
		}

		fmt.Printf("added %s source: %s\n", sourceType, path)
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
			fmt.Println("no sources configured")
			return nil
		}

		for _, config := range configs {
			fmt.Printf("%s: %s\n", config.Type, config.Path)
			if len(config.Metadata) > 0 {
				for k, v := range config.Metadata {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}
		}

		return nil
	},
}

var sourceRemoveCmd = &cobra.Command{
	Use:   "remove [type] [path]",
	Short: "Remove a data source",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceType := args[0]
		path := args[1]

		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		if err := store.RemoveSource(sourceType, path); err != nil {
			return fmt.Errorf("failed to remove source: %w", err)
		}

		fmt.Printf("removed %s source: %s\n", sourceType, path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sourceCmd)
	sourceCmd.AddCommand(sourceAddCmd)
	sourceCmd.AddCommand(sourceListCmd)
	sourceCmd.AddCommand(sourceRemoveCmd)

	sourceAddCmd.Flags().StringSliceVar(&markdownTags, "tags", nil, "Filter markdown by tags (comma-separated)")
	sourceAddCmd.Flags().StringSliceVar(&markdownHeadings, "headings", nil, "Filter markdown by headings (comma-separated)")
}
