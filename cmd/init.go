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

// initCounts tracks how many sources were added per type.
type initCounts struct {
	git      int
	claude   int
	obsidian int
	markdown int
}

func (c initCounts) total() int {
	return c.git + c.claude + c.obsidian + c.markdown
}

func (c initCounts) summary() string {
	if c.total() == 0 {
		return ""
	}
	var parts []string
	if c.git > 0 {
		parts = append(parts, fmt.Sprintf("%d git %s", c.git, initPlural(c.git, "repo", "repos")))
	}
	if c.claude > 0 {
		parts = append(parts, fmt.Sprintf("%d claude %s", c.claude, initPlural(c.claude, "source", "sources")))
	}
	if c.obsidian > 0 {
		parts = append(parts, fmt.Sprintf("%d obsidian %s", c.obsidian, initPlural(c.obsidian, "vault", "vaults")))
	}
	if c.markdown > 0 {
		parts = append(parts, fmt.Sprintf("%d markdown %s", c.markdown, initPlural(c.markdown, "directory", "directories")))
	}
	return strings.Join(parts, ", ")
}

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
			fmt.Println(styleBold.Render("Welcome to anker. Let's find your sources."))
			fmt.Println("This wizard adds the data sources anker reads for your recap.")
			fmt.Println()
		}

		var counts initCounts
		var runErr error

		counts.git, runErr = initStepGit(store, registered)
		if runErr != nil {
			return runErr
		}

		counts.claude, runErr = initStepClaude(store, registered)
		if runErr != nil {
			return runErr
		}

		counts.obsidian, runErr = initStepObsidian(store, registered)
		if runErr != nil {
			return runErr
		}

		counts.markdown, runErr = initStepMarkdown(store, registered)
		if runErr != nil {
			return runErr
		}

		if !initYes {
			initStepEmail()
		}

		if _, cfgErr := config.EnsureConfigFile(); cfgErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: could not write config file: %v\n", cfgErr)
		}

		fmt.Println()

		if counts.total() > 0 {
			fmt.Println(styleBold.Render("Done. Added " + counts.summary() + "."))
			fmt.Println()
			fmt.Println("Try: anker recap thisweek")
		} else {
			fmt.Println(styleBold.Render("Done. Nothing new to add."))
			fmt.Println()
			fmt.Println("Run: anker recap thisweek")
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

	if initYes {
		return initStepGitAuto(store, registered, defaultCodeDir, home)
	}

	// Pre-check default path: if all repos there are already registered, show
	// a compact status line and skip the interactive section entirely.
	if initDirExists(defaultCodeDir) {
		allFound, _ := sources.DiscoverSources(defaultCodeDir, 2, nil)
		alreadyReg := 0
		newCount := 0
		for _, d := range allFound {
			if d.Type != "git" || initIsHomeDir(d.Path) {
				continue
			}
			if initIsRegistered(registered, "git", d.Path) {
				alreadyReg++
			} else {
				newCount++
			}
		}
		if newCount == 0 && alreadyReg > 0 {
			initCheckLine("Git repositories",
				fmt.Sprintf("%d %s in %s (already configured)",
					alreadyReg,
					initPlural(alreadyReg, "repo", "repos"),
					initShortenHome(defaultCodeDir, home),
				),
			)
			return 0, nil
		}
	}

	// Interactive section
	initSectionHeader("Git repositories")

	codeDir := initShortenHome(defaultCodeDir, home)
	if err := huh.NewInput().
		Title("Where do you keep your code?").
		Value(&codeDir).
		Run(); initIsAbort(err) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	if codeDir == "" {
		codeDir = initShortenHome(defaultCodeDir, home)
	}
	codeDir = initExpandHome(codeDir, home)

	if _, statErr := os.Stat(codeDir); statErr != nil {
		fmt.Println()
		fmt.Printf("Directory not found: %s\n", codeDir)
		fmt.Println(styleMuted.Render("Add a repository manually: anker source add git ~/path/to/repo"))
		fmt.Println()
		return 0, nil
	}

	fmt.Println()
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
			fmt.Println(styleMuted.Render(
				fmt.Sprintf("All %d %s in %s already registered.",
					alreadyReg, initPlural(alreadyReg, "repo", "repos"), displayDir)))
		} else {
			fmt.Printf("No git repositories found in %s.\n", displayDir)
			fmt.Println(styleMuted.Render("Add a repository manually: anker source add git ~/path/to/repo"))
		}
		fmt.Println()
		return 0, nil
	}

	fmt.Println()

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
		fmt.Println(styleSuccess.Render(
			fmt.Sprintf("✓ Added %d git %s", added, initPlural(added, "repository", "repositories"))))
		fmt.Println()
	}
	return added, nil
}

// initStepGitAuto handles git discovery for --yes mode without any prompts.
func initStepGitAuto(store *storage.Store, registered []sources.Config, codeDir, home string) (int, error) {
	if _, err := os.Stat(codeDir); err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Scanning %s ... directory not found.\n", initShortenHome(codeDir, home))
		return 0, nil
	}
	allDiscovered, err := sources.DiscoverSources(codeDir, 2, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "warning: could not scan %s: %v\n", codeDir, err)
		return 0, nil
	}
	var gitRepos []sources.DetectedSource
	for _, d := range allDiscovered {
		if d.Type == "git" && !initIsHomeDir(d.Path) && !initIsRegistered(registered, "git", d.Path) {
			gitRepos = append(gitRepos, d)
		}
	}
	_, _ = fmt.Fprintf(os.Stdout, "Found %d new git %s in %s.\n",
		len(gitRepos), initPlural(len(gitRepos), "repository", "repositories"), initShortenHome(codeDir, home))
	added := 0
	for _, r := range gitRepos {
		if err := initAddGitSource(store, r.Path); err != nil {
			return added, err
		}
		added++
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
		// Not installed -- skip silently.
		return 0, nil
	}

	claudePath := claudesource.DefaultClaudeHome()

	if initIsRegistered(registered, "claude", claudePath) {
		if !initYes {
			initCheckLine("Claude sessions", "~/.claude (already configured)")
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

	initSectionHeader("Claude sessions")

	var confirm bool
	if err := huh.NewConfirm().
		Title("Found Claude Code session log at ~/.claude/projects. Add?").
		Affirmative("Yes").
		Negative("No").
		Value(&confirm).
		Run(); initIsAbort(err) {
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
	fmt.Println(styleSuccess.Render("✓ Added Claude sessions"))
	fmt.Println()
	return 1, nil
}

func initStepObsidian(store *storage.Store, registered []sources.Config) (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, nil
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

	if initYes {
		added := 0
		for _, v := range detected {
			if initIsRegistered(registered, "obsidian", v) {
				continue
			}
			_, _ = fmt.Fprintf(os.Stdout, "Found Obsidian vault at %s.\n", initShortenHome(v, home))
			if err := initAddSource(store, "obsidian", v); err != nil {
				return added, err
			}
			added++
		}
		return added, nil
	}

	// Check if all detected vaults are already registered.
	if len(detected) > 0 {
		newCount := 0
		for _, v := range detected {
			if !initIsRegistered(registered, "obsidian", v) {
				newCount++
			}
		}
		if newCount == 0 {
			if len(detected) == 1 {
				initCheckLine("Obsidian vault", initShortenHome(detected[0], home)+" (already configured)")
			} else {
				initCheckLine("Obsidian vault",
					fmt.Sprintf("%d vaults (already configured)", len(detected)))
			}
			return 0, nil
		}
	}

	// Interactive section
	initSectionHeader("Obsidian vault")

	added := 0

	for _, v := range detected {
		if initIsRegistered(registered, "obsidian", v) {
			fmt.Println(styleMuted.Render(
				fmt.Sprintf("  %s already registered, skipping.", initShortenHome(v, home))))
			continue
		}
		var confirm bool
		if err := huh.NewConfirm().
			Title(fmt.Sprintf("Found Obsidian vault at %s. Add?", initShortenHome(v, home))).
			Affirmative("Yes").
			Negative("No").
			Value(&confirm).
			Run(); initIsAbort(err) {
			return added, nil
		} else if err != nil {
			return added, err
		}
		fmt.Println()
		if confirm {
			if err := initAddSource(store, "obsidian", v); err != nil {
				return added, err
			}
			fmt.Println(styleSuccess.Render("✓ Added Obsidian vault: " + initShortenHome(v, home)))
			fmt.Println()
			added++
		}
	}

	if len(detected) == 0 {
		fmt.Println("No Obsidian vaults found at common locations.")
		fmt.Println()
		n, addErr := initObsidianPromptManual(store, registered, home, "Enter path to a vault (or press Enter to skip)")
		added += n
		if addErr != nil {
			return added, addErr
		}
	}

	if len(detected) > 0 || added > 0 {
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
		fmt.Println()
	}

	return added, nil
}

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
		fmt.Println(styleMuted.Render("Already registered."))
		fmt.Println()
		return 0, nil
	}

	if err := initAddSource(store, "obsidian", vaultPath); err != nil {
		return 0, err
	}
	fmt.Println(styleSuccess.Render("✓ Added Obsidian vault: " + initShortenHome(vaultPath, home)))
	fmt.Println()
	return 1, nil
}

func initStepMarkdown(store *storage.Store, registered []sources.Config) (int, error) {
	if initYes {
		return 0, nil
	}

	// Check for already-registered markdown sources
	var regMarkdown []sources.Config
	for _, r := range registered {
		if r.Type == "markdown" {
			regMarkdown = append(regMarkdown, r)
		}
	}
	if len(regMarkdown) > 0 {
		home, _ := os.UserHomeDir()
		paths := make([]string, len(regMarkdown))
		for i, r := range regMarkdown {
			paths[i] = initShortenHome(r.Path, home)
		}
		initCheckLine("Markdown directories", strings.Join(paths, ", ")+" (already configured)")
		return 0, nil
	}

	home, _ := os.UserHomeDir()

	initSectionHeader("Markdown directories")

	var wantMarkdown bool
	if err := huh.NewConfirm().
		Title("Do you have standalone Markdown directories not inside Git or Obsidian?").
		Affirmative("Yes").
		Negative("No").
		Value(&wantMarkdown).
		Run(); initIsAbort(err) || !wantMarkdown {
		fmt.Println()
		return 0, nil
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
		fmt.Println(styleMuted.Render("Already registered."))
		fmt.Println()
		return 0, nil
	}

	if err := initAddSource(store, "markdown", mdPath); err != nil {
		return 0, err
	}
	fmt.Println(styleSuccess.Render("✓ Added Markdown directory: " + initShortenHome(mdPath, home)))
	fmt.Println()
	return 1, nil
}

func initStepEmail() {
	email, err := git.GetAuthorEmail()
	if err != nil || email == "" {
		return // no email configured, nothing to confirm
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
	fmt.Println(styleMuted.Render(
		fmt.Sprintf("Note: update your git config with: git config --global user.email %s", newEmail)))
	fmt.Println()
}

// initCheckLine prints a compact "✓ label   detail" line for already-configured steps.
func initCheckLine(label, detail string) {
	check := styleSuccess.Render("✓")
	lbl := styleBold.Render(fmt.Sprintf("%-22s", label))
	det := styleMuted.Render(detail)
	fmt.Printf("%s %s%s\n", check, lbl, det)
}

// initSectionHeader prints a styled section header for interactive steps.
func initSectionHeader(title string) {
	fmt.Println(styleHeader.Render(title))
	fmt.Println()
}

// initIsAbort reports whether err is a user abort (Ctrl+C).
func initIsAbort(err error) bool {
	return errors.Is(err, huh.ErrUserAborted)
}

// initPlural returns singular or plural based on n.
func initPlural(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
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
