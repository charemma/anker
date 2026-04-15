package sources

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectedSource is a candidate source found by auto-detection.
type DetectedSource struct {
	Path   string
	Type   string
	Reason string // human-readable: "found .git/"
}

// DetectType inspects path and returns matching source types using priority rules:
//   - .git/ present: only git (plus claude if applicable), obsidian and markdown skipped
//   - .obsidian/ present: only obsidian (plus claude if applicable), markdown skipped
//   - claude (.claude/projects/ child or path under ~/.claude): only claude, markdown skipped
//   - markdown: only when none of the above match
//
// Returns an empty slice (no error) when no type can be inferred.
func DetectType(path string) ([]DetectedSource, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	var results []DetectedSource

	hasDotGit := isDir(filepath.Join(abs, ".git"))
	hasBareGit := !hasDotGit && isBareGitRepo(abs)
	hasGit := hasDotGit || hasBareGit
	hasObsidian := isDir(filepath.Join(abs, ".obsidian"))
	hasClaude := isClaudePath(abs)

	// git takes highest priority; obsidian is next. They are mutually exclusive.
	if hasDotGit {
		results = append(results, DetectedSource{Path: abs, Type: "git", Reason: "found .git/"})
	} else if hasBareGit {
		results = append(results, DetectedSource{Path: abs, Type: "git", Reason: "found bare git repository"})
	} else if hasObsidian {
		results = append(results, DetectedSource{Path: abs, Type: "obsidian", Reason: "found .obsidian/"})
	}

	// claude is independent of git/obsidian but blocks markdown.
	if hasClaude {
		results = append(results, DetectedSource{Path: abs, Type: "claude", Reason: "found .claude/projects/"})
	}

	// markdown only when no git, obsidian, or claude source was found.
	if !hasGit && !hasObsidian && !hasClaude && hasMDFiles(abs) {
		results = append(results, DetectedSource{Path: abs, Type: "markdown", Reason: "found .md files"})
	}

	return results, nil
}

// DiscoverSources scans dir up to the given depth and returns all detected sources
// that are not already in registered. depth=1 scans direct children only.
// Permission errors on child directories are silently skipped.
// Only an error reading dir itself is returned as a fatal error.
func DiscoverSources(dir string, depth int, registered []Config) ([]DetectedSource, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	registeredSet := buildRegisteredSet(registered)
	return discoverRecursive(absDir, depth, registeredSet)
}

func discoverRecursive(dir string, depth int, registered map[string]bool) ([]DetectedSource, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var results []DetectedSource
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip symlinks
		info, err := os.Lstat(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}

		abs := filepath.Join(dir, entry.Name())

		// Skip if already registered
		if registered[abs] {
			continue
		}

		detected, err := DetectType(abs)
		if err != nil {
			continue // silently skip unreadable children
		}
		results = append(results, detected...)

		if depth > 1 {
			sub, err := discoverRecursive(abs, depth-1, registered)
			if err != nil {
				continue // silently skip permission errors
			}
			results = append(results, sub...)
		}
	}

	return results, nil
}

// isDir returns true if path is an existing directory (follows symlinks).
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// isBareGitRepo returns true if path looks like a bare git repository.
// A bare repo contains HEAD as a regular file and objects/ as a directory.
func isBareGitRepo(path string) bool {
	headInfo, err := os.Stat(filepath.Join(path, "HEAD"))
	if err != nil || headInfo.IsDir() {
		return false
	}
	return isDir(filepath.Join(path, "objects"))
}

// hasMDFiles returns true if path contains at least one .md file as a direct child.
func hasMDFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			return true
		}
	}
	return false
}

// isClaudePath returns true if path is ~/.claude itself or any path under it.
func isClaudePath(abs string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	claudeHome := filepath.Join(home, ".claude")
	return abs == claudeHome || strings.HasPrefix(abs, claudeHome+string(filepath.Separator))
}

// buildRegisteredSet builds an abs-path lookup map from the registered configs.
func buildRegisteredSet(registered []Config) map[string]bool {
	set := make(map[string]bool, len(registered))
	for _, cfg := range registered {
		abs, err := filepath.Abs(cfg.Path)
		if err != nil {
			continue
		}
		set[abs] = true
	}
	return set
}
