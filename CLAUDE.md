# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**anker** is a Go-based CLI tool for reconstructing your workday after the fact.

## Project goals

- Build a calm, explicit, text-first CLI
- No background tracking or time measurement
- Focus on reconstructing work *after the fact*
- Prioritize clarity, predictability, and good UX

## Design principles

- Local-first: all data stays on your machine
- Explicit commands over implicit behavior
- CLI arguments override configuration
- No automatic scanning of the filesystem
- No plugins or extension systems (yet)
- Deferred analysis: work first, summarize later

## Initial scope

Focus on implementing:

1. `anker source`
   - Add/list/remove data sources (git, markdown, obsidian, etc.)
   - Extensible source system for multiple data types
   - Positional arguments: `anker source add git .`

2. `anker report`
   - Analyze all tracked sources for a time period
   - Flexible time specs (today, thisweek, date ranges)
   - Generate human-readable summary

3. `anker note`
   - Store one-off work notes
   - (not yet implemented)

## Technical constraints

- Language: Go
- CLI parsing: Cobra (allows easy addition of subcommands)
- Storage: YAML / Markdown (human-readable)
- Build automation: Just (wrapper) + Dagger (build logic)
- No database, no network dependencies

## Development commands

Build and run:
- `just build` - build binary to bin/anker
- `just build-to <path>` - build binary to custom location
- `go run . source list` - quick run without building

Testing:
- `just test` - run all tests
- `just coverage` - run tests with coverage report
- `go test ./internal/...` - test specific package

Code quality:
- `just lint` - run linter
- `just check` - run all checks (test + lint + build)
- `just clean` - remove build artifacts

## Code architecture

```
main.go                  - entry point, calls cmd.Execute()
cmd/
  root.go                - Cobra root command setup
  source.go              - manage data sources (add/list/remove)
  report.go              - generate work summaries
internal/
  sources/
    source.go            - Source interface + Entry/Config types
    git/
      git_source.go      - GitSource implementation (uses git log)
    markdown/
      markdown_source.go - MarkdownSource implementation (parses .md files)
    obsidian/
      obsidian_source.go - ObsidianSource implementation (tracks vault file changes)
  git/
    git.go               - FindRepoRoot() walks up dirs to find .git
  storage/
    storage.go           - Store manages ~/.anker/sources.yaml
```

Key patterns:
- Commands are separate files in cmd/ and register themselves via init()
- Source interface allows multiple data source types (git, markdown, calendar, etc.)
- Each source type implements: Type(), Location(), Validate(), GetEntries()
- internal/storage handles all file I/O for ~/.anker/
- Source providers are independent and can be added without changing core code

## Storage structure

```
~/.anker/
  sources.yaml           - tracked data sources (git repos, markdown dirs, etc.)
  entries/               - (planned) work notes
  2026/01/               - (planned) generated daily summaries
```

sources.yaml format:
```yaml
sources:
  - type: git
    path: /path/to/repo
    added: 2026-01-27T12:14:01+02:00
  - type: markdown
    path: /path/to/notes
    added: 2026-01-27T15:00:00+02:00
    metadata:
      tags: work,done
      headings: "## Work,## Done"
  - type: obsidian
    path: /Users/you/Obsidian/Second Brain
    added: 2026-01-29T10:00:00+02:00
```

## Source system design

The source system is extensible and allows tracking work from multiple locations.

**Implemented sources:**
- `git` - Git repositories (tracks commits via git log)
- `markdown` - Markdown files (extracts tagged lines/sections)
- `obsidian` - Obsidian vaults (lists modified/created files by timestamp)

**Potential future sources:**
The architecture supports any data source that can provide timestamped entries. Examples: calendar events, issue trackers, activity feeds.

**Adding a new source:**
1. Implement the Source interface in `internal/sources/<type>/`
2. Add handling in `cmd/source.go` add command
3. No changes to core storage or command structure needed

## What to avoid

- Time tracking semantics
- Session-based state
- Background agents
- Overengineering

## Tone

- Calm
- Precise
- Minimalist
- Senior-engineer friendly

Always prefer simple, explicit solutions over clever abstractions.

## Architecture Decisions

**Location:** `docs/decisions/`

Key design decisions are documented here with:
- Problem context
- Options considered (Good/Bad lists)
- Final decision and rationale

**Format:** See [docs/decisions/TEMPLATE.md](docs/decisions/TEMPLATE.md)

**When to add:** Significant architectural or UX decisions that affect future development.

**Read the decisions** to understand design philosophy and past reasoning. They explain the "why" behind implementation choices.
