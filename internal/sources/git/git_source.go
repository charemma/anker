package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charemma/anker/internal/sources"
)

// GitSource implements the Source interface for git repositories.
type GitSource struct {
	repoPath string
}

// NewGitSource creates a new Git source for the given repository path.
func NewGitSource(repoPath string) *GitSource {
	return &GitSource{repoPath: repoPath}
}

func (g *GitSource) Type() string {
	return "git"
}

func (g *GitSource) Location() string {
	return g.repoPath
}

func (g *GitSource) Validate() error {
	cmd := exec.Command("git", "-C", g.repoPath, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a valid git repository: %w", err)
	}
	return nil
}

func (g *GitSource) GetEntries(from, to time.Time) ([]sources.Entry, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}

	// Format: %H|%an|%ae|%at|%s
	// Hash|Author Name|Author Email|Timestamp|Subject
	format := "--pretty=format:%H|%an|%ae|%at|%s"
	since := fmt.Sprintf("--since=%s", from.Format(time.RFC3339))
	until := fmt.Sprintf("--until=%s", to.Format(time.RFC3339))

	cmd := exec.Command("git", "-C", g.repoPath, "log", format, since, until, "--all")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []sources.Entry{}, nil
	}

	entries := make([]sources.Entry, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 5)
		if len(parts) != 5 {
			continue
		}

		timestamp, err := parseUnixTimestamp(parts[3])
		if err != nil {
			continue
		}

		// Client-side filtering to ensure we only include commits within the time range
		if timestamp.Before(from) || timestamp.After(to) {
			continue
		}

		entry := sources.Entry{
			Timestamp: timestamp,
			Source:    "git",
			Location:  g.repoPath,
			Content:   parts[4], // commit subject
			Metadata: map[string]string{
				"hash":   parts[0],
				"author": parts[1],
				"email":  parts[2],
			},
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func parseUnixTimestamp(s string) (time.Time, error) {
	var timestamp int64
	_, err := fmt.Sscanf(s, "%d", &timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(timestamp, 0), nil
}
