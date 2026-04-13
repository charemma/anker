package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/charemma/ikno/internal/ai"
	"github.com/charemma/ikno/internal/config"
	"github.com/charemma/ikno/internal/paths"
	"github.com/charemma/ikno/internal/recap"
	"github.com/charemma/ikno/internal/sources"
	"github.com/charemma/ikno/internal/storage"
	"github.com/charemma/ikno/internal/timerange"
	"github.com/charemma/ikno/internal/ui"
	"github.com/spf13/cobra"
)

var (
	recapPrompt  string
	recapAPIKey  string
	recapRaw     bool
	recapJSON    bool
	recapStyle   string
	recapLang    string
	recapStyles  bool
	recapVerbose bool
)

var recapCmd = &cobra.Command{
	Use:   "recap [timespec]",
	Short: "Recap your work for a time period",
	Long: `Recap your work from tracked sources via AI summary.

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

Output styles (--style):
  brief            Done/Next/Blocker, tight, max 12 lines (defaults to yesterday)
  digest           Thematic overview, all sources -- default
  status           Progress/Blocker/Next, progress-focused
  report           Polished formal report, deliveries
  retro            Retrospective: good/bad/learnings

Output modes:
  (default)        AI-generated summary (requires AI backend)
  --raw            Unformatted entry dump, one per line -- for pipes, grep
  --json           Structured JSON

Examples:
  anker recap
  anker recap today
  anker recap thisweek
  anker recap thisweek --style digest
  anker recap --style brief
  anker recap lastweek --raw | grep feat
  anker recap 2025-12-01..2025-12-31 --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if recapStyles {
			return runListStyles(os.Stdout, recapVerbose)
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Resolve style: --style flag > config ai_default_style > "digest"
		style := resolveStyle(recapStyle, cfg.AIDefaultStyle)

		// Resolve language: --lang flag > config ai_language > "deutsch"
		lang := resolveLanguage(recapLang, cfg.AILanguage)

		// Resolve timespec: explicit arg > style default
		timespec := ai.DefaultTimespec(style)
		if len(args) > 0 {
			timespec = args[0]
		}

		parser := timerange.NewParser(cfg.GetTimerangeConfig())
		tr, err := parser.Parse(timespec)
		if err != nil {
			return fmt.Errorf("invalid time specification: %w", err)
		}

		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		sourceConfigs, err := store.GetSources()
		if err != nil {
			return fmt.Errorf("failed to load sources: %w", err)
		}

		if len(sourceConfigs) == 0 {
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleNormal.Render("No sources configured yet."))
			_, _ = fmt.Fprintln(os.Stdout)
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleMuted.Render("Quick setup options:"))
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleMuted.Render("  anker source add        Add the current directory"))
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleMuted.Render("  anker source add ~/code Add a directory of repos"))
			return nil
		}

		result, err := recap.BuildRecap(sourceConfigs, tr, timespec, recap.BuildOptions{EnrichDiffs: false}, createSource, os.Stderr)
		if err != nil {
			return err
		}

		// Apply source filter for style before any rendering.
		result.Entries = filterByStyle(result.Entries, style)

		if len(result.Entries) == 0 {
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleMuted.Render("No activity found for "+timespec))
			return nil
		}

		if recapJSON {
			return recap.RenderJSON(os.Stdout, result)
		}

		if recapRaw {
			return renderRaw(os.Stdout, result)
		}

		// Default: AI summary
		if !isAIConfigured(cfg) {
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleNormal.Render("anker needs an AI backend to generate readable reports."))
			_, _ = fmt.Fprintln(os.Stdout)
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleMuted.Render("Quick setup:"))
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleMuted.Render(`  anker config set ai_backend cli`))
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleMuted.Render(`  anker config set ai_cli_command "claude -p"`))
			_, _ = fmt.Fprintln(os.Stdout)
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleMuted.Render("Or run: anker init"))
			_, _ = fmt.Fprintln(os.Stdout)
			_, _ = fmt.Fprintln(os.Stdout, ui.StyleMuted.Render("For raw data without AI: anker recap --raw"))
			return nil
		}

		var buf bytes.Buffer
		if err := recap.RenderForAI(&buf, result); err != nil {
			return fmt.Errorf("failed to render recap: %w", err)
		}

		// Resolve prompt: --prompt flag > config ai_prompt > custom template file > style template
		promptOverride := recapPrompt
		if promptOverride == "" && cfg.AIPrompt == "" {
			if custom, found, err := loadCustomTemplate(string(style), lang); err != nil {
				return fmt.Errorf("failed to load custom template: %w", err)
			} else if found {
				promptOverride = custom
			} else {
				promptOverride = ai.PromptWithLanguage(style, lang)
			}
		}

		period := fmt.Sprintf("%s (%s to %s)", timespec, tr.From.Format("2006-01-02"), tr.To.Format("2006-01-02"))
		return ai.Transform(context.Background(), os.Stdout, buf.String(), period, ai.TransformConfig{
			AIPrompt:     cfg.AIPrompt,
			AIBackend:    cfg.AIBackend,
			AICLICommand: cfg.AICLICommand,
			AIBaseURL:    cfg.AIBaseURL,
			AIModel:      cfg.AIModel,
			AIAPIKey:     cfg.AIAPIKey,
			EntryCount:   len(result.Entries),
			Style:        string(style),
			Language:     lang,
		}, promptOverride, recapAPIKey)
	},
}

// parsedTemplate holds the result of parsing a .md template file.
type parsedTemplate struct {
	Description string
	Body        string
}

// parseTemplateFile extracts the YAML frontmatter description and prompt body
// from a .md template file. The expected format is:
//
//	---
//	description: Short description shown in --styles
//	---
//
//	## Prompt
//
//	Prompt text here...
//
// Frontmatter and the optional "## Prompt" heading are stripped from the body.
// Files without frontmatter are returned as-is with an empty description.
func parseTemplateFile(data []byte) parsedTemplate {
	lines := strings.Split(string(data), "\n")

	var description, body string

	// Detect frontmatter: first non-empty line must be "---" (after trimming).
	firstLine := ""
	if len(lines) > 0 {
		firstLine = strings.TrimSpace(lines[0])
	}

	if firstLine == "---" {
		// Find closing "---" marker.
		closeIdx := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				closeIdx = i
				break
			}
		}
		if closeIdx != -1 {
			for _, line := range lines[1:closeIdx] {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "description:") {
					description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
					description = strings.Trim(description, `"'`)
				}
			}
			body = strings.TrimSpace(strings.Join(lines[closeIdx+1:], "\n"))
		} else {
			body = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	} else {
		body = strings.TrimSpace(strings.Join(lines, "\n"))
	}

	// Strip optional "## Prompt" section heading.
	if strings.HasPrefix(body, "## Prompt\n") {
		body = strings.TrimSpace(strings.TrimPrefix(body, "## Prompt\n"))
	}

	return parsedTemplate{Description: description, Body: body}
}

// loadCustomTemplate looks for ~/.anker/templates/<name>.md, parses its
// frontmatter, and returns the prompt body with {language} injected.
// Returns (prompt, true, nil) if found, ("", false, nil) if not found.
func loadCustomTemplate(name, lang string) (string, bool, error) {
	home, err := paths.GetAnkerHome()
	if err != nil {
		return "", false, err
	}
	p := filepath.Join(home, "templates", name+".md")
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	tmpl := parseTemplateFile(data)
	prompt := strings.ReplaceAll(tmpl.Body, "{language}", lang)
	return prompt, true, nil
}

// runListStyles prints all available styles to w.
// If verbose is true, the full prompt body for each style is also shown.
func runListStyles(w *os.File, verbose bool) error {
	_, _ = fmt.Fprintln(w, ui.StyleSectionHeader.Render("Available styles:"))
	_, _ = fmt.Fprintln(w)

	infos := ai.StyleInfoList()
	maxLen := 0
	for _, s := range infos {
		if len(s.Name) > maxLen {
			maxLen = len(s.Name)
		}
	}

	defaultStyle := ai.StyleDigest
	for _, s := range infos {
		name := string(s.Name)
		pad := strings.Repeat(" ", maxLen-len(name))
		label := ui.StyleBold.Render(name) + pad
		desc := s.Description
		if s.Name == defaultStyle {
			desc += " " + ui.StyleMuted.Render("[default]")
		}
		_, _ = fmt.Fprintf(w, "  %s   %s\n", label, desc)

		if verbose {
			_, _ = fmt.Fprintln(w)
			for _, line := range strings.Split(ai.Prompt(s.Name), "\n") {
				_, _ = fmt.Fprintf(w, "    %s\n", ui.StyleMuted.Render(line))
			}
			_, _ = fmt.Fprintln(w)
		}
	}

	// List custom templates if any exist.
	if home, err := paths.GetAnkerHome(); err == nil {
		tmplDir := filepath.Join(home, "templates")
		if entries, err := os.ReadDir(tmplDir); err == nil {
			type customEntry struct {
				name string
				tmpl parsedTemplate
			}
			var customs []customEntry
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				name := strings.TrimSuffix(e.Name(), ".md")
				if data, err := os.ReadFile(filepath.Join(tmplDir, e.Name())); err == nil {
					customs = append(customs, customEntry{name, parseTemplateFile(data)})
				}
			}
			if len(customs) > 0 {
				_, _ = fmt.Fprintln(w)
				_, _ = fmt.Fprintln(w, ui.StyleSectionHeader.Render("Custom styles:"))
				_, _ = fmt.Fprintln(w)
				customMax := 0
				for _, c := range customs {
					if len(c.name) > customMax {
						customMax = len(c.name)
					}
				}
				for _, c := range customs {
					pad := strings.Repeat(" ", customMax-len(c.name))
					label := ui.StyleBold.Render(c.name) + pad
					desc := c.tmpl.Description
					if desc == "" {
						desc = "(no description)"
					}
					_, _ = fmt.Fprintf(w, "  %s   %s\n", label, desc)
					if verbose && c.tmpl.Body != "" {
						_, _ = fmt.Fprintln(w)
						for _, line := range strings.Split(c.tmpl.Body, "\n") {
							_, _ = fmt.Fprintf(w, "    %s\n", ui.StyleMuted.Render(line))
						}
						_, _ = fmt.Fprintln(w)
					}
				}
			}
		}
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, ui.StyleMuted.Render("Show full prompt: anker recap --styles --verbose"))
	_, _ = fmt.Fprintln(w, ui.StyleMuted.Render("Add custom style: ~/.anker/templates/<name>.md"))
	return nil
}

// resolveLanguage returns the effective output language from flag, config, or "deutsch".
func resolveLanguage(flagValue, configDefault string) string {
	if flagValue != "" {
		return flagValue
	}
	if configDefault != "" {
		return configDefault
	}
	return "deutsch"
}

// resolveStyle returns the effective style from flag, config default, or "self".
func resolveStyle(flagValue, configDefault string) ai.Style {
	if flagValue != "" {
		return ai.Style(strings.ToLower(flagValue))
	}
	if configDefault != "" {
		return ai.Style(strings.ToLower(configDefault))
	}
	return ai.StyleDigest
}

// filterByStyle removes entries whose source type is not allowed for the given style.
// If the style has no source restriction, the entries are returned unchanged.
func filterByStyle(entries []sources.Entry, style ai.Style) []sources.Entry {
	allowed := ai.AllowedSources(style)
	if len(allowed) == 0 {
		return entries
	}
	filtered := entries[:0]
	for _, e := range entries {
		if slices.Contains(allowed, e.Source) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// isAIConfigured reports whether the config has a usable AI backend.
func isAIConfigured(cfg *config.Config) bool {
	if cfg.AIBackend == "cli" {
		return cfg.AICLICommand != ""
	}
	// api backend: needs an API key (config or env)
	return cfg.AIAPIKey != "" || os.Getenv("AI_API_KEY") != ""
}

// renderRaw writes a plain-text entry dump, one line per entry, sorted by timestamp.
// Format: YYYY-MM-DD <source-label>: <content>
// Intended for pipes, grep, and debugging.
func renderRaw(w *os.File, result *recap.RecapResult) error {
	entries := make([]sources.Entry, len(result.Entries))
	copy(entries, result.Entries)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	for _, e := range entries {
		label := rawSourceLabel(e)
		_, _ = fmt.Fprintf(w, "%s %s: %s\n", e.Timestamp.Format("2006-01-02"), label, e.Content)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "%d entries (%s to %s)\n",
		len(entries),
		result.TimeRange.From.Format("2006-01-02"),
		result.TimeRange.To.Format("2006-01-02"))
	return nil
}

// rawSourceLabel returns "git/<repo>" for git sources and the source type for others.
func rawSourceLabel(e sources.Entry) string {
	if e.Source != "git" {
		return e.Source
	}
	name := e.Location
	if idx := len(name) - 1; idx >= 0 {
		for i := len(name) - 1; i >= 0; i-- {
			if name[i] == '/' {
				name = name[i+1:]
				break
			}
		}
	}
	return "git/" + name
}

func init() {
	rootCmd.AddCommand(recapCmd)
	recapCmd.Flags().StringVar(&recapPrompt, "prompt", "", "Custom prompt for AI summary")
	recapCmd.Flags().StringVar(&recapAPIKey, "api-key", "", "API key for AI summary")
	recapCmd.Flags().BoolVar(&recapRaw, "raw", false, "Unformatted entry dump -- for pipes, scripts, grep")
	recapCmd.Flags().BoolVar(&recapJSON, "json", false, "Structured JSON output")
	recapCmd.Flags().StringVar(&recapStyle, "style", "", "Summary style: brief, digest, status, report, retro")
	recapCmd.Flags().StringVar(&recapLang, "lang", "", "Report language passed to the AI model (e.g. deutsch, english, greek -- use full names, not ISO codes)")
	recapCmd.Flags().BoolVar(&recapStyles, "styles", false, "List available styles and exit")
	recapCmd.Flags().BoolVar(&recapVerbose, "verbose", false, "Show full prompt text when used with --styles")
}
