package recap

import (
	"fmt"
	"io"
	"sort"

	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/sources/git"
	"github.com/charemma/anker/internal/timerange"
)

// SourceFactory creates a Source from a stored Config.
// Injected by the command layer to avoid a circular dependency on cmd/.
type SourceFactory func(sources.Config) (sources.Source, error)

// BuildOptions controls optional behaviour during recap collection.
type BuildOptions struct {
	EnrichDiffs bool // fetch full diffs for git sources
}

// BuildRecap collects entries from all configured sources for the given time
// range. When opts.EnrichDiffs is true, git sources are enriched with diffs.
// Warnings about individual source failures are written to warn.
func BuildRecap(sourceConfigs []sources.Config, tr *timerange.TimeRange, timespec string, opts BuildOptions, factory SourceFactory, warn io.Writer) (*RecapResult, error) {
	var allEntries []sources.Entry

	for _, cfg := range sourceConfigs {
		source, err := factory(cfg)
		if err != nil {
			_, _ = fmt.Fprintf(warn, "Warning: %v at %s\n", err, cfg.Path)
			continue
		}

		entries, err := source.GetEntries(tr.From, tr.To)
		if err != nil {
			_, _ = fmt.Fprintf(warn, "Warning: failed to get entries from %s %s: %v\n", cfg.Type, cfg.Path, err)
			continue
		}

		// Enrich git entries with diffs when requested
		if opts.EnrichDiffs {
			if gs, ok := source.(*git.GitSource); ok {
				if err := gs.EnrichWithDiffs(entries); err != nil {
					_, _ = fmt.Fprintf(warn, "Warning: failed to enrich diffs for %s: %v\n", gs.Location(), err)
				}
			}
		}

		allEntries = append(allEntries, entries...)
	}

	// Sort entries by timestamp (newest first)
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Timestamp.After(allEntries[j].Timestamp)
	})

	return &RecapResult{
		TimeRange: tr,
		Timespec:  timespec,
		Entries:   allEntries,
	}, nil
}
