package ai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charemma/ikno/internal/ui"
	"github.com/charmbracelet/glamour"
	"github.com/mattn/go-isatty"
)

// defaultPrompt is a structured template that produces a consistent summary format.
const defaultPrompt = `Write a weekly work summary in this exact structure:

## Was wurde fertiggestellt
[2-4 bullet points, completed things only]

## Womit wurde die meiste Zeit verbracht
[1-2 sentences, focus and themes]

## Was ist noch offen
[Optional section, only if clearly visible from the data]

Rules:
- Max 12 lines total
- No individual commit listings
- No timestamps or hashes
- Language: German
- Be concrete, not generic`

// TransformConfig holds the AI-related fields needed by Transform.
// This avoids importing the config package directly.
type TransformConfig struct {
	AIPrompt     string
	AIBackend    string
	AICLICommand string
	AIBaseURL    string
	AIModel      string
	AIAPIKey     string
	EntryCount   int    // shown in the footer line
	Style        string // shown in the status line
	Language     string // shown in the status line
}

// Transform sends rendered recap text through an AI backend for summarization.
// It resolves the prompt (promptOverride > config > default) and API key
// (apiKeyOverride > AI_API_KEY env > config), then dispatches to CLI or API.
// The AI response is buffered and glamour-rendered when stdout is a terminal.
// Status messages are written to stderr so they disappear when stdout is piped.
func Transform(ctx context.Context, w io.Writer, renderedText string, period string, cfg TransformConfig, promptOverride, apiKeyOverride string) error {
	// Resolve prompt: override > config > default
	prompt := defaultPrompt
	if cfg.AIPrompt != "" {
		prompt = cfg.AIPrompt
	}
	if promptOverride != "" {
		prompt = promptOverride
	}

	// Inject time range context
	prompt = fmt.Sprintf("Period: %s\n\n%s", period, prompt)

	// Spinner on stderr -- invisible when piped, clears when done
	spinnerMsg := "Generating summary..."
	if cfg.Style != "" || cfg.Language != "" {
		details := fmt.Sprintf("style: %s, lang: %s", cfg.Style, cfg.Language)
		spinnerMsg = "Generating summary... " + ui.StyleMuted.Render("("+details+")")
	}
	stopSpinner := ui.StartSpinner(spinnerMsg)

	// Buffer AI output so we can glamour-render it
	var aiOut bytes.Buffer

	var err error
	if cfg.AIBackend == "cli" {
		err = RunCLI(cfg.AICLICommand, prompt, renderedText, &aiOut)
	} else {
		// Resolve API key: override > env > config
		apiKey := cfg.AIAPIKey
		if envKey := os.Getenv("AI_API_KEY"); envKey != "" {
			apiKey = envKey
		}
		if apiKeyOverride != "" {
			apiKey = apiKeyOverride
		}

		client := &Client{
			BaseURL: cfg.AIBaseURL,
			APIKey:  apiKey,
			Model:   cfg.AIModel,
		}
		err = client.StreamCompletion(ctx, prompt, renderedText, &aiOut)
	}

	stopSpinner()

	if err != nil {
		return err
	}

	sep := strings.Repeat("─", 65)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, sep)
	_, _ = fmt.Fprintln(w)

	aiText := aiOut.String()
	// Stats style has pre-formatted ASCII bars -- glamour would break the layout.
	// Only glamour-render prose styles (digest, report, brief, etc.)
	if isatty.IsTerminal(os.Stdout.Fd()) && cfg.Style != "stats" {
		rendered, renderErr := glamour.Render(aiText, "auto")
		if renderErr == nil {
			_, _ = fmt.Fprint(w, rendered)
		} else {
			_, _ = fmt.Fprint(w, aiText)
		}
	} else {
		_, _ = fmt.Fprint(w, aiText)
	}

	_, _ = fmt.Fprintln(w, sep)

	// Footer: entry count + model + period
	model := cfg.AIModel
	if model == "" {
		model = "ai"
	}
	footerParts := []string{}
	if cfg.EntryCount > 0 {
		footerParts = append(footerParts, fmt.Sprintf("Generated from %d entries", cfg.EntryCount))
	} else {
		footerParts = append(footerParts, "Generated")
	}
	footerParts = append(footerParts, model)
	footerParts = append(footerParts, period)
	_, _ = fmt.Fprintln(w, strings.Join(footerParts, " · "))

	return nil
}
