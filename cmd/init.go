package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charemma/anker/internal/config"
	"github.com/charemma/anker/internal/git"
	"github.com/charemma/anker/internal/sources"
	claudesource "github.com/charemma/anker/internal/sources/claude"
	"github.com/charemma/anker/internal/storage"
	"github.com/spf13/cobra"
)

const initListThreshold = 8

var initYes bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup wizard",
	Long: `Set up anker sources step by step.

Walks through each source type and offers to add what it finds:
  Git repositories in ~/code (or a path you specify)
  Claude Code session history in ~/.claude
  Obsidian vault at common locations
  Markdown directories (opt-in only)

Examples:
  anker init
  anker init --yes`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isTTY() && !initYes {
			return fmt.Errorf("interactive confirmation required, use --yes to skip")
		}

		store, err := storage.NewStore()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		registered, err := store.GetSources()
		if err != nil {
			return fmt.Errorf("failed to load sources: %w", err)
		}

		if !initYes {
			fmt.Println("Welcome to anker. Let's find your sources.")
			fmt.Println("This wizard adds the data sources anker reads for your recap.")
			fmt.Println()
		}

		added := 0

		n, err := initStepGit(store, registered)
		if err != nil {
			return err
		}
		added += n

		n, err = initStepClaude(store, registered)
		if err != nil {
			return err
		}
		added += n

		n, err = initStepObsidian(store, registered)
		if err != nil {
			return err
		}
		added += n

		n, err = initStepMarkdown(store, registered)
		if err != nil {
			return err
		}
		added += n

		initStepEmail()

		if _, cfgErr := config.EnsureConfigFile(); cfgErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: could not write config file: %v\n", cfgErr)
		}

		if !initYes {
			fmt.Println(strings.Repeat("-", 74))
			fmt.Println()
		}

		fmt.Printf("Done. %d source(s) added.\n", added)
		fmt.Println()
		if added > 0 {
			fmt.Println("Try: anker recap thisweek")
		} else {
			fmt.Println("Try: anker source add git . to track the current directory.")
		}
		return nil
	},
}

func initStepGit(store *storage.Store, registered []sources.Config) (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, nil
	}
	defaultCodeDir := filepath.Join(home, "code")

	if !initYes {
		fmt.Println("-- Git repositories " + strings.Repeat("-", 54))
		fmt.Println()
	}

	codeDir := defaultCodeDir

	if !initYes {
		_, _ = fmt.Fprintf(os.Stdout, "Where do you keep your code? [~/code]: ")
		input := strings.TrimSpace(readLine())
		if input != "" {
			codeDir = initExpandHome(input, home)
		}
	}

	if _, statErr := os.Stat(codeDir); statErr != nil {
		if initYes {
			_, _ = fmt.Fprintf(os.Stdout, "Scanning %s ... directory not found.\n", initShortenHome(codeDir, home))
		} else {
			fmt.Printf("Scanning %s ... directory not found.\n\n", initShortenHome(codeDir, home))
			fmt.Println("Add a repository manually later:")
			fmt.Println("  anker source add git ~/projects/my-repo")
			fmt.Println()
		}
		return 0, nil
	}

	// Scan all git repos, then split into new vs already registered.
	allDiscovered, discErr := sources.DiscoverSources(codeDir, 2, nil)
	if discErr != nil {
		if initYes {
			_, _ = fmt.Fprintf(os.Stderr, "warning: could not scan %s: %v\n", codeDir, discErr)
		} else {
			fmt.Printf("Scanning %s ... error: %v\n\n", initShortenHome(codeDir, home), discErr)
		}
		return 0, nil
	}

	var gitRepos []sources.DetectedSource
	alreadyReg := 0
	for _, d := range allDiscovered {
		if d.Type != "git" || initIsHomeDir(d.Path) {
			continue
		}
		if initIsRegistered(registered, "git", d.Path) {
			alreadyReg++
		} else {
			gitRepos = append(gitRepos, d)
		}
	}

	total := len(gitRepos)
	displayDir := initShortenHome(codeDir, home)

	if total == 0 && alreadyReg == 0 {
		if initYes {
			_, _ = fmt.Fprintf(os.Stdout, "Scanning %s ... no git repositories found.\n", displayDir)
		} else {
			fmt.Printf("Scanning %s ... no git repositories found.\n\n", displayDir)
			fmt.Println("Add a repository manually later:")
			fmt.Printf("  anker source add git %s/my-repo\n", displayDir)
			fmt.Println()
		}
		return 0, nil
	}

	if total == 0 {
		if !initYes {
			fmt.Printf("Scanning %s ... found %d git repositories (%d already registered).\n\n", displayDir, alreadyReg, alreadyReg)
		}
		return 0, nil
	}

	if initYes {
		_, _ = fmt.Fprintf(os.Stdout, "Scanning %s ... found %d git repositories.\n", displayDir, total+alreadyReg)
		added := 0
		for _, r := range gitRepos {
			if err := initAddGitSource(store, r.Path); err != nil {
				return added, err
			}
			added++
		}
		return added, nil
	}

	// Interactive: show summary line
	regSuffix := ""
	if alreadyReg > 0 {
		regSuffix = fmt.Sprintf(" (%d already registered)", alreadyReg)
	}

	if total <= initListThreshold {
		fmt.Printf("Scanning %s ... found %d git repositories%s:\n\n", displayDir, total+alreadyReg, regSuffix)
		for _, r := range gitRepos {
			fmt.Printf("  %s\n", r.Path)
		}
	} else {
		fmt.Printf("Scanning %s ... found %d git repositories%s.\n", displayDir, total+alreadyReg, regSuffix)
		fmt.Println()
		_, _ = fmt.Fprintf(os.Stdout, "Show list? [y/N]: ")
		if strings.ToLower(strings.TrimSpace(readLine())) == "y" {
			fmt.Println()
			for _, r := range gitRepos {
				fmt.Printf("  %s\n", r.Path)
			}
		}
	}

	fmt.Println()

	addLabel := fmt.Sprintf("Add all %d?", total)
	if alreadyReg > 0 {
		addLabel = fmt.Sprintf("Add %d new repositories?", total)
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s [Y/n]: ", addLabel)
	answer := strings.ToLower(strings.TrimSpace(readLine()))
	if answer != "" && answer != "y" && answer != "yes" {
		fmt.Println()
		return 0, nil
	}

	added := 0
	for _, r := range gitRepos {
		if err := initAddGitSource(store, r.Path); err != nil {
			return added, err
		}
		added++
	}
	fmt.Printf("Added %d git repositories.\n\n", added)
	return added, nil
}

func initStepClaude(store *storage.Store, registered []sources.Config) (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, nil
	}

	claudeProjects := filepath.Join(home, ".claude", "projects")
	if !initDirExists(claudeProjects) {
		if !initYes {
			fmt.Println("-- Claude sessions " + strings.Repeat("-", 55))
			fmt.Println()
			fmt.Println("No Claude Code sessions found at ~/.claude/projects. Skipping.")
			fmt.Println()
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "No Claude Code sessions found at ~/.claude/projects. Skipping.\n")
		}
		return 0, nil
	}

	claudePath := claudesource.DefaultClaudeHome()

	if initIsRegistered(registered, "claude", claudePath) {
		if !initYes {
			fmt.Println("-- Claude sessions " + strings.Repeat("-", 55))
			fmt.Println()
			fmt.Println("Already registered. Skipping.")
			fmt.Println()
		}
		return 0, nil
	}

	if initYes {
		_, _ = fmt.Fprintf(os.Stdout, "Found Claude Code sessions at ~/.claude/projects.\n")
		if err := initAddSource(store, "claude", claudePath); err != nil {
			return 0, err
		}
		return 1, nil
	}

	fmt.Println("-- Claude sessions " + strings.Repeat("-", 55))
	fmt.Println()
	fmt.Println("Found Claude Code session log at ~/.claude/projects.")
	_, _ = fmt.Fprintf(os.Stdout, "Add? [Y/n]: ")
	answer := strings.ToLower(strings.TrimSpace(readLine()))
	if answer != "" && answer != "y" && answer != "yes" {
		fmt.Println()
		return 0, nil
	}

	if err := initAddSource(store, "claude", claudePath); err != nil {
		return 0, err
	}
	fmt.Println("Added Claude sessions.")
	fmt.Println()
	return 1, nil
}

func initStepObsidian(store *storage.Store, registered []sources.Config) (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, nil
	}

	if !initYes {
		fmt.Println("-- Obsidian vault " + strings.Repeat("-", 56))
		fmt.Println()
	}

	// Auto-detect: check known candidate paths for .obsidian/
	candidates := []string{
		filepath.Join(home, "Documents", "Notes"),
		filepath.Join(home, "Obsidian"),
		filepath.Join(home, "obsidian"),
	}
	if env := os.Getenv("OBSIDIAN_VAULT"); env != "" {
		candidates = append([]string{env}, candidates...)
	}

	seen := make(map[string]bool)
	var detected []string
	for _, c := range candidates {
		abs, absErr := filepath.Abs(c)
		if absErr != nil || seen[abs] {
			continue
		}
		seen[abs] = true
		if initDirExists(filepath.Join(abs, ".obsidian")) {
			detected = append(detected, abs)
		}
	}

	added := 0

	// Offer auto-detected vaults
	for _, v := range detected {
		if initIsRegistered(registered, "obsidian", v) {
			if !initYes {
				fmt.Printf("Found Obsidian vault at %s -- already registered. Skipping.\n\n", initShortenHome(v, home))
			}
			continue
		}
		if initYes {
			_, _ = fmt.Fprintf(os.Stdout, "Found Obsidian vault at %s.\n", initShortenHome(v, home))
			if err := initAddSource(store, "obsidian", v); err != nil {
				return added, err
			}
			added++
			continue
		}
		fmt.Printf("Found Obsidian vault at %s.\n", initShortenHome(v, home))
		_, _ = fmt.Fprintf(os.Stdout, "Add? [Y/n]: ")
		answer := strings.ToLower(strings.TrimSpace(readLine()))
		if answer == "" || answer == "y" || answer == "yes" {
			if err := initAddSource(store, "obsidian", v); err != nil {
				return added, err
			}
			fmt.Println("Added Obsidian vault.")
			added++
		}
		fmt.Println()
	}

	if !initYes {
		nothingFound := len(detected) == 0

		if nothingFound {
			// Nothing auto-detected: offer one manual entry first.
			fmt.Println("No Obsidian vaults found.")
			_, _ = fmt.Fprintf(os.Stdout, "Enter path to a vault (or press Enter to skip): ")
			n, addErr := initObsidianAddManual(store, registered, home)
			added += n
			if addErr != nil {
				return added, addErr
			}
		}

		// Loop "Add another vault?" when there is already context (auto-detected
		// or manually added at least one). Skip if nothing was found and user
		// pressed Enter on the first prompt.
		if !nothingFound || added > 0 {
			for {
				_, _ = fmt.Fprintf(os.Stdout, "Add another vault? [y/N]: ")
				answer := strings.ToLower(strings.TrimSpace(readLine()))
				if answer != "y" && answer != "yes" {
					break
				}
				_, _ = fmt.Fprintf(os.Stdout, "Path: ")
				n, addErr := initObsidianAddManual(store, registered, home)
				added += n
				if addErr != nil {
					return added, addErr
				}
			}
		}
		fmt.Println()
	}

	return added, nil
}

// initObsidianAddManual reads a vault path from stdin and adds it if valid.
// The caller is responsible for printing the prompt before calling this.
func initObsidianAddManual(store *storage.Store, registered []sources.Config, home string) (int, error) {
	input := strings.TrimSpace(readLine())
	if input == "" {
		return 0, nil
	}
	vaultPath := initExpandHome(input, home)
	switch {
	case initIsHomeDir(vaultPath):
		fmt.Println("Home directory cannot be added as a source.")
		return 0, nil
	case !initDirExists(filepath.Join(vaultPath, ".obsidian")):
		fmt.Printf("No .obsidian/ found at %s.\n", vaultPath)
		return 0, nil
	case initIsRegistered(registered, "obsidian", vaultPath):
		fmt.Println("Already registered.")
		return 0, nil
	}
	if err := initAddSource(store, "obsidian", vaultPath); err != nil {
		return 0, err
	}
	fmt.Println("Added Obsidian vault.")
	return 1, nil
}

func initStepMarkdown(store *storage.Store, registered []sources.Config) (int, error) {
	if initYes {
		return 0, nil
	}

	home, _ := os.UserHomeDir()

	fmt.Println("-- Markdown directories " + strings.Repeat("-", 50))
	fmt.Println()
	_, _ = fmt.Fprintf(os.Stdout, "Do you have standalone Markdown directories not inside Git or Obsidian? [y/N]: ")
	answer := strings.ToLower(strings.TrimSpace(readLine()))
	if answer != "y" && answer != "yes" {
		fmt.Println()
		return 0, nil
	}

	_, _ = fmt.Fprintf(os.Stdout, "Path to Markdown directory: ")
	mdPath := strings.TrimSpace(readLine())
	fmt.Println()
	if mdPath == "" {
		return 0, nil
	}

	mdPath = initExpandHome(mdPath, home)

	if initIsHomeDir(mdPath) {
		fmt.Println("Home directory cannot be added as a source. Skipping.")
		fmt.Println()
		return 0, nil
	}

	if _, err := os.Stat(mdPath); err != nil {
		fmt.Printf("Path %s not found. Skipping.\n\n", mdPath)
		return 0, nil
	}

	if !initHasMDFiles(mdPath) {
		fmt.Printf("No .md files found in %s. Skipping.\n\n", mdPath)
		return 0, nil
	}

	if initIsRegistered(registered, "markdown", mdPath) {
		fmt.Println("Already registered. Skipping.")
		fmt.Println()
		return 0, nil
	}

	if err := initAddSource(store, "markdown", mdPath); err != nil {
		return 0, err
	}
	fmt.Printf("Added Markdown directory: %s.\n\n", mdPath)
	return 1, nil
}

func initStepEmail() {
	email, err := git.GetAuthorEmail()

	if initYes {
		if err == nil && email != "" {
			_, _ = fmt.Fprintf(os.Stdout, "Using %s from git config.\n", email)
		}
		return
	}

	fmt.Println("-- Git author email " + strings.Repeat("-", 54))
	fmt.Println()

	if err != nil || email == "" {
		fmt.Println("No git author email configured. Set one with:")
		fmt.Println("  git config --global user.email you@example.com")
		fmt.Println()
		return
	}

	fmt.Printf("Using %s from your git config.\n", email)
	_, _ = fmt.Fprintf(os.Stdout, "Change? [y/N]: ")
	answer := strings.ToLower(strings.TrimSpace(readLine()))
	if answer == "y" || answer == "yes" {
		_, _ = fmt.Fprintf(os.Stdout, "New email: ")
		newEmail := strings.TrimSpace(readLine())
		if newEmail != "" {
			fmt.Printf("Note: update your git config with: git config --global user.email %s\n", newEmail)
		}
	}
	fmt.Println()
}

// initAddGitSource adds a git source with author email from git config.
func initAddGitSource(store *storage.Store, path string) error {
	cfg := sources.Config{
		Type:     "git",
		Path:     path,
		Metadata: make(map[string]string),
	}
	if email, err := git.GetAuthorEmail(); err == nil && email != "" {
		cfg.Metadata["author"] = email
	}
	return store.AddSource(cfg)
}

// initAddSource adds a source to the store without extra metadata.
func initAddSource(store *storage.Store, sourceType, path string) error {
	return store.AddSource(sources.Config{
		Type:     sourceType,
		Path:     path,
		Metadata: make(map[string]string),
	})
}

// initIsRegistered reports whether a source with the given type and path is already registered.
func initIsRegistered(registered []sources.Config, sourceType, path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, r := range registered {
		if r.Type != sourceType {
			continue
		}
		regAbs, err := filepath.Abs(r.Path)
		if err != nil {
			continue
		}
		if regAbs == absPath {
			return true
		}
	}
	return false
}

// initIsHomeDir reports whether path is the user's home directory.
func initIsHomeDir(path string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return abs == home
}

// initDirExists reports whether path is an existing directory.
func initDirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// initHasMDFiles reports whether path contains at least one .md file as a direct child.
func initHasMDFiles(path string) bool {
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

// initExpandHome expands a leading ~/ to the user home directory.
func initExpandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

// initShortenHome replaces the home directory prefix with ~.
func initShortenHome(path, home string) string {
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + path[len(home):]
	}
	return path
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "Skip interactive confirmation, add all discovered sources")
}
