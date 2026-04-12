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

const defaultAIPrompt = `Summarize my workday based on the activity log below.
Start with the period as a heading (e.g. "# Recap: 2026-02-23 to 2026-03-01").
Group by topic or theme, not chronologically.
Keep it concise -- a few bullet points per topic.
Skip trivial entries (typo fixes, formatting, etc.) unless they are part of a larger change.
For each topic, also highlight:
- Decisions made and why (e.g. chose X over Y because...)
- Key insights or lessons learned
- Open threads or unfinished business
Only include these if they are actually present in the data -- don't invent them.`

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
			fmt.Println("No sources configured. Use 'anker source add' to get started.")
			return nil
		}

		// Validate format
		validFormats := map[string]bool{"simple": true, "detailed": true, "json": true, "markdown": true, "ai": true}
		if !validFormats[recapFormat] {
			return fmt.Errorf("invalid format: %s (must be simple, detailed, json, markdown, or ai)", recapFormat)
		}

		// Collect entries from all sources
		result, err := recap.BuildRecap(sourceConfigs, tr, timespec, recapFormat, createSource, os.Stderr)
		if err != nil {
			return err
		}

		if len(result.Entries) == 0 {
			fmt.Printf("No activity found for %s\n", timespec)
			return nil
		}

		// AI format uses detailed rendering as input, then transforms via LLM
		if recapFormat == "ai" {
			return printAIRecap(cfg, result, tr, timespec)
		}

		return recap.Render(os.Stdout, result, recapFormat)
	},
}

func printAIRecap(cfg *config.Config, result *recap.RecapResult, tr *timerange.TimeRange, timespec string) error {
	// Render detailed output to a string
	var buf bytes.Buffer
	if err := recap.RenderDetailed(&buf, result); err != nil {
		return fmt.Errorf("failed to render recap: %w", err)
	}

	// Resolve prompt: --prompt flag > config > default
	prompt := defaultAIPrompt
	if cfg.AIPrompt != "" {
		prompt = cfg.AIPrompt
	}
	if recapPrompt != "" {
		prompt = recapPrompt
	}

	// Inject time range context
	period := fmt.Sprintf("%s to %s", tr.From.Format("2006-01-02"), tr.To.Format("2006-01-02"))
	prompt = fmt.Sprintf("Period: %s (%s)\n\n%s", timespec, period, prompt)

	if cfg.AIBackend == "cli" {
		return ai.RunCLI(cfg.AICLICommand, prompt, buf.String(), os.Stdout)
	}

	// Resolve API key: --api-key flag > AI_API_KEY env > config
	apiKey := cfg.AIAPIKey
	if envKey := os.Getenv("AI_API_KEY"); envKey != "" {
		apiKey = envKey
	}
	if recapAPIKey != "" {
		apiKey = recapAPIKey
	}

	client := &ai.Client{
		BaseURL: cfg.AIBaseURL,
		APIKey:  apiKey,
		Model:   cfg.AIModel,
	}

	return client.StreamCompletion(context.Background(), prompt, buf.String(), os.Stdout)
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
