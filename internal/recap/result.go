package recap

import (
	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/timerange"
)

// RecapResult holds the collected activity data for a time period.
// It is the data contract between collection (BuildRecap) and rendering.
type RecapResult struct {
	TimeRange *timerange.TimeRange
	Timespec  string
	Entries   []sources.Entry
}
