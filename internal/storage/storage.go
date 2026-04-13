package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charemma/ikno/internal/paths"
	"github.com/charemma/ikno/internal/sources"
	"gopkg.in/yaml.v3"
)

type SourceRegistry struct {
	Sources []sources.Config `yaml:"sources"`
}

type Store struct {
	baseDir string
}

// NewStore creates a new Store instance.
// The base directory can be set via IKNO_HOME environment variable,
// otherwise defaults to ~/.config/ikno/
func NewStore() (*Store, error) {
	baseDir, err := paths.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get ikno config directory: %w", err)
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Store{baseDir: baseDir}, nil
}

// AddSource adds or updates a source in the registry.
func (s *Store) AddSource(config sources.Config) error {
	sourcesPath := filepath.Join(s.baseDir, "sources.yaml")

	registry, err := s.loadSourceRegistry(sourcesPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load source registry: %w", err)
	}

	config.Added = time.Now()
	found := false

	for i := range registry.Sources {
		if registry.Sources[i].Type == config.Type && registry.Sources[i].Path == config.Path {
			registry.Sources[i] = config
			found = true
			break
		}
	}

	if !found {
		registry.Sources = append(registry.Sources, config)
	}

	return s.saveSourceRegistry(sourcesPath, registry)
}

// GetSources returns all configured sources.
func (s *Store) GetSources() ([]sources.Config, error) {
	sourcesPath := filepath.Join(s.baseDir, "sources.yaml")

	registry, err := s.loadSourceRegistry(sourcesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []sources.Config{}, nil
		}
		return nil, fmt.Errorf("failed to load source registry: %w", err)
	}

	return registry.Sources, nil
}

// RemoveSource removes a source from the registry by type and path.
func (s *Store) RemoveSource(sourceType, path string) error {
	sourcesPath := filepath.Join(s.baseDir, "sources.yaml")

	registry, err := s.loadSourceRegistry(sourcesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to load source registry: %w", err)
	}

	filtered := make([]sources.Config, 0)
	for _, src := range registry.Sources {
		if src.Type != sourceType || src.Path != path {
			filtered = append(filtered, src)
		}
	}

	registry.Sources = filtered
	return s.saveSourceRegistry(sourcesPath, registry)
}

// RemoveSourceByPath removes a source from the registry by path only.
// If multiple sources exist with the same path, returns an error.
func (s *Store) RemoveSourceByPath(path string) (*sources.Config, error) {
	sourcesPath := filepath.Join(s.baseDir, "sources.yaml")

	registry, err := s.loadSourceRegistry(sourcesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no sources configured")
		}
		return nil, fmt.Errorf("failed to load source registry: %w", err)
	}

	// Find matching sources
	var matches []sources.Config
	var filtered []sources.Config
	for _, src := range registry.Sources {
		if src.Path == path {
			matches = append(matches, src)
		} else {
			filtered = append(filtered, src)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no source found for path: %s", path)
	}

	if len(matches) > 1 {
		types := make([]string, len(matches))
		for i, m := range matches {
			types[i] = m.Type
		}
		return nil, fmt.Errorf("multiple sources found for path %s (%v). Use 'ikno source remove <type> <path>' to specify", path, types)
	}

	registry.Sources = filtered
	if err := s.saveSourceRegistry(sourcesPath, registry); err != nil {
		return nil, err
	}

	return &matches[0], nil
}

func (s *Store) loadSourceRegistry(path string) (*SourceRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &SourceRegistry{}, err
	}

	var registry SourceRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse source registry: %w", err)
	}

	return &registry, nil
}

func (s *Store) saveSourceRegistry(path string, registry *SourceRegistry) error {
	data, err := yaml.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal source registry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write source registry: %w", err)
	}

	return nil
}
