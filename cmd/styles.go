package cmd

import "github.com/charmbracelet/lipgloss"

// Shared lipgloss styles for the CLI output.
// AdaptiveColor picks the light variant when the terminal background is light,
// and the dark variant when it is dark.
var (
	// styleHeader is used for section titles and step headings.
	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "14"})

	// styleSuccess is used for "✓ Added ..." confirmation lines.
	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "22", Dark: "10"})

	// styleMuted is used for hints, already-configured notes, and skip messages.
	styleMuted = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "8"})

	// styleBold is used for emphasis without color (welcome text, summary line).
	styleBold = lipgloss.NewStyle().Bold(true)
)
