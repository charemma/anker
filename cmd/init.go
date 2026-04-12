package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charemma/anker/internal/config"
	"github.com/charemma/anker/internal/git"
	"github.com/charemma/anker/internal/sources"
	claudesource "github.com/charemma/anker/internal/sources/claude"
	"github.com/charemma/anker/internal/storage"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

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

	codeDir := initShortenHome(defaultCodeDir, home)

	if !initYes {
		if err := huh.NewInput().
			Title("Where do you keep your code?").
			Value(&codeDir).
			Run(); initIsAbort(err) {
			return 0, nil
		} else if err != nil {
			return 0, err
		}
		if codeDir == "" {
			codeDir = defaultCodeDir
		}
	}

	codeDir = initExpandHome(codeDir, home)

	if _, statErr := os.Stat(codeDir); statErr != nil {
		if initYes {
			_, _ = fmt.Fprintf(os.Stdout, "Scanning %s ... directory not found.\n", initShortenHome(codeDir, home))
		} else {
			fmt.Printf("Directory not found: %s\n\n", codeDir)
			fmt.Println("Add a repository manually later:")
			fmt.Println("  anker source add git ~/projects/my-repo")
			fmt.Println()
		}
		return 0, nil
	}

	fmt.Printf("Scanning %s ...\n", initShortenHome(codeDir, home))

	allDiscovered, discErr := sources.DiscoverSources(codeDir, 2, nil)
	if discErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "warning: could not scan %s: %v\n", codeDir, discErr)
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

	if total == 0 {
		if alreadyReg > 0 {
			fmt.Printf("All %d git repositories in %s already registered.\n\n", alreadyReg, displayDir)
		} else {
			if initYes {
				_, _ = fmt.Fprintf(os.Stdout, "No git repositories found in %s.\n", displayDir)
			} else {
				fmt.Printf("No git repositories found in %s.\n\n", displayDir)
				fmt.Println("Add a repository manually later:")
				fmt.Printf("  anker source add git %s/my-repo\n", displayDir)
				fmt.Println()
			}
		}
		return 0, nil
	}

	if initYes {
		_, _ = fmt.Fprintf(os.Stdout, "Found %d git repositories.\n", total)
		added := 0
		for _, r := range gitRepos {
			if err := initAddGitSource(store, r.Path); err != nil {
				return added, err
			}
			added++
		}
		return added, nil
	}

	// Build MultiSelect options -- all selected by default.
	options := make([]huh.Option[string], total)
	for i, r := range gitRepos {
		options[i] = huh.NewOption(initShortenHome(r.Path, home), r.Path).Selected(true)
	}

	title := fmt.Sprintf("Select repositories to add (%d found", total)
	if alreadyReg > 0 {
		title += fmt.Sprintf(", %d already registered", alreadyReg)
	}
	title += ")"

	var selected []string
	height := min(total+3, 16)
	if err := huh.NewMultiSelect[string]().
		Title(title).
		Options(options...).
		Height(height).
		Value(&selected).
		Run(); initIsAbort(err) {
		fmt.Println()
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	fmt.Println()
	added := 0
	for _, path := range selected {
		if err := initAddGitSource(store, path); err != nil {
			return added, err
		}
		added++
	}
	if added > 0 {
		fmt.Printf("Added %d git repositories.\n\n", added)
	}
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

	var confirm bool
	if err := huh.NewConfirm().
		Title("Found Claude Code session log at ~/.claude/projects. Add?").
		Affirmative("Yes").
		Negative("No").
		Value(&confirm).
		Run(); initIsAbort(err) {
		fmt.Println()
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	fmt.Println()
	if !confirm {
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

		var confirm bool
		if err := huh.NewConfirm().
			Title(fmt.Sprintf("Found Obsidian vault at %s. Add?", initShortenHome(v, home))).
			Affirmative("Yes").
			Negative("No").
			Value(&confirm).
			Run(); initIsAbort(err) {
			fmt.Println()
			return added, nil
		} else if err != nil {
			return added, err
		}

		fmt.Println()
		if confirm {
			if err := initAddSource(store, "obsidian", v); err != nil {
				return added, err
			}
			fmt.Println("Added Obsidian vault.")
			fmt.Println()
			added++
		}
	}

	if !initYes {
		nothingFound := len(detected) == 0
		if nothingFound {
			fmt.Println("No Obsidian vaults found.")
			fmt.Println()
		}

		// Manual entry: one attempt when nothing found, then loop.
		if nothingFound {
			n, addErr := initObsidianPromptManual(store, registered, home, "Enter path to a vault (or press Enter to skip)")
			added += n
			if addErr != nil {
				return added, addErr
			}
		}

		// "Add another vault?" loop -- only when context exists.
		if !nothingFound || added > 0 {
			for {
				var addAnother bool
				if err := huh.NewConfirm().
					Title("Add another vault?").
					Affirmative("Yes").
					Negative("No").
					Value(&addAnother).
					Run(); initIsAbort(err) || !addAnother {
					break
				} else if err != nil {
					return added, err
				}

				n, addErr := initObsidianPromptManual(store, registered, home, "Path to Obsidian vault")
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

// initObsidianPromptManual shows an Input for a vault path and adds it if valid.
func initObsidianPromptManual(store *storage.Store, registered []sources.Config, home, title string) (int, error) {
	var vaultPath string
	if err := huh.NewInput().
		Title(title).
		Value(&vaultPath).
		Validate(func(s string) error {
			if s == "" {
				return nil
			}
			expanded := initExpandHome(s, home)
			if initIsHomeDir(expanded) {
				return fmt.Errorf("home directory cannot be added as a source")
			}
			if !initDirExists(filepath.Join(expanded, ".obsidian")) {
				return fmt.Errorf("no .obsidian/ found at %s", s)
			}
			return nil
		}).
		Run(); initIsAbort(err) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	fmt.Println()

	if vaultPath == "" {
		return 0, nil
	}

	vaultPath = initExpandHome(vaultPath, home)

	if initIsRegistered(registered, "obsidian", vaultPath) {
		fmt.Println("Already registered.")
		fmt.Println()
		return 0, nil
	}

	if err := initAddSource(store, "obsidian", vaultPath); err != nil {
		return 0, err
	}
	fmt.Println("Added Obsidian vault.")
	fmt.Println()
	return 1, nil
}

func initStepMarkdown(store *storage.Store, registered []sources.Config) (int, error) {
	if initYes {
		return 0, nil
	}

	home, _ := os.UserHomeDir()

	fmt.Println("-- Markdown directories " + strings.Repeat("-", 50))
	fmt.Println()

	var wantMarkdown bool
	if err := huh.NewConfirm().
		Title("Do you have standalone Markdown directories not inside Git or Obsidian?").
		Affirmative("Yes").
		Negative("No").
		Value(&wantMarkdown).
		Run(); initIsAbort(err) || !wantMarkdown {
		fmt.Println()
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	fmt.Println()

	var mdPath string
	if err := huh.NewInput().
		Title("Path to Markdown directory").
		Value(&mdPath).
		Validate(func(s string) error {
			if s == "" {
				return nil
			}
			expanded := initExpandHome(s, home)
			if initIsHomeDir(expanded) {
				return fmt.Errorf("home directory cannot be added as a source")
			}
			if _, statErr := os.Stat(expanded); statErr != nil {
				return fmt.Errorf("directory not found: %s", s)
			}
			if !initHasMDFiles(expanded) {
				return fmt.Errorf("no .md files found in %s", s)
			}
			return nil
		}).
		Run(); initIsAbort(err) {
		fmt.Println()
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	fmt.Println()

	if mdPath == "" {
		return 0, nil
	}

	mdPath = initExpandHome(mdPath, home)

	if initIsRegistered(registered, "markdown", mdPath) {
		fmt.Println("Already registered.")
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

	var change bool
	if runErr := huh.NewConfirm().
		Title(fmt.Sprintf("Using %s from your git config. Change?", email)).
		Affirmative("Yes").
		Negative("No").
		Value(&change).
		Run(); initIsAbort(runErr) || !change {
		fmt.Println()
		return
	}

	fmt.Println()

	var newEmail string
	if runErr := huh.NewInput().
		Title("New git author email").
		Value(&newEmail).
		Run(); initIsAbort(runErr) || newEmail == "" {
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Printf("Note: update your git config with: git config --global user.email %s\n\n", newEmail)
}

// initIsAbort reports whether err is a user abort (Ctrl+C).
func initIsAbort(err error) bool {
	return errors.Is(err, huh.ErrUserAborted)
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
