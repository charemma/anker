package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/charemma/anker/internal/ai"
	"github.com/charemma/anker/internal/config"
	"github.com/charemma/anker/internal/recap"
	"github.com/charemma/anker/internal/storage"
	"github.com/charemma/anker/internal/timerange"
	"github.com/spf13/cobra"
)

var (
	recapFormat string
	recapPrompt string
	recapAPIKey string
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
  ai               AI-generated summary via OpenAI-compatible API

Examples:
  anker recap
  anker recap today
  anker recap thisweek --format detailed
  anker recap "December 2025" --format markdown > recap.md
  anker recap 2025-12-01..2025-12-31
  anker recap today --format ai`,
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
			fmt.Println("No sources configured yet.")
			fmt.Println()
			fmt.Println("Quick setup options:")
			fmt.Println("  anker source add        Add the current directory")
			fmt.Println("  anker source add ~/code Add a directory of repos")
			return nil
		}

		// Validate format
		validFormats := map[string]bool{"simple": true, "detailed": true, "json": true, "markdown": true, "ai": true}
		if !validFormats[recapFormat] {
			return fmt.Errorf("invalid format: %s (must be simple, detailed, json, markdown, or ai)", recapFormat)
		}

		// Collect entries from all sources
		result, err := recap.BuildRecap(sourceConfigs, tr, timespec, recap.BuildOptions{EnrichDiffs: recapFormat == "markdown"}, createSource, os.Stderr)
		if err != nil {
			return err
		}

		if len(result.Entries) == 0 {
			fmt.Printf("No activity found for %s\n", timespec)
			return nil
		}

		// AI format uses detailed rendering as input, then transforms via LLM
		if recapFormat == "ai" {
			var buf bytes.Buffer
			if err := recap.RenderDetailed(&buf, result); err != nil {
				return fmt.Errorf("failed to render recap: %w", err)
			}

			period := fmt.Sprintf("%s (%s to %s)", timespec, tr.From.Format("2006-01-02"), tr.To.Format("2006-01-02"))
			return ai.Transform(context.Background(), os.Stdout, buf.String(), period, ai.TransformConfig{
				AIPrompt:     cfg.AIPrompt,
				AIBackend:    cfg.AIBackend,
				AICLICommand: cfg.AICLICommand,
				AIBaseURL:    cfg.AIBaseURL,
				AIModel:      cfg.AIModel,
				AIAPIKey:     cfg.AIAPIKey,
			}, recapPrompt, recapAPIKey)
		}

		return recap.Render(os.Stdout, result, recapFormat)
	},
}

func init() {
	rootCmd.AddCommand(recapCmd)
	recapCmd.Flags().StringVarP(&recapFormat, "format", "f", "simple", "Output format (simple, detailed, json, markdown, ai)")
	recapCmd.Flags().StringVar(&recapPrompt, "prompt", "", "Custom prompt for AI summary (--format ai)")
	recapCmd.Flags().StringVar(&recapAPIKey, "api-key", "", "API key for AI summary (--format ai)")

	_ = recapCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"simple", "detailed", "json", "markdown", "ai"}, cobra.ShellCompDirectiveNoFileComp
	})
}
