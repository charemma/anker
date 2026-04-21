package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charemma/ikno/internal/config"
	"github.com/charemma/ikno/internal/git"
	"github.com/charemma/ikno/internal/sources"
	claudesource "github.com/charemma/ikno/internal/sources/claude"
	"github.com/charemma/ikno/internal/storage"
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

// scanResults holds the results of the unified home directory scan.
type scanResults struct {
	gitRepos       []string
	obsidianVaults []string
}

var (
	initYes       bool
	initScanDepth int
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup wizard",
	Long: `Set up ikno sources step by step.

Scans your home directory for git repositories and Obsidian vaults,
then walks through each source type and offers to add what it finds:
  Git repositories (discovered via .git directories)
  Claude Code session history in ~/.claude
  Obsidian vaults (discovered via .obsidian directories)
  Markdown directories (opt-in only)

Examples:
  ikno init
  ikno init --yes
  ikno init --scan-depth 3`,
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

		cfg, cfgErr := config.Load()
		if cfgErr != nil {
			cfg = config.DefaultConfig()
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine home directory: %w", err)
		}

		if !initYes {
			fmt.Println(styleBold.Render("Welcome to ikno. Let's find your sources."))
			fmt.Println("This wizard adds the data sources ikno reads for your recap.")
			fmt.Println()
		}

		// Unified scan of $HOME for .git and .obsidian directories.
		showProgress := isTTY() && !initYes
		scan := initScanHome(home, initScanDepth, showProgress)

		_, _ = fmt.Fprintf(os.Stdout, "Found %d git %s and %d Obsidian %s.\n",
			len(scan.gitRepos), initPlural(len(scan.gitRepos), "repo", "repos"),
			len(scan.obsidianVaults), initPlural(len(scan.obsidianVaults), "vault", "vaults"))
		fmt.Println()

		var counts initCounts
		var runErr error

		counts.git, runErr = initStepGit(store, registered, scan.gitRepos, home)
		if runErr != nil {
			return runErr
		}

		counts.claude, runErr = initStepClaude(store, registered)
		if runErr != nil {
			return runErr
		}

		counts.obsidian, runErr = initStepObsidian(store, registered, scan.obsidianVaults, home)
		if runErr != nil {
			return runErr
		}

		counts.markdown, runErr = initStepMarkdown(store, registered)
		if runErr != nil {
			return runErr
		}

		configChanged := false
		if !initYes {
			emailChanged, runErr := initStepEmail(cfg)
			if runErr != nil {
				return runErr
			}
			if emailChanged {
				configChanged = true
			}

			langChanged, runErr := initStepLanguage(cfg)
			if runErr != nil {
				return runErr
			}
			if langChanged {
				configChanged = true
			}
		}

		if configChanged {
			if err := config.Save(cfg); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warning: could not save config: %v\n", err)
			}
		} else {
			if _, cfgErr := config.EnsureConfigFile(); cfgErr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warning: could not write config file: %v\n", cfgErr)
			}
		}

		fmt.Println()

		if counts.total() > 0 {
			fmt.Println(styleBold.Render("Done. Added " + counts.summary() + "."))
			fmt.Println()
			fmt.Println("Try: ikno recap thisweek")
		} else {
			fmt.Println(styleBold.Render("Done. Nothing new to add."))
			fmt.Println()
			fmt.Println("Run: ikno recap thisweek")
		}
		return nil
	},
}

// initScanHome walks home up to maxDepth levels and collects paths containing
// .git or .obsidian directories. The walk runs in a goroutine and sends results
// through a channel. When showProgress is true, a live counter is displayed on
// stderr that updates as results come in.
func initScanHome(home string, maxDepth int, showProgress bool) scanResults {
	skipDirs := map[string]bool{
		"node_modules": true, "Library": true, ".Trash": true,
		".cache": true, ".local": true, ".npm": true, ".cargo": true,
		".nix-defexpr": true, ".nix-profile": true,
		"vendor": true, "dist": true, "build": true,
	}

	type result struct {
		path     string
		category string // "git" or "obsidian"
	}

	ch := make(chan result, 64)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		homeDepth := strings.Count(filepath.Clean(home), string(filepath.Separator))

		_ = filepath.WalkDir(home, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			if !d.IsDir() {
				return nil
			}

			name := d.Name()
			depth := strings.Count(filepath.Clean(path), string(filepath.Separator)) - homeDepth

			// Enforce max depth.
			if depth > maxDepth {
				return filepath.SkipDir
			}

			// Skip the home directory itself -- we only care about children.
			if depth == 0 {
				return nil
			}

			// Skip hidden directories. These are generally noise (.cache,
			// .local, etc.) and skipping them is a big performance win.
			if strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}

			// Skip known heavy directories.
			if skipDirs[name] {
				return filepath.SkipDir
			}

			// Check for .git directory (don't recurse into the repo).
			hasGit := initDirExists(filepath.Join(path, ".git"))
			hasObsidian := initDirExists(filepath.Join(path, ".obsidian"))

			if hasGit {
				abs, absErr := filepath.Abs(path)
				if absErr == nil {
					ch <- result{path: abs, category: "git"}
				}
				// Don't recurse into git repos -- they won't contain
				// independent .obsidian vaults at a useful level.
				return filepath.SkipDir
			}

			if hasObsidian {
				abs, absErr := filepath.Abs(path)
				if absErr == nil {
					ch <- result{path: abs, category: "obsidian"}
				}
				// Obsidian vaults don't nest, skip subtree.
				return filepath.SkipDir
			}

			return nil
		})
	}()

	// Close channel when walker is done.
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Collect results from the channel and optionally show live progress.
	var gitRepos, obsidianVaults []string
	for r := range ch {
		switch r.category {
		case "git":
			gitRepos = append(gitRepos, r.path)
		case "obsidian":
			obsidianVaults = append(obsidianVaults, r.path)
		}
		if showProgress {
			_, _ = fmt.Fprintf(os.Stderr, "\rScanning... found %d git %s, %d %s",
				len(gitRepos), initPlural(len(gitRepos), "repo", "repos"),
				len(obsidianVaults), initPlural(len(obsidianVaults), "vault", "vaults"))
		}
	}

	// Clear the progress line.
	if showProgress {
		_, _ = fmt.Fprintf(os.Stderr, "\r%80s\r", "")
	}

	// Add OBSIDIAN_VAULT env var if set and valid.
	if env := os.Getenv("OBSIDIAN_VAULT"); env != "" {
		abs, absErr := filepath.Abs(env)
		if absErr == nil && initDirExists(filepath.Join(abs, ".obsidian")) {
			obsidianVaults = append(obsidianVaults, abs)
		}
	}

	return scanResults{
		gitRepos:       initDedup(gitRepos),
		obsidianVaults: initDedup(obsidianVaults),
	}
}

func initStepGit(store *storage.Store, registered []sources.Config, scannedRepos []string, home string) (int, error) {
	if initYes {
		return initStepGitAuto(store, registered, scannedRepos, home)
	}

	// Separate new repos from already-registered ones.
	var newRepos []string
	alreadyReg := 0
	for _, path := range scannedRepos {
		if initIsHomeDir(path) {
			continue
		}
		if initIsRegistered(registered, "git", path) {
			alreadyReg++
		} else {
			newRepos = append(newRepos, path)
		}
	}

	if len(newRepos) == 0 {
		if alreadyReg > 0 {
			initCheckLine("Git repositories",
				fmt.Sprintf("%d %s (already configured)",
					alreadyReg, initPlural(alreadyReg, "repo", "repos")))
		} else {
			initCheckLine("Git repositories", "none found")
		}
		return 0, nil
	}

	// Interactive section -- show multi-select with scan results.
	initSectionHeader("Git repositories")

	options := make([]huh.Option[string], len(newRepos))
	for i, r := range newRepos {
		options[i] = huh.NewOption(initShortenHome(r, home), r)
	}

	title := fmt.Sprintf("Select repositories to add with Space (%d found", len(newRepos))
	if alreadyReg > 0 {
		title += fmt.Sprintf(", %d already registered", alreadyReg)
	}
	title += ")"

	var selected []string
	height := min(len(newRepos)+3, 16)
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
			fmt.Sprintf("Added %d git %s", added, initPlural(added, "repository", "repositories"))))
		fmt.Println()
	}
	return added, nil
}

// initStepGitAuto handles git discovery for --yes mode without any prompts.
func initStepGitAuto(store *storage.Store, registered []sources.Config, scannedRepos []string, home string) (int, error) {
	var newRepos []string
	for _, path := range scannedRepos {
		if !initIsHomeDir(path) && !initIsRegistered(registered, "git", path) {
			newRepos = append(newRepos, path)
		}
	}
	_, _ = fmt.Fprintf(os.Stdout, "Found %d new git %s.\n",
		len(newRepos), initPlural(len(newRepos), "repository", "repositories"))
	added := 0
	for _, path := range newRepos {
		if err := initAddGitSource(store, path); err != nil {
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
	fmt.Println(styleSuccess.Render("Added Claude sessions"))
	fmt.Println()
	return 1, nil
}

func initStepObsidian(store *storage.Store, registered []sources.Config, scannedVaults []string, home string) (int, error) {
	if initYes {
		added := 0
		for _, v := range scannedVaults {
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
	if len(scannedVaults) > 0 {
		newCount := 0
		for _, v := range scannedVaults {
			if !initIsRegistered(registered, "obsidian", v) {
				newCount++
			}
		}
		if newCount == 0 {
			if len(scannedVaults) == 1 {
				initCheckLine("Obsidian vault", initShortenHome(scannedVaults[0], home)+" (already configured)")
			} else {
				initCheckLine("Obsidian vault",
					fmt.Sprintf("%d vaults (already configured)", len(scannedVaults)))
			}
			return 0, nil
		}
	}

	// Interactive section
	initSectionHeader("Obsidian vault")

	added := 0

	for _, v := range scannedVaults {
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
			fmt.Println(styleSuccess.Render("Added Obsidian vault: " + initShortenHome(v, home)))
			fmt.Println()
			added++
		}
	}

	if len(scannedVaults) == 0 {
		fmt.Println("No Obsidian vaults found.")
		fmt.Println()
		n, addErr := initObsidianPromptManual(store, registered, home, "Enter path to a vault (or press Enter to skip)")
		added += n
		if addErr != nil {
			return added, addErr
		}
	}

	if len(scannedVaults) > 0 || added > 0 {
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
	fmt.Println(styleSuccess.Render("Added Obsidian vault: " + initShortenHome(vaultPath, home)))
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
	fmt.Println(styleSuccess.Render("Added Markdown directory: " + initShortenHome(mdPath, home)))
	fmt.Println()
	return 1, nil
}

// initStepEmail handles the email configuration step.
// Returns true if the email was updated and config should be saved.
func initStepEmail(cfg *config.Config) (bool, error) {
	// Already persisted in config.yaml: show compact line, nothing to do.
	if cfg.AuthorEmail != "" {
		initCheckLine("Git author", cfg.AuthorEmail+" (already configured)")
		return false, nil
	}

	// Found in git config: just show it, no question needed.
	gitEmail, _ := git.GetAuthorEmail()
	if gitEmail != "" {
		initCheckLine("Git author", gitEmail)
		return false, nil
	}

	// Nothing known: ask.
	initSectionHeader("Git author email")

	var email string
	if err := huh.NewInput().
		Title("Git author email for filtering commits").
		Value(&email).
		Run(); initIsAbort(err) {
		fmt.Println()
		return false, nil
	} else if err != nil {
		return false, err
	}
	fmt.Println()

	if email == "" {
		return false, nil
	}

	cfg.AuthorEmail = email
	fmt.Println(styleSuccess.Render("Git author: " + email))
	fmt.Println()
	return true, nil
}

// initStepLanguage handles the recap language configuration step.
// Returns true if the language was updated and config should be saved.
func initStepLanguage(cfg *config.Config) (bool, error) {
	if cfg.AILanguage != "" {
		initCheckLine("Recap language", cfg.AILanguage+" (already configured)")
		return false, nil
	}

	initSectionHeader("Recap language")

	var lang string
	if err := huh.NewInput().
		Title("Language for AI recaps (e.g. english, deutsch, español)").
		Value(&lang).
		Run(); initIsAbort(err) {
		fmt.Println()
		return false, nil
	} else if err != nil {
		return false, err
	}
	fmt.Println()

	if lang == "" {
		return false, nil
	}

	cfg.AILanguage = lang
	fmt.Println(styleSuccess.Render("Recap language: " + lang))
	fmt.Println()
	return true, nil
}

// initCheckLine prints a compact "label   detail" line for already-configured steps.
func initCheckLine(label, detail string) {
	check := styleSuccess.Render("*")
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

// initDedup deduplicates paths using filepath.EvalSymlinks to handle
// case-insensitive filesystems (macOS) and symlinks.
func initDedup(paths []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, p := range paths {
		key := p
		if resolved, err := filepath.EvalSymlinks(p); err == nil {
			key = resolved
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, p)
	}
	return result
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "Skip interactive confirmation, add all discovered sources")
	initCmd.Flags().IntVar(&initScanDepth, "scan-depth", 4, "Maximum directory depth to scan for sources")
}
