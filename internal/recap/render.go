package recap

import (
	"fmt"
	"io"
)

// Render dispatches to the appropriate renderer based on format string.
// plain disables ANSI styling (honoured by renderers that support it).
func Render(w io.Writer, result *RecapResult, format string, plain bool) error {
	switch format {
	case "simple", "":
		return RenderSummary(w, result, plain)
	case "detailed":
		return RenderDetailed(w, result)
	case "json":
		return RenderJSON(w, result)
	case "markdown":
		return RenderMarkdown(w, result)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}
