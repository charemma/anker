package recap

import (
	"fmt"
	"io"
)

// RenderMarkdown writes a full markdown recap with diffs.
func RenderMarkdown(w io.Writer, result *RecapResult) error {
	_, _ = fmt.Fprintf(w, "# Work Recap\n\n")
	_, _ = fmt.Fprintf(w, "**Period:** %s to %s\n", result.TimeRange.From.Format("2006-01-02"), result.TimeRange.To.Format("2006-01-02"))
	_, _ = fmt.Fprintf(w, "**Total Activities:** %d\n\n", len(result.Entries))
	_, _ = fmt.Fprintf(w, "---\n\n")
	_, _ = fmt.Fprintf(w, "This recap contains git commits with full diffs for the specified period.\n")
	_, _ = fmt.Fprintf(w, "Each commit includes the message and the complete code changes.\n\n")

	for _, group := range GroupByRepo(result.Entries) {
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

		for i, entry := range group.Entries {
			_, _ = fmt.Fprintf(w, "### %d.\n\n", i+1)
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
	}

	return nil
}
