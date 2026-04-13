package recap

import (
	"github.com/charemma/ikno/internal/sources"
	"github.com/charemma/ikno/internal/timerange"
)

// RecapResult holds the collected activity data for a time period.
// It is the data contract between collection (BuildRecap) and rendering.
type RecapResult struct {
	TimeRange *timerange.TimeRange
	Timespec  string
	Entries   []sources.Entry
}
