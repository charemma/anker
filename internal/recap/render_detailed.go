package recap

import (
	"fmt"
	"io"
)

// RenderDetailed writes a recap with timestamps, authors, and commit hashes.
func RenderDetailed(w io.Writer, result *RecapResult) error {
	_, _ = fmt.Fprint(w, "\n")
	_, _ = fmt.Fprint(w, "Work Recap (Detailed)\n")
	_, _ = fmt.Fprint(w, "=====================\n")
	_, _ = fmt.Fprintf(w, "Period: %s - %s\n", result.TimeRange.From.Format("02 Jan 2006"), result.TimeRange.To.Format("02 Jan 2006"))
	_, _ = fmt.Fprintf(w, "Total: %d activities\n\n", len(result.Entries))

	for _, group := range GroupByRepo(result.Entries) {
		switch group.Source {
		case "obsidian":
			_, _ = fmt.Fprintf(w, "Obsidian Vault\n")
			_, _ = fmt.Fprintf(w, "%s\n\n", group.Name)
		case "git":
			_, _ = fmt.Fprintf(w, "Git Repository: %s\n", group.Name)
			_, _ = fmt.Fprintf(w, "(%s)\n\n", group.Path)
		case "markdown":
			_, _ = fmt.Fprintf(w, "Markdown Notes: %s\n", group.Name)
			_, _ = fmt.Fprintf(w, "(%s)\n\n", group.Path)
		case "claude":
			_, _ = fmt.Fprintf(w, "Claude Sessions: %s\n", group.Name)
			_, _ = fmt.Fprintf(w, "(%s)\n\n", group.Path)
		default:
			_, _ = fmt.Fprintf(w, "%s\n\n", group.Name)
		}

		for _, entry := range group.Entries {
			_, _ = fmt.Fprintf(w, "  %s\n", entry.Timestamp.Format("Mon Jan 2, 15:04"))
			_, _ = fmt.Fprintf(w, "  %s\n", entry.Content)
			if author, ok := entry.Metadata["author"]; ok {
				_, _ = fmt.Fprintf(w, "  Author: %s\n", author)
			}
			if hash, ok := entry.Metadata["hash"]; ok {
				_, _ = fmt.Fprintf(w, "  Commit: %s\n", hash[:8])
			}
			if slug, ok := entry.Metadata["slug"]; ok {
				_, _ = fmt.Fprintf(w, "  Session: %s\n", slug)
			}
			_, _ = fmt.Fprintln(w)
		}
	}

	return nil
}
