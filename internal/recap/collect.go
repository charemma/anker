package recap

import (
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/charemma/ikno/internal/sources"
	"github.com/charemma/ikno/internal/sources/git"
	"github.com/charemma/ikno/internal/timerange"
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
	// Collect entries from all sources concurrently.
	type sourceResult struct {
		source  sources.Source
		entries []sources.Entry
	}

	results := make([]sourceResult, len(sourceConfigs))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, cfg := range sourceConfigs {
		wg.Add(1)
		go func(idx int, cfg sources.Config) {
			defer wg.Done()

			source, err := factory(cfg)
			if err != nil {
				mu.Lock()
				_, _ = fmt.Fprintf(warn, "Warning: %v at %s\n", err, cfg.Path)
				mu.Unlock()
				return
			}

			entries, err := source.GetEntries(tr.From, tr.To)
			if err != nil {
				mu.Lock()
				_, _ = fmt.Fprintf(warn, "Warning: failed to get entries from %s %s: %v\n", cfg.Type, cfg.Path, err)
				mu.Unlock()
				return
			}

			// Enrich git entries with diffs when requested
			if opts.EnrichDiffs {
				if gs, ok := source.(*git.GitSource); ok {
					if err := gs.EnrichWithDiffs(entries); err != nil {
						mu.Lock()
						_, _ = fmt.Fprintf(warn, "Warning: failed to enrich diffs for %s: %v\n", gs.Location(), err)
						mu.Unlock()
					}
				}
			}

			results[idx] = sourceResult{source: source, entries: entries}
		}(i, cfg)
	}

	wg.Wait()

	var allEntries []sources.Entry
	for _, r := range results {
		allEntries = append(allEntries, r.entries...)
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
