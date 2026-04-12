package ui

import "github.com/charmbracelet/lipgloss"

// Source-type colors. One color per type, used consistently across all renderers.
var (
	ColorGit      = lipgloss.AdaptiveColor{Dark: "#82AAFF", Light: "#1A56DB"} // blue
	ColorObsidian = lipgloss.AdaptiveColor{Dark: "#C3E88D", Light: "#2E7D32"} // green
	ColorMarkdown = lipgloss.AdaptiveColor{Dark: "#FFCB6B", Light: "#B45309"} // amber
	ColorClaude   = lipgloss.AdaptiveColor{Dark: "#F78C6C", Light: "#C05621"} // orange

	ColorDay    = lipgloss.AdaptiveColor{Dark: "#666666", Light: "#999999"} // dimmed grey
	ColorMuted  = lipgloss.AdaptiveColor{Dark: "#4A5568", Light: "#9CA3AF"} // very dimmed
	ColorNormal = lipgloss.AdaptiveColor{Dark: "#E2E8F0", Light: "#1A202C"} // standard text
)

// SourceColor returns the AdaptiveColor for the given source type.
// Unknown types fall back to ColorNormal.
func SourceColor(sourceType string) lipgloss.AdaptiveColor {
	switch sourceType {
	case "git":
		return ColorGit
	case "obsidian":
		return ColorObsidian
	case "markdown":
		return ColorMarkdown
	case "claude":
		return ColorClaude
	default:
		return ColorNormal
	}
}

// StyleHeader renders bold text in the day color.
var StyleHeader = lipgloss.NewStyle().Bold(true)

// StyleMuted renders text in the muted color.
var StyleMuted = lipgloss.NewStyle().Foreground(ColorMuted)

// StyleDay renders date headers in the day color.
var StyleDay = lipgloss.NewStyle().Foreground(ColorDay)

// StyleNormal renders standard output text.
var StyleNormal = lipgloss.NewStyle().Foreground(ColorNormal)

// StyleSuccess renders confirmation lines.
var StyleSuccess = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "22", Dark: "10"})

// StyleBold renders bold text without color.
var StyleBold = lipgloss.NewStyle().Bold(true)

// StyleSectionHeader renders section titles and step headings.
var StyleSectionHeader = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "14"})
