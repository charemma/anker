package recap

import (
	"encoding/json"
	"io"

	"github.com/charemma/anker/internal/sources"
)

// RenderJSON writes the recap as structured JSON.
func RenderJSON(w io.Writer, result *RecapResult) error {
	type JSONReport struct {
		Period struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"period"`
		Total      int             `json:"total"`
		Activities []sources.Entry `json:"activities"`
	}

	report := JSONReport{
		Total:      len(result.Entries),
		Activities: result.Entries,
	}
	report.Period.From = result.TimeRange.From.Format("2006-01-02")
	report.Period.To = result.TimeRange.To.Format("2006-01-02")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
