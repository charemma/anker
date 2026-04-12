package recap

import (
	"fmt"
	"io"
)

// Render dispatches to the appropriate renderer based on format string.
func Render(w io.Writer, result *RecapResult, format string) error {
	switch format {
	case "simple":
		return RenderSimple(w, result)
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
