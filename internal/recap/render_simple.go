package recap

import (
	"fmt"
	"io"
)

// RenderSimple writes a compact recap with commit messages only.
func RenderSimple(w io.Writer, result *RecapResult) error {
	_, _ = fmt.Fprintf(w, "\n")
	_, _ = fmt.Fprintf(w, "Work Recap\n")
	_, _ = fmt.Fprintf(w, "==========\n")
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
			_, _ = fmt.Fprintf(w, "  \u2022 %s\n", entry.Content)
		}
		_, _ = fmt.Fprintln(w)
	}

	return nil
}
