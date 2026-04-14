package recap

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charemma/ikno/internal/sources"
	"github.com/charmbracelet/glamour"
	"github.com/mattn/go-isatty"
)

// maxAIPromptLength is the maximum length of a first_prompt in AI input.
// Longer prompts are truncated with "..." to save tokens.
const maxAIPromptLength = 100

// RenderForAI writes a markdown representation of the recap suitable as AI input.
// Claude session entries are condensed to metadata-only summaries to reduce
// token usage (typically 60-80% reduction for Claude-heavy recaps).
// Non-Claude entries are rendered identically to renderMarkdownRaw.
func RenderForAI(w io.Writer, result *RecapResult) error {
	_, _ = fmt.Fprintf(w, "# Work Recap\n\n")
	_, _ = fmt.Fprintf(w, "**Period:** %s to %s\n", result.TimeRange.From.Format("2006-01-02"), result.TimeRange.To.Format("2006-01-02"))
	_, _ = fmt.Fprintf(w, "**Total Activities:** %d\n\n", len(result.Entries))
	_, _ = fmt.Fprintf(w, "---\n\n")
	_, _ = fmt.Fprintf(w, "This recap contains git commits with full diffs for the specified period.\n")
	_, _ = fmt.Fprintf(w, "Each commit includes the message and the complete code changes.\n\n")

	for _, group := range GroupByRepo(result.Entries) {
		if group.Source == "claude" {
			renderClaudeGroupForAI(w, group)
			continue
		}

		renderGroupHeader(w, group)
		for i, entry := range group.Entries {
			renderEntryFull(w, entry, i+1)
		}
	}

	return nil
}

// RenderMarkdown writes a full markdown recap with diffs.
// When stdout is a terminal the output is rendered via glamour for readability.
func RenderMarkdown(w io.Writer, result *RecapResult) error {
	var buf bytes.Buffer
	if err := renderMarkdownRaw(&buf, result); err != nil {
		return err
	}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		rendered, err := glamour.Render(buf.String(), "auto")
		if err == nil {
			_, _ = fmt.Fprint(w, rendered)
			return nil
		}
	}

	// Fallback: plain markdown (also used when piped)
	_, _ = fmt.Fprint(w, buf.String())
	return nil
}

func renderMarkdownRaw(w io.Writer, result *RecapResult) error {
	_, _ = fmt.Fprintf(w, "# Work Recap\n\n")
	_, _ = fmt.Fprintf(w, "**Period:** %s to %s\n", result.TimeRange.From.Format("2006-01-02"), result.TimeRange.To.Format("2006-01-02"))
	_, _ = fmt.Fprintf(w, "**Total Activities:** %d\n\n", len(result.Entries))
	_, _ = fmt.Fprintf(w, "---\n\n")
	_, _ = fmt.Fprintf(w, "This recap contains git commits with full diffs for the specified period.\n")
	_, _ = fmt.Fprintf(w, "Each commit includes the message and the complete code changes.\n\n")

	for _, group := range GroupByRepo(result.Entries) {
		renderGroupHeader(w, group)
		for i, entry := range group.Entries {
			renderEntryFull(w, entry, i+1)
		}
	}

	return nil
}

// renderGroupHeader writes the section header for a source group.
func renderGroupHeader(w io.Writer, group RepoGroup) {
	switch group.Source {
	case "obsidian":
		_, _ = fmt.Fprintf(w, "## Obsidian Vault\n\n")
		_, _ = fmt.Fprintf(w, "**%s**\n\n", group.Name)
		_, _ = fmt.Fprintf(w, "`%s`\n\n", group.Path)
	case "git":
		_, _ = fmt.Fprintf(w, "## Git Repository: %s\n\n", group.Name)
		_, _ = fmt.Fprintf(w, "`%s`\n\n", group.Path)
	case "markdown":
		_, _ = fmt.Fprintf(w, "## Markdown Notes: %s\n\n", group.Name)
		_, _ = fmt.Fprintf(w, "`%s`\n\n", group.Path)
	case "claude":
		_, _ = fmt.Fprintf(w, "## Claude Sessions: %s\n\n", group.Name)
		_, _ = fmt.Fprintf(w, "`%s`\n\n", group.Path)
	default:
		_, _ = fmt.Fprintf(w, "## %s\n\n", group.Name)
		_, _ = fmt.Fprintf(w, "`%s`\n\n", group.Path)
	}
}

// renderEntryFull writes a single entry with all metadata and diff.
func renderEntryFull(w io.Writer, entry sources.Entry, num int) {
	_, _ = fmt.Fprintf(w, "### %d.\n\n", num)
	_, _ = fmt.Fprintf(w, "**Date:** %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))
	if author, ok := entry.Metadata["author"]; ok {
		_, _ = fmt.Fprintf(w, "**Author:** %s\n", author)
	}
	if hash, ok := entry.Metadata["hash"]; ok {
		_, _ = fmt.Fprintf(w, "**Hash:** `%s`\n", hash)
	}
	_, _ = fmt.Fprintf(w, "**Message:** %s\n\n", entry.Content)

	if diff, ok := entry.Metadata["diff"]; ok && diff != "" {
		_, _ = fmt.Fprintf(w, "**Changes:**\n\n```diff\n%s\n```\n\n", diff)
	} else {
		_, _ = fmt.Fprintf(w, "*(No diff available)*\n\n")
	}

	_, _ = fmt.Fprintf(w, "---\n\n")
}

// renderClaudeGroupForAI writes a condensed summary of Claude sessions for AI input.
// Instead of rendering each session with full prompt text, it outputs a compact
// table with just the key metadata: date, project, turns, duration, branch, and
// a truncated topic line.
func renderClaudeGroupForAI(w io.Writer, group RepoGroup) {
	_, _ = fmt.Fprintf(w, "## Claude Sessions (%d sessions)\n\n", len(group.Entries))

	for _, entry := range group.Entries {
		date := entry.Timestamp.Format("2006-01-02")
		project := entry.Metadata["project_name"]
		turns := entry.Metadata["turn_count"]
		duration := entry.Metadata["duration_minutes"]
		branch := entry.Metadata["git_branch"]

		// Build a short topic from the first prompt, truncated
		topic := truncatePrompt(entry.Metadata["first_prompt"], maxAIPromptLength)

		// Compact one-liner per session
		_, _ = fmt.Fprintf(w, "- %s **%s**", date, project)
		if turns != "" {
			_, _ = fmt.Fprintf(w, " (%s turns, %s min)", turns, duration)
		}
		if branch != "" {
			_, _ = fmt.Fprintf(w, " [%s]", branch)
		}
		if topic != "" {
			_, _ = fmt.Fprintf(w, ": %s", topic)
		}
		_, _ = fmt.Fprintf(w, "\n")
	}

	_, _ = fmt.Fprintf(w, "\n---\n\n")
}

// truncatePrompt shortens a prompt to maxLen characters, adding "..." if truncated.
// It also replaces newlines with spaces for compact display.
func truncatePrompt(prompt string, maxLen int) string {
	if prompt == "" {
		return ""
	}

	// Replace newlines with spaces for single-line display
	prompt = strings.ReplaceAll(prompt, "\n", " ")
	prompt = strings.Join(strings.Fields(prompt), " ")

	if len(prompt) <= maxLen {
		return prompt
	}
	return prompt[:maxLen] + "..."
}
