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
	repoPath     string
	authorEmails []string // optional: if set, only return commits from these authors
}

// NewGitSource creates a new Git source for the given repository path.
// If authorEmails is provided (comma-separated), only commits from those authors will be included.
func NewGitSource(repoPath, authorEmails string) *GitSource {
	var emails []string
	if authorEmails != "" {
		// Split comma-separated emails and trim whitespace
		for _, email := range strings.Split(authorEmails, ",") {
			trimmed := strings.TrimSpace(email)
			if trimmed != "" {
				emails = append(emails, trimmed)
			}
		}
	}
	return &GitSource{
		repoPath:     repoPath,
		authorEmails: emails,
	}
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

		// Filter by author email if specified
		if len(g.authorEmails) > 0 {
			found := false
			for _, email := range g.authorEmails {
				if parts[2] == email {
					found = true
					break
				}
			}
			if !found {
				continue
			}
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

// GetDiff retrieves the full diff for a commit hash.
// Returns the diff as a string, or empty string if the diff cannot be retrieved.
func (g *GitSource) GetDiff(commitHash string) (string, error) {
	cmd := exec.Command("git", "-C", g.repoPath, "show", "--format=", "--no-color", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff for %s: %w", commitHash, err)
	}
	return string(output), nil
}

// EnrichWithDiffs adds diff information to entries that have a commit hash in metadata.
// Modifies entries in place by adding a "diff" field to their Metadata.
func (g *GitSource) EnrichWithDiffs(entries []sources.Entry) error {
	for i := range entries {
		if hash, ok := entries[i].Metadata["hash"]; ok {
			diff, err := g.GetDiff(hash)
			if err != nil {
				// Log warning but don't fail - continue with other entries
				continue
			}
			entries[i].Metadata["diff"] = diff
		}
	}
	return nil
}
