package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charemma/anker/internal/sources"
)

// ObsidianSource implements the Source interface for Obsidian vaults.
// It tracks which markdown files were created or modified in a time range.
type ObsidianSource struct {
	vaultPath string
}

// NewObsidianSource creates a new Obsidian vault source.
func NewObsidianSource(vaultPath string) *ObsidianSource {
	return &ObsidianSource{
		vaultPath: vaultPath,
	}
}

func (o *ObsidianSource) Type() string {
	return "obsidian"
}

func (o *ObsidianSource) Location() string {
	return o.vaultPath
}

func (o *ObsidianSource) Validate() error {
	info, err := os.Stat(o.vaultPath)
	if err != nil {
		return fmt.Errorf("vault path not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("vault path is not a directory")
	}

	// Check if .obsidian folder exists (indicates it's an Obsidian vault)
	obsidianDir := filepath.Join(o.vaultPath, ".obsidian")
	if _, err := os.Stat(obsidianDir); err != nil {
		return fmt.Errorf("not an Obsidian vault (missing .obsidian folder)")
	}

	return nil
}

func (o *ObsidianSource) GetEntries(from, to time.Time) ([]sources.Entry, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	var entries []sources.Entry

	err := filepath.Walk(o.vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip .obsidian and .trash folders
			if info.Name() == ".obsidian" || info.Name() == ".trash" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		modTime := info.ModTime()

		// Check if file was modified in the time range
		if modTime.Before(from) || modTime.After(to) {
			return nil
		}

		// Get relative path within vault
		relPath, err := filepath.Rel(o.vaultPath, path)
		if err != nil {
			relPath = path
		}

		// Determine if file was created or modified in this period
		// We consider it "created" if mtime is very close to the period start,
		// otherwise "modified"
		action := "modified"
		if modTime.Sub(from) < 5*time.Minute {
			action = "created"
		}

		entry := sources.Entry{
			Timestamp: modTime,
			Source:    "obsidian",
			Location:  o.vaultPath,
			Content:   fmt.Sprintf("%s (%s)", relPath, action),
			Metadata: map[string]string{
				"file":   filepath.Base(path),
				"path":   relPath,
				"action": action,
			},
		}

		entries = append(entries, entry)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk vault: %w", err)
	}

	return entries, nil
}
