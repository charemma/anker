package sources

import "time"

// Entry represents a single activity entry from any source.
type Entry struct {
	Timestamp time.Time
	Source    string // "git", "markdown", etc.
	Location  string // repo path, file path, etc.
	Content   string // commit message, note content, etc.
	Metadata  map[string]string
}

// Source is the interface that all data sources must implement.
type Source interface {
	// Type returns the source type identifier (git, markdown, calendar, etc.)
	Type() string

	// GetEntries retrieves all entries within the given time range.
	GetEntries(from, to time.Time) ([]Entry, error)

	// Validate checks if the source configuration is valid and accessible.
	Validate() error

	// Location returns the primary location/path of this source.
	Location() string
}

// Config represents a source configuration as stored in sources.yaml
type Config struct {
	Type     string            `yaml:"type"`
	Path     string            `yaml:"path"`
	Added    time.Time         `yaml:"added"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
}
