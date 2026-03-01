package markdown

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charemma/anker/internal/sources"
)

// MarkdownSource implements the Source interface for markdown files.
type MarkdownSource struct {
	basePath string
	tags     []string
	headings []string
}

// NewMarkdownSource creates a new Markdown source.
// tags: optional list of tags to filter (e.g., ["work", "done"])
// headings: optional list of heading patterns to extract from (e.g., ["## Work", "## Done"])
func NewMarkdownSource(basePath string, tags, headings []string) *MarkdownSource {
	return &MarkdownSource{
		basePath: basePath,
		tags:     tags,
		headings: headings,
	}
}

func (m *MarkdownSource) Type() string {
	return "markdown"
}

func (m *MarkdownSource) Location() string {
	return m.basePath
}

func (m *MarkdownSource) Validate() error {
	info, err := os.Stat(m.basePath)
	if err != nil {
		return fmt.Errorf("path not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}
	return nil
}

func (m *MarkdownSource) GetEntries(from, to time.Time) ([]sources.Entry, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}

	var entries []sources.Entry

	err := filepath.Walk(m.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Use file modification time as rough timestamp
		modTime := info.ModTime()
		if modTime.Before(from) || modTime.After(to) {
			return nil
		}

		fileEntries, err := m.extractEntries(path, modTime)
		if err != nil {
			return nil
		}

		entries = append(entries, fileEntries...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return entries, nil
}

func (m *MarkdownSource) extractEntries(path string, timestamp time.Time) ([]sources.Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var entries []sources.Entry
	var currentHeading string
	var inRelevantSection bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Track headings
		if strings.HasPrefix(line, "#") {
			currentHeading = line
			inRelevantSection = m.isRelevantHeading(line)
			continue
		}

		// Skip if we have heading filters and we're not in a relevant section
		if len(m.headings) > 0 && !inRelevantSection {
			continue
		}

		// Extract lines with tags
		if m.hasRelevantTags(line) {
			entry := sources.Entry{
				Timestamp: timestamp,
				Source:    "markdown",
				Location:  path,
				Content:   strings.TrimSpace(line),
				Metadata: map[string]string{
					"heading": currentHeading,
					"file":    filepath.Base(path),
				},
			}
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (m *MarkdownSource) isRelevantHeading(heading string) bool {
	if len(m.headings) == 0 {
		return true
	}

	for _, pattern := range m.headings {
		if strings.Contains(strings.ToLower(heading), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func (m *MarkdownSource) hasRelevantTags(line string) bool {
	// If no tag filter, include all non-empty lines
	if len(m.tags) == 0 {
		trimmed := strings.TrimSpace(line)
		return trimmed != "" && !strings.HasPrefix(trimmed, "#")
	}

	// Check for tags in format #tag or [[tag]]
	tagRegex := regexp.MustCompile(`#(\w+)|\[\[([^\]]+)\]\]`)
	matches := tagRegex.FindAllStringSubmatch(line, -1)

	for _, match := range matches {
		tag := match[1]
		if tag == "" {
			tag = match[2]
		}

		for _, filterTag := range m.tags {
			if strings.EqualFold(tag, filterTag) {
				return true
			}
		}
	}

	return false
}
