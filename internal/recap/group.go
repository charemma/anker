package recap

import (
	"sort"
	"strings"

	"github.com/charemma/ikno/internal/sources"
)

// RepoGroup represents entries from a single source location, with
// pre-computed display name and source type for rendering.
type RepoGroup struct {
	Path    string
	Name    string
	Source  string
	Entries []sources.Entry
}

// GroupByRepo groups entries by location, extracts display names,
// and returns groups sorted alphabetically by path.
func GroupByRepo(entries []sources.Entry) []RepoGroup {
	byRepo := make(map[string][]sources.Entry)
	for _, entry := range entries {
		byRepo[entry.Location] = append(byRepo[entry.Location], entry)
	}

	paths := make([]string, 0, len(byRepo))
	for path := range byRepo {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	groups := make([]RepoGroup, 0, len(paths))
	for _, path := range paths {
		repoEntries := byRepo[path]
		name := path
		if idx := strings.LastIndex(name, "/"); idx != -1 {
			name = name[idx+1:]
		}

		source := ""
		if len(repoEntries) > 0 {
			source = repoEntries[0].Source
		}

		groups = append(groups, RepoGroup{
			Path:    path,
			Name:    name,
			Source:  source,
			Entries: repoEntries,
		})
	}

	return groups
}

// SourceLabel returns a human-readable label for a source type.
// For unknown types, returns the type name with a capitalized first letter.
func SourceLabel(sourceType string) string {
	switch sourceType {
	case "git":
		return "Git Repository"
	case "obsidian":
		return "Obsidian Vault"
	case "markdown":
		return "Markdown Notes"
	case "claude":
		return "Claude Sessions"
	default:
		if sourceType == "" {
			return ""
		}
		return strings.ToUpper(sourceType[:1]) + sourceType[1:]
	}
}
