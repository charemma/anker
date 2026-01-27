package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charemma/anker/internal/config"
	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/sources/git"
	"github.com/charemma/anker/internal/sources/markdown"
	"github.com/charemma/anker/internal/storage"
	"github.com/charemma/anker/internal/timerange"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report [timespec]",
	Short: "Generate a work summary for a time period",
	Long: `Generate a summary of your work from tracked sources.

Time specifications:
  today            Current day (default)
  yesterday        Previous day
  thisweek         Current week
  lastweek         Previous week
  week 32          Specific calendar week
  week 32 2024     Week in specific year
  2025-12-02       Specific date
  2025-12-01..31   Date range
  last 7 days      Relative range

Examples:
  anker report
  anker report today
  anker report thisweek
  anker report week 25
  anker report 2025-12-01..2025-12-31
  anker report last 7 days`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to "today" if no argument provided
		timespec := "today"
		if len(args) > 0 {
			timespec = args[0]
		}

		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Parse time specification
		parser := timerange.NewParser(cfg.GetTimerangeConfig())
		tr, err := parser.Parse(timespec)
		if err != nil {
			return fmt.Errorf("invalid time specification: %w", err)
		}

		// Load tracked sources
		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		sourceConfigs, err := store.GetSources()
		if err != nil {
			return fmt.Errorf("failed to load sources: %w", err)
		}

		if len(sourceConfigs) == 0 {
			fmt.Println("No sources configured. Use 'anker track' or 'anker source add' to get started.")
			return nil
		}

		// Collect entries from all sources
		var allEntries []sources.Entry

		for _, cfg := range sourceConfigs {
			var source sources.Source

			switch cfg.Type {
			case "git":
				source = git.NewGitSource(cfg.Path)
			case "markdown":
				tags := []string{}
				headings := []string{}

				if tagsStr, ok := cfg.Metadata["tags"]; ok && tagsStr != "" {
					tags = splitTrimmed(tagsStr, ",")
				}
				if headingsStr, ok := cfg.Metadata["headings"]; ok && headingsStr != "" {
					headings = splitTrimmed(headingsStr, ",")
				}

				source = markdown.NewMarkdownSource(cfg.Path, tags, headings)
			default:
				fmt.Printf("Warning: unsupported source type '%s' at %s\n", cfg.Type, cfg.Path)
				continue
			}

			entries, err := source.GetEntries(tr.From, tr.To)
			if err != nil {
				fmt.Printf("Warning: failed to get entries from %s %s: %v\n", cfg.Type, cfg.Path, err)
				continue
			}

			allEntries = append(allEntries, entries...)
		}

		if len(allEntries) == 0 {
			fmt.Printf("No activity found for %s\n", timespec)
			return nil
		}

		// Sort entries by timestamp (newest first)
		sort.Slice(allEntries, func(i, j int) bool {
			return allEntries[i].Timestamp.After(allEntries[j].Timestamp)
		})

		// Generate report
		fmt.Printf("Work Summary: %s\n", timespec)
		fmt.Printf("Period: %s to %s\n", tr.From.Format("2006-01-02 15:04"), tr.To.Format("2006-01-02 15:04"))
		fmt.Printf("Found %d entries\n\n", len(allEntries))

		// Group by source type
		bySource := make(map[string][]sources.Entry)
		for _, entry := range allEntries {
			bySource[entry.Source] = append(bySource[entry.Source], entry)
		}

		// Print grouped entries
		for sourceType, entries := range bySource {
			fmt.Printf("=== %s (%d entries) ===\n", sourceType, len(entries))

			for _, entry := range entries {
				timestamp := entry.Timestamp.Format("Mon 15:04")
				fmt.Printf("[%s] %s\n", timestamp, entry.Content)

				// Print location for context
				if sourceType == "git" {
					if hash, ok := entry.Metadata["hash"]; ok {
						fmt.Printf("        %s (%s)\n", entry.Location, hash[:7])
					}
				} else if sourceType == "markdown" {
					if file, ok := entry.Metadata["file"]; ok {
						fmt.Printf("        %s\n", file)
					}
				}
			}
			fmt.Println()
		}

		return nil
	},
}

func splitTrimmed(s, sep string) []string {
	parts := []string{}
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
