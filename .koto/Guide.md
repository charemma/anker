# ikno (ikno) -- Agent Guide

You are working on **ikno**, a Go CLI tool for reconstructing your workday after the fact. No time tracking, no background agents -- just explicit, local, text-first summaries built from data sources you already have.

## Repository

- Language: Go 1.24, module `github.com/charemma/ikno`
- Build: Nix flake (`nix build`, `nix flake check`)
- Quick dev: `go build -o bin/ikno .` / `go test ./...`
- Dev shell: `nix develop` (provides go, gopls, golangci-lint)

## Architecture

```
main.go        -- entry, calls cmd.Execute()
cmd/           -- Cobra commands (root, source, recap)
internal/
  sources/     -- Source interface + implementations (git, markdown, obsidian)
  config/      -- User config from ~/.config/ikno/config.yaml
  storage/     -- Source registry (~/.config/ikno/sources.yaml)
  timerange/   -- Human-friendly time spec parser (today, thisweek, "october 2025")
  git/         -- Git helpers
  paths/       -- Config dir resolution (IKNO_HOME env var)
```

## Key Patterns

- Source interface: `Type()`, `Location()`, `Validate()`, `GetEntries(TimeRange)`
- New sources: implement in `internal/sources/<type>/`, wire in `cmd/source.go`
- Config chain: CLI flags > config.yaml > git config fallback
- Tests use `IKNO_HOME=t.TempDir()` for isolation
- All state in `~/.config/ikno/` (overridable via IKNO_HOME)

## Design Principles

- Local-first: all data stays on machine, no network
- Explicit over implicit: nothing auto-tracked
- Deferred analysis: work first, summarize later
- CLI args override config
- Prefer configurable values over hardcoded ones where it matters -- timeouts, limits, URLs, thresholds should be settable via config or flags. Constants that are genuinely fixed (buffer sizes, internal defaults) are fine hardcoded.
- Simple and predictable -- no clever abstractions

## Code Style

- Go idioms: accept interfaces, return structs
- errcheck enforced: handle all fmt.Fprint* return values with `_, _ =`
- Modern stdlib: slices.Contains, log/slog (Go 1.21+)
- Table-driven tests, t.TempDir() for filesystem tests
- Conventional commits, feature branches

## Running Flows

```
koto up review-pr pr=67          # review a PR
koto up fix-issue issue=68       # fix a GitHub issue
koto up development -t "Add obsidian source type for daily notes"
```

Outputs are written to `~/.koto/stacks/ikno/`. Steps with `print_output: true` display their result in the terminal after the flow completes.

## Decisions

ADRs live in `docs/decisions/`. They are for internal use by the maintainer.

## Output Style

- No emojis or icons in any output
- Plain, clean markdown

## Review Guidelines

When reviewing PRs from external contributors:
- Never suggest creating ADRs or decision records
- Documentation suggestions only when the PR introduces something genuinely new
- Keep suggestions practical and actionable
- Be constructive and welcoming -- suggest improvements, don't demand rewrites
