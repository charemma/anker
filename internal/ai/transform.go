package ai

import (
	"context"
	"fmt"
	"io"
	"os"
)

const defaultPrompt = `Summarize my workday based on the activity log below.
Start with the period as a heading (e.g. "# Recap: 2026-02-23 to 2026-03-01").
Group by topic or theme, not chronologically.
Keep it concise -- a few bullet points per topic.
Skip trivial entries (typo fixes, formatting, etc.) unless they are part of a larger change.
For each topic, also highlight:
- Decisions made and why (e.g. chose X over Y because...)
- Key insights or lessons learned
- Open threads or unfinished business
Only include these if they are actually present in the data -- don't invent them.`

// TransformConfig holds the AI-related fields needed by Transform.
// This avoids importing the config package directly.
type TransformConfig struct {
	AIPrompt     string
	AIBackend    string
	AICLICommand string
	AIBaseURL    string
	AIModel      string
	AIAPIKey     string
}

// Transform sends rendered recap text through an AI backend for summarization.
// It resolves the prompt (promptOverride > config > default) and API key
// (apiKeyOverride > AI_API_KEY env > config), then dispatches to CLI or API.
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

	if cfg.AIBackend == "cli" {
		return RunCLI(cfg.AICLICommand, prompt, renderedText, w)
	}

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

	return client.StreamCompletion(ctx, prompt, renderedText, w)
}
