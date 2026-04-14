# 0018: Smart Source Detection and CLI/TUI Modernization

**Date:** 2026-04-12
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Two related pain points in the user experience:

1. **Source setup was tedious.** Users had to manually identify each source type and add them one by one (`ikno source add git ~/code/repo1`, `ikno source add git ~/code/repo2`, ...). No way to scan a directory or auto-detect what type a path is.

2. **CLI output looked dated.** All user-facing output used raw `fmt.Println` with hand-drawn ASCII separators. Interactive prompts used `bufio.Scanner`. This felt inconsistent with modern CLI tools.

## Options Considered

### Source detection

**Manual-only (status quo):**
- Good: Explicit, no surprises
- Bad: Tedious for users with many repos
- Bad: Users must know source types upfront

**Auto-detection with confirmation:**
- Good: Detects type from directory signals (.git/, .obsidian/, .md files, .claude/)
- Good: `ikno source add ~/code` scans children and presents candidates
- Good: `ikno init` wizard discovers sources interactively
- Good: User confirms every addition (no implicit tracking)
- Bad: Detection heuristics can be wrong

### CLI/TUI framework

**Keep raw fmt.Println + bufio.Scanner:**
- Good: No dependencies
- Bad: No colors, no structure, no polish
- Bad: Interactive prompts are fragile (no validation, no cursor control)

**charmbracelet stack (huh, lipgloss, glamour):**
- Good: Modern TUI forms with validation (huh)
- Good: Consistent terminal styling (lipgloss)
- Good: Markdown rendering in terminal (glamour)
- Good: Well-maintained, widely adopted in Go CLI tools
- Bad: Adds dependencies

**bubbletea full TUI:**
- Good: Maximum control over terminal UI
- Bad: Overkill for a CLI tool (not a full-screen TUI app)
- Bad: Much more code to write and maintain

## Decision

We chose **auto-detection with confirmation** for source discovery and **charmbracelet stack** for CLI/TUI.

Why: Auto-detection removes tedium while preserving the explicit-registration principle (user confirms every source). The charmbracelet stack provides polished output and proper form handling without building a full TUI application.

### Source detection

- `DetectType(path)` checks signals with priority: git > obsidian > claude > markdown
- `DiscoverSources(path)` scans a directory's children and returns candidates
- `ikno init` wizard walks through source discovery step by step using huh forms
- `ikno source add <path>` infers type when not specified (see ADR 0002 amendment)

### TUI components

- `charmbracelet/huh` for all interactive prompts (init wizard, confirmations)
- `charmbracelet/lipgloss` for all styled terminal output (colors, borders, status lines)
- `charmbracelet/glamour` for rendering AI-generated markdown in the terminal
- No raw `fmt.Println` for user-facing output
