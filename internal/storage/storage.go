package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charemma/anker/internal/sources"
	"gopkg.in/yaml.v3"
)

type Repo struct {
	Path     string    `yaml:"path"`
	LastSeen time.Time `yaml:"last_seen"`
}

type RepoRegistry struct {
	Repos []Repo `yaml:"repos"`
}

type SourceRegistry struct {
	Sources []sources.Config `yaml:"sources"`
}

type Store struct {
	baseDir string
}

// NewStore creates a new Store instance.
// The base directory defaults to ~/.anker
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(home, ".anker")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Store{baseDir: baseDir}, nil
}

// TrackRepo adds or updates a repository in the registry.
func (s *Store) TrackRepo(repoPath string) error {
	registryPath := filepath.Join(s.baseDir, "repos.yaml")

	registry, err := s.loadRegistry(registryPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	now := time.Now()
	found := false

	for i := range registry.Repos {
		if registry.Repos[i].Path == repoPath {
			registry.Repos[i].LastSeen = now
			found = true
			break
		}
	}

	if !found {
		registry.Repos = append(registry.Repos, Repo{
			Path:     repoPath,
			LastSeen: now,
		})
	}

	return s.saveRegistry(registryPath, registry)
}

func (s *Store) loadRegistry(path string) (*RepoRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &RepoRegistry{}, err
	}

	var registry RepoRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	return &registry, nil
}

func (s *Store) saveRegistry(path string, registry *RepoRegistry) error {
	data, err := yaml.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
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

// RemoveSource removes a source from the registry.
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
