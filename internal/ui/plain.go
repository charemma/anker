package ui

import (
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// IsPlain reports whether output should be plain (no ANSI, no lipgloss).
// Returns true when --plain flag is set, NO_COLOR env var is non-empty,
// or stdout is not a terminal.
func IsPlain(cmd *cobra.Command) bool {
	plain, _ := cmd.Flags().GetBool("plain")
	return plain ||
		os.Getenv("NO_COLOR") != "" ||
		!isatty.IsTerminal(os.Stdout.Fd())
}

// TermWidth returns the terminal width, capped at 120.
// Falls back to 80 when stdout is not a terminal or width cannot be determined.
func TermWidth() int {
	w, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		return 80
	}
	if w > 120 {
		return 120
	}
	return w
}
