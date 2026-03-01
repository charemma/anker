# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**anker** is a Go CLI tool for reconstructing your workday after the fact. No time tracking, no background agents -- just explicit, local, text-first summaries built from data sources you already have.

## Design principles

- Local-first: all data stays on your machine, no network dependencies
- Explicit over implicit: nothing is tracked automatically, every source is added by the user
- Deferred analysis: work first, summarize later
- CLI arguments override configuration
- Simple and predictable -- prefer straightforward solutions over clever abstractions
- Not a time tracker, not a productivity optimizer, not a daemon

## Development commands

Enter the dev shell (provides go, gopls, golangci-lint):
- `nix develop` or use direnv (`.envrc` with `use flake`)

Build and run:
- `nix build` -- reproducible build to `./result/bin/anker`
- `go build -o bin/anker .` -- quick build during development
- `go run . recap today` -- quick run without building

Testing and checks:
- `nix flake check` -- all checks (tests + lint + build + pre-commit)
- `go test ./...` -- quick tests during development
- `go test ./internal/timerange/...` -- test a specific package
- `go test -run TestParseToday ./internal/timerange/` -- run a single test

## Code architecture

```
main.go              -- entry point, calls cmd.Execute()
flake.nix            -- Nix flake (package, checks, devShell)
cmd/                  -- Cobra commands, each file registers via init()
  root.go             -- root command + version info (ldflags: Version, Commit, Date)
  source.go           -- source add/list/remove subcommands
  recap.go            -- recap command with output formatting (simple/detailed/json/markdown)
internal/
  sources/
    source.go          -- Source interface (Type, Location, Validate, GetEntries) + Entry/Config types
    git/               -- GitSource: commits via git log, diff enrichment for markdown format
    markdown/          -- MarkdownSource: extracts tagged lines/sections from .md files
    obsidian/          -- ObsidianSource: tracks vault file changes by timestamp
  config/              -- User config from ~/.anker/config.yaml (week_start, author_email)
  storage/             -- Store manages ~/.anker/sources.yaml (add/list/remove sources)
  timerange/           -- Flexible time spec parser (today, thisweek, "october 2025", date ranges)
    locales/           -- Multilingual month name support (en, de); add new locales here
  git/                 -- Git helpers: FindRepoRoot, GetAuthorEmail from git config
  paths/               -- GetAnkerHome() resolves ANKER_HOME env var or defaults to ~/.anker
```

Key patterns:
- Source interface allows adding new data source types without changing core code. Implement in `internal/sources/<type>/`, wire up in `cmd/source.go`.
- `cmd/recap.go` collects entries from all sources, sorts by timestamp, and formats output. The markdown format enriches git entries with full diffs.
- `internal/timerange` parses human-friendly time specs into `TimeRange{From, To}`. Locale-aware month names are registered in `locales/`.
- Config resolution chain: CLI flags > `~/.anker/config.yaml` > git config fallback (for author_email).

## Storage

All state lives in `~/.anker/` (overridable via `ANKER_HOME` env var):
- `config.yaml` -- user configuration (week_start, author_email)
- `sources.yaml` -- registered data sources with type, path, metadata

Tests use `ANKER_HOME` pointed at `t.TempDir()` to isolate filesystem state.

## Build system

Nix flake for reproducible builds, checks, and dev environment. All tools are pinned via `flake.lock`.

The flake provides: package build via `buildGoModule`, checks (tests, lint, build, pre-commit hooks), and a devShell with all required tools. CI runs `nix flake check`.

## Key design decisions

Documented in `docs/decisions/` as numbered ADRs. Read these before making architectural changes.

Most relevant decisions:
- Sources require explicit registration (`anker source add git .`), no auto-discovery (ADR 0001)
- Source type is a positional argument, not a flag or auto-detected (ADR 0002)
- Future community extensions via executable plugins `anker-source-<type>` with JSON protocol (ADR 0003)
- Git-style aliases planned for user-defined shortcuts (ADR 0004)
- Report templating via Lua planned but not yet implemented; current output formats (simple/detailed/json/markdown) are hardcoded in Go (ADR 0009)
- Nix flake build system, replacing earlier Just + Dagger setup (ADR 0013, supersedes ADR 0012)
