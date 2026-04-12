package recap

import (
	"fmt"
	"io"
)

// RenderSimple writes a compact recap with commit messages only.
func RenderSimple(w io.Writer, result *RecapResult) error {
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Work Recap\n")
	fmt.Fprintf(w, "==========\n")
	fmt.Fprintf(w, "Period: %s - %s\n", result.TimeRange.From.Format("02 Jan 2006"), result.TimeRange.To.Format("02 Jan 2006"))
	fmt.Fprintf(w, "Total: %d activities\n\n", len(result.Entries))

	for _, group := range GroupByRepo(result.Entries) {
		switch group.Source {
		case "obsidian":
			fmt.Fprintf(w, "Obsidian Vault\n")
			fmt.Fprintf(w, "%s\n\n", group.Name)
		case "git":
			fmt.Fprintf(w, "Git Repository: %s\n", group.Name)
			fmt.Fprintf(w, "(%s)\n\n", group.Path)
		case "markdown":
			fmt.Fprintf(w, "Markdown Notes: %s\n", group.Name)
			fmt.Fprintf(w, "(%s)\n\n", group.Path)
		case "claude":
			fmt.Fprintf(w, "Claude Sessions: %s\n", group.Name)
			fmt.Fprintf(w, "(%s)\n\n", group.Path)
		default:
			fmt.Fprintf(w, "%s\n\n", group.Name)
		}

		for _, entry := range group.Entries {
			fmt.Fprintf(w, "  \u2022 %s\n", entry.Content)
		}
		fmt.Fprintln(w)
	}

	return nil
}
