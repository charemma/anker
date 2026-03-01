package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charemma/anker/internal/config"
	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/sources/git"
	"github.com/charemma/anker/internal/sources/markdown"
	"github.com/charemma/anker/internal/sources/obsidian"
	"github.com/charemma/anker/internal/storage"
	"github.com/charemma/anker/internal/timerange"
	"github.com/spf13/cobra"
)

var (
	recapFormat string
)

var recapCmd = &cobra.Command{
	Use:   "recap [timespec]",
	Short: "Recap your work for a time period",
	Long: `Recap your work from tracked sources - reconstruct what you did after the fact.

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

Output formats (--format):
  simple           Commit messages only (default)
  detailed         Commit messages with timestamps and stats
  json             Structured JSON for programmatic use
  markdown         Markdown with full diffs (for AI/documentation)

Examples:
  anker recap
  anker recap today
  anker recap thisweek --format detailed
  anker recap "December 2025" --format markdown > recap.md
  anker recap 2025-12-01..2025-12-31`,
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
			fmt.Println("No sources configured. Use 'anker source add' to get started.")
			return nil
		}

		// Validate format
		validFormats := map[string]bool{"simple": true, "detailed": true, "json": true, "markdown": true}
		if !validFormats[recapFormat] {
			return fmt.Errorf("invalid format: %s (must be simple, detailed, json, or markdown)", recapFormat)
		}

		// Collect entries from all sources
		var allEntries []sources.Entry
		var gitSources []*git.GitSource // Keep git sources for diff enrichment

		for _, cfg := range sourceConfigs {
			var source sources.Source

			switch cfg.Type {
			case "git":
				authorEmail := ""
				if email, ok := cfg.Metadata["author"]; ok {
					authorEmail = email
				}
				gitSource := git.NewGitSource(cfg.Path, authorEmail)
				source = gitSource
				if recapFormat == "markdown" {
					gitSources = append(gitSources, gitSource)
				}
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
			case "obsidian":
				source = obsidian.NewObsidianSource(cfg.Path)
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

		// Enrich with diffs if markdown format requested
		if recapFormat == "markdown" {
			for _, gitSource := range gitSources {
				// Find entries from this git source and enrich them
				var sourceEntries []sources.Entry
				for _, entry := range allEntries {
					if entry.Location == gitSource.Location() {
						sourceEntries = append(sourceEntries, entry)
					}
				}
				if err := gitSource.EnrichWithDiffs(sourceEntries); err != nil {
					fmt.Printf("Warning: failed to enrich diffs for %s: %v\n", gitSource.Location(), err)
				}
				// Update entries in allEntries with enriched data
				for i := range allEntries {
					if allEntries[i].Location == gitSource.Location() {
						for _, enriched := range sourceEntries {
							if allEntries[i].Metadata["hash"] == enriched.Metadata["hash"] {
								allEntries[i] = enriched
								break
							}
						}
					}
				}
			}
		}

		if len(allEntries) == 0 {
			fmt.Printf("No activity found for %s\n", timespec)
			return nil
		}

		// Sort entries by timestamp (newest first)
		sort.Slice(allEntries, func(i, j int) bool {
			return allEntries[i].Timestamp.After(allEntries[j].Timestamp)
		})

		// Generate report based on format
		switch recapFormat {
		case "simple":
			return printSimpleRecap(allEntries, tr, timespec)
		case "detailed":
			return printDetailedRecap(allEntries, tr, timespec)
		case "json":
			return printJSONRecap(allEntries, tr, timespec)
		case "markdown":
			return printMarkdownRecap(allEntries, tr, timespec)
		default:
			return fmt.Errorf("unknown format: %s", recapFormat)
		}
	},
}

func printSimpleRecap(allEntries []sources.Entry, tr *timerange.TimeRange, timespec string) error {
	fmt.Printf("\n")
	fmt.Printf("Work Recap\n")
	fmt.Printf("==========\n")
	fmt.Printf("Period: %s - %s\n", tr.From.Format("02 Jan 2006"), tr.To.Format("02 Jan 2006"))
	fmt.Printf("Total: %d activities\n\n", len(allEntries))

	// Group by repository/source location
	byRepo := make(map[string][]sources.Entry)
	for _, entry := range allEntries {
		byRepo[entry.Location] = append(byRepo[entry.Location], entry)
	}

	// Get sorted repo names
	repos := make([]string, 0, len(byRepo))
	for repo := range byRepo {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	// Print entries grouped by repository
	for _, repoPath := range repos {
		entries := byRepo[repoPath]
		repoName := repoPath
		if idx := strings.LastIndex(repoName, "/"); idx != -1 {
			repoName = repoName[idx+1:]
		}

		// Determine source type and format header
		if len(entries) > 0 {
			switch entries[0].Source {
			case "obsidian":
				fmt.Printf("Obsidian Vault\n")
				fmt.Printf("%s\n\n", repoName)
			case "git":
				fmt.Printf("Git Repository: %s\n", repoName)
				fmt.Printf("(%s)\n\n", repoPath)
			case "markdown":
				fmt.Printf("Markdown Notes: %s\n", repoName)
				fmt.Printf("(%s)\n\n", repoPath)
			default:
				fmt.Printf("%s\n\n", repoName)
			}
		}

		for _, entry := range entries {
			fmt.Printf("  • %s\n", entry.Content)
		}
		fmt.Println()
	}

	return nil
}

func printDetailedRecap(allEntries []sources.Entry, tr *timerange.TimeRange, timespec string) error {
	fmt.Printf("\n")
	fmt.Printf("Work Recap (Detailed)\n")
	fmt.Printf("=====================\n")
	fmt.Printf("Period: %s - %s\n", tr.From.Format("02 Jan 2006"), tr.To.Format("02 Jan 2006"))
	fmt.Printf("Total: %d activities\n\n", len(allEntries))

	// Group by repository
	byRepo := make(map[string][]sources.Entry)
	for _, entry := range allEntries {
		byRepo[entry.Location] = append(byRepo[entry.Location], entry)
	}

	repos := make([]string, 0, len(byRepo))
	for repo := range byRepo {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	for _, repoPath := range repos {
		entries := byRepo[repoPath]
		repoName := repoPath
		if idx := strings.LastIndex(repoName, "/"); idx != -1 {
			repoName = repoName[idx+1:]
		}

		// Determine source type and format header
		if len(entries) > 0 {
			switch entries[0].Source {
			case "obsidian":
				fmt.Printf("Obsidian Vault\n")
				fmt.Printf("%s\n\n", repoName)
			case "git":
				fmt.Printf("Git Repository: %s\n", repoName)
				fmt.Printf("(%s)\n\n", repoPath)
			case "markdown":
				fmt.Printf("Markdown Notes: %s\n", repoName)
				fmt.Printf("(%s)\n\n", repoPath)
			default:
				fmt.Printf("%s\n\n", repoName)
			}
		}

		for _, entry := range entries {
			fmt.Printf("  %s\n", entry.Timestamp.Format("Mon Jan 2, 15:04"))
			fmt.Printf("  %s\n", entry.Content)
			if author, ok := entry.Metadata["author"]; ok {
				fmt.Printf("  Author: %s\n", author)
			}
			if hash, ok := entry.Metadata["hash"]; ok {
				fmt.Printf("  Commit: %s\n", hash[:8])
			}
			fmt.Println()
		}
	}

	return nil
}

func printJSONRecap(allEntries []sources.Entry, tr *timerange.TimeRange, timespec string) error {
	type JSONReport struct {
		Period struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"period"`
		Total      int             `json:"total"`
		Activities []sources.Entry `json:"activities"`
	}

	report := JSONReport{
		Total:      len(allEntries),
		Activities: allEntries,
	}
	report.Period.From = tr.From.Format("2006-01-02")
	report.Period.To = tr.To.Format("2006-01-02")

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func printMarkdownRecap(allEntries []sources.Entry, tr *timerange.TimeRange, timespec string) error {
	fmt.Printf("# Work Recap\n\n")
	fmt.Printf("**Period:** %s to %s\n", tr.From.Format("2006-01-02"), tr.To.Format("2006-01-02"))
	fmt.Printf("**Total Activities:** %d\n\n", len(allEntries))
	fmt.Printf("---\n\n")
	fmt.Printf("This recap contains git commits with full diffs for the specified period.\n")
	fmt.Printf("Each commit includes the message and the complete code changes.\n\n")

	// Group by repository
	byRepo := make(map[string][]sources.Entry)
	for _, entry := range allEntries {
		byRepo[entry.Location] = append(byRepo[entry.Location], entry)
	}

	repos := make([]string, 0, len(byRepo))
	for repo := range byRepo {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	for _, repoPath := range repos {
		entries := byRepo[repoPath]
		repoName := repoPath
		if idx := strings.LastIndex(repoName, "/"); idx != -1 {
			repoName = repoName[idx+1:]
		}

		// Determine source type and format header
		if len(entries) > 0 {
			switch entries[0].Source {
			case "obsidian":
				fmt.Printf("## Obsidian Vault\n\n")
				fmt.Printf("**%s**\n\n", repoName)
				fmt.Printf("`%s`\n\n", repoPath)
			case "git":
				fmt.Printf("## Git Repository: %s\n\n", repoName)
				fmt.Printf("`%s`\n\n", repoPath)
			case "markdown":
				fmt.Printf("## Markdown Notes: %s\n\n", repoName)
				fmt.Printf("`%s`\n\n", repoPath)
			default:
				fmt.Printf("## %s\n\n", repoName)
				fmt.Printf("`%s`\n\n", repoPath)
			}
		}

		for i, entry := range entries {
			fmt.Printf("### %d.\n\n", i+1)
			fmt.Printf("**Date:** %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))
			if author, ok := entry.Metadata["author"]; ok {
				fmt.Printf("**Author:** %s\n", author)
			}
			if hash, ok := entry.Metadata["hash"]; ok {
				fmt.Printf("**Hash:** `%s`\n", hash)
			}
			fmt.Printf("**Message:** %s\n\n", entry.Content)

			if diff, ok := entry.Metadata["diff"]; ok && diff != "" {
				fmt.Printf("**Changes:**\n\n```diff\n%s\n```\n\n", diff)
			} else {
				fmt.Printf("*(No diff available)*\n\n")
			}

			fmt.Printf("---\n\n")
		}
	}

	return nil
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
	rootCmd.AddCommand(recapCmd)
	recapCmd.Flags().StringVarP(&recapFormat, "format", "f", "simple", "Output format (simple, detailed, json, markdown)")
}
