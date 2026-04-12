package aisession

import (
	"math"
	"time"
)

// ToolInvocation represents a single tool/function call made by the AI assistant.
type ToolInvocation struct {
	Name string // "Read", "Edit", "Bash", etc.
	Path string // file path if applicable, empty otherwise
}

// SessionSummary holds aggregated metadata for one AI coding session.
// Sources produce these; the caller converts them to sources.Entry.
type SessionSummary struct {
	SessionID   string
	Slug        string
	Project     string           // decoded project name (human-readable)
	ProjectDir  string           // raw encoded dir name (for metadata)
	FirstPrompt string           // full text of the first real user message
	TurnCount   int              // number of real user messages (excludes tool_result, meta)
	Model       string           // primary model used (from first assistant message)
	CWD         string           // working directory
	GitBranch   string           // branch name if available
	StartTime   time.Time        // earliest timestamp in session
	EndTime     time.Time        // latest timestamp in session
	ToolsUsed   []ToolInvocation // deduplicated list of tools invoked
}

// DurationMinutes returns the session duration rounded to the nearest minute.
func (s SessionSummary) DurationMinutes() int {
	return int(math.Round(s.EndTime.Sub(s.StartTime).Minutes()))
}
