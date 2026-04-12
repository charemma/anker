package recap

import (
	"fmt"
	"io"
)

// RenderMarkdown writes a full markdown recap with diffs.
func RenderMarkdown(w io.Writer, result *RecapResult) error {
	fmt.Fprintf(w, "# Work Recap\n\n")
	fmt.Fprintf(w, "**Period:** %s to %s\n", result.TimeRange.From.Format("2006-01-02"), result.TimeRange.To.Format("2006-01-02"))
	fmt.Fprintf(w, "**Total Activities:** %d\n\n", len(result.Entries))
	fmt.Fprintf(w, "---\n\n")
	fmt.Fprintf(w, "This recap contains git commits with full diffs for the specified period.\n")
	fmt.Fprintf(w, "Each commit includes the message and the complete code changes.\n\n")

	for _, group := range GroupByRepo(result.Entries) {
		switch group.Source {
		case "obsidian":
			fmt.Fprintf(w, "## Obsidian Vault\n\n")
			fmt.Fprintf(w, "**%s**\n\n", group.Name)
			fmt.Fprintf(w, "`%s`\n\n", group.Path)
		case "git":
			fmt.Fprintf(w, "## Git Repository: %s\n\n", group.Name)
			fmt.Fprintf(w, "`%s`\n\n", group.Path)
		case "markdown":
			fmt.Fprintf(w, "## Markdown Notes: %s\n\n", group.Name)
			fmt.Fprintf(w, "`%s`\n\n", group.Path)
		case "claude":
			fmt.Fprintf(w, "## Claude Sessions: %s\n\n", group.Name)
			fmt.Fprintf(w, "`%s`\n\n", group.Path)
		default:
			fmt.Fprintf(w, "## %s\n\n", group.Name)
			fmt.Fprintf(w, "`%s`\n\n", group.Path)
		}

		for i, entry := range group.Entries {
			fmt.Fprintf(w, "### %d.\n\n", i+1)
			fmt.Fprintf(w, "**Date:** %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))
			if author, ok := entry.Metadata["author"]; ok {
				fmt.Fprintf(w, "**Author:** %s\n", author)
			}
			if hash, ok := entry.Metadata["hash"]; ok {
				fmt.Fprintf(w, "**Hash:** `%s`\n", hash)
			}
			fmt.Fprintf(w, "**Message:** %s\n\n", entry.Content)

			if diff, ok := entry.Metadata["diff"]; ok && diff != "" {
				fmt.Fprintf(w, "**Changes:**\n\n```diff\n%s\n```\n\n", diff)
			} else {
				fmt.Fprintf(w, "*(No diff available)*\n\n")
			}

			fmt.Fprintf(w, "---\n\n")
		}
	}

	return nil
}
