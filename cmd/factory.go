package cmd

import (
	"fmt"
	"strings"

	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/sources/claude"
	"github.com/charemma/anker/internal/sources/git"
	"github.com/charemma/anker/internal/sources/markdown"
	"github.com/charemma/anker/internal/sources/obsidian"
)

// createSource instantiates a Source from a stored Config.
func createSource(cfg sources.Config) (sources.Source, error) {
	switch cfg.Type {
	case "git":
		authorEmail := cfg.Metadata["author"]
		return git.NewGitSource(cfg.Path, authorEmail), nil
	case "markdown":
		tags := splitTrimmed(cfg.Metadata["tags"], ",")
		headings := splitTrimmed(cfg.Metadata["headings"], ",")
		return markdown.NewMarkdownSource(cfg.Path, tags, headings), nil
	case "obsidian":
		return obsidian.NewObsidianSource(cfg.Path), nil
	case "claude":
		return claude.NewClaudeSource(cfg.Path), nil
	default:
		return nil, fmt.Errorf("unsupported source type: %s", cfg.Type)
	}
}

func splitTrimmed(s, sep string) []string {
	var parts []string
	for part := range strings.SplitSeq(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
