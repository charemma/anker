package markdown

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestMarkdownDir(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	return tmpDir
}

func writeMarkdownFile(t *testing.T, dir, filename, content string) string {
	t.Helper()

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMarkdownSource_Type(t *testing.T) {
	source := NewMarkdownSource("/path", nil, nil)
	if source.Type() != "markdown" {
		t.Errorf("expected type 'markdown', got %s", source.Type())
	}
}

func TestMarkdownSource_Location(t *testing.T) {
	path := "/path/to/notes"
	source := NewMarkdownSource(path, nil, nil)
	if source.Location() != path {
		t.Errorf("expected location %s, got %s", path, source.Location())
	}
}

func TestMarkdownSource_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() string
		expectErr bool
	}{
		{
			name: "valid directory",
			setup: func() string {
				return setupTestMarkdownDir(t)
			},
			expectErr: false,
		},
		{
			name: "non-existent directory",
			setup: func() string {
				return "/does/not/exist"
			},
			expectErr: true,
		},
		{
			name: "file instead of directory",
			setup: func() string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "file.txt")
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatal(err)
				}
				return filePath
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			source := NewMarkdownSource(path, nil, nil)

			err := source.Validate()
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMarkdownSource_GetEntries_NoTags(t *testing.T) {
	dir := setupTestMarkdownDir(t)

	content := `# My Notes

This is a regular line.
Another line of text.

## Work Section

Did some work today.
Fixed a bug.
`

	writeMarkdownFile(t, dir, "notes.md", content)

	source := NewMarkdownSource(dir, nil, nil)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(yesterday, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 4 {
		t.Fatalf("expected 4 entries (non-empty, non-heading lines), got %d", len(entries))
	}
}

func TestMarkdownSource_GetEntries_WithTags(t *testing.T) {
	dir := setupTestMarkdownDir(t)

	content := `# Daily Notes

Regular line without tag.
Line with #work tag.
Another line with #done marker.
Line with #work and #done tags.
Line without relevant tags #other.
`

	writeMarkdownFile(t, dir, "notes.md", content)

	source := NewMarkdownSource(dir, []string{"work", "done"}, nil)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(yesterday, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries with work/done tags, got %d", len(entries))
	}

	expectedContents := []string{
		"Line with #work tag.",
		"Another line with #done marker.",
		"Line with #work and #done tags.",
	}

	for i, entry := range entries {
		if entry.Source != "markdown" {
			t.Errorf("entry %d: expected source 'markdown', got %s", i, entry.Source)
		}

		if entry.Content != expectedContents[i] {
			t.Errorf("entry %d: expected content %q, got %q", i, expectedContents[i], entry.Content)
		}

		if entry.Metadata["file"] != "notes.md" {
			t.Errorf("entry %d: expected file 'notes.md', got %s", i, entry.Metadata["file"])
		}
	}
}

func TestMarkdownSource_GetEntries_WithHeadings(t *testing.T) {
	dir := setupTestMarkdownDir(t)

	content := `# Daily Notes

Regular line before any heading.

## Work

Work item 1.
Work item 2.

## Personal

Personal note.

## Work

More work items.
`

	writeMarkdownFile(t, dir, "notes.md", content)

	source := NewMarkdownSource(dir, nil, []string{"## Work"})

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(yesterday, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries under Work headings, got %d", len(entries))
	}

	for i, entry := range entries {
		heading := entry.Metadata["heading"]
		if heading != "## Work" {
			t.Errorf("entry %d: expected heading '## Work', got %s", i, heading)
		}
	}
}

func TestMarkdownSource_GetEntries_TimeFilter(t *testing.T) {
	dir := setupTestMarkdownDir(t)

	oldFile := writeMarkdownFile(t, dir, "old.md", "Old content #work")
	_ = writeMarkdownFile(t, dir, "new.md", "New content #work")

	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	source := NewMarkdownSource(dir, []string{"work"}, nil)

	now := time.Now()
	oneDayAgo := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(oneDayAgo, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry from recent file, got %d", len(entries))
	}

	if entries[0].Metadata["file"] != "new.md" {
		t.Errorf("expected file 'new.md', got %s", entries[0].Metadata["file"])
	}
}

func TestMarkdownSource_GetEntries_MultipleFiles(t *testing.T) {
	dir := setupTestMarkdownDir(t)

	writeMarkdownFile(t, dir, "file1.md", "Content from file 1 #work")
	writeMarkdownFile(t, dir, "file2.md", "Content from file 2 #work")
	writeMarkdownFile(t, dir, "file3.md", "Content from file 3 #work")

	source := NewMarkdownSource(dir, []string{"work"}, nil)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(yesterday, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries from 3 files, got %d", len(entries))
	}

	files := make(map[string]bool)
	for _, entry := range entries {
		files[entry.Metadata["file"]] = true
	}

	expectedFiles := []string{"file1.md", "file2.md", "file3.md"}
	for _, file := range expectedFiles {
		if !files[file] {
			t.Errorf("expected to find entry from %s", file)
		}
	}
}

func TestMarkdownSource_GetEntries_WikiLinkTags(t *testing.T) {
	dir := setupTestMarkdownDir(t)

	content := `# Notes

Line with [[work]] wikilink tag.
Line with [[done]] marker.
Line without relevant wikilinks [[other]].
`

	writeMarkdownFile(t, dir, "notes.md", content)

	source := NewMarkdownSource(dir, []string{"work", "done"}, nil)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(yesterday, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries with work/done wikilinks, got %d", len(entries))
	}
}

func TestMarkdownSource_GetEntries_SkipsNonMarkdownFiles(t *testing.T) {
	dir := setupTestMarkdownDir(t)

	writeMarkdownFile(t, dir, "notes.md", "Markdown content #work")

	txtPath := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(txtPath, []byte("Text content #work"), 0644); err != nil {
		t.Fatal(err)
	}

	source := NewMarkdownSource(dir, []string{"work"}, nil)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	entries, err := source.GetEntries(yesterday, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry from .md file only, got %d", len(entries))
	}

	if entries[0].Metadata["file"] != "notes.md" {
		t.Errorf("expected file 'notes.md', got %s", entries[0].Metadata["file"])
	}
}
