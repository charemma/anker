# Building and Testing

This guide covers the build system, development workflow, and CI/CD setup for anker.

## Overview

anker uses a **Nix flake** for reproducible builds, tests, and linting. All tools and dependencies are pinned via `flake.lock` -- no manual installation needed.

## Prerequisites

- [Nix](https://nixos.org/) with flakes enabled
- Optional: [direnv](https://direnv.net/) for automatic shell activation

## Quick Reference

```bash
# Build the binary
nix build                    # -> ./result/bin/anker

# Run all checks (tests + lint + build + pre-commit)
nix flake check

# Enter development shell (go, gopls, golangci-lint)
nix develop

# Format Nix files
nix fmt
```

## Development Shell

The flake provides a dev shell with all required tools:

```bash
# Enter manually
nix develop

# Or use direnv (activates automatically when entering the directory)
direnv allow
```

The dev shell includes:
- `go` (compiler + tools)
- `gopls` (language server)
- `golangci-lint` (linter)
- Pre-commit hooks (gofmt, typos, alejandra)

## Building

```bash
# Reproducible build via Nix (output in ./result/bin/anker)
nix build

# Quick build during development (from dev shell)
go build -o bin/anker .

# Quick run without building
go run . recap today
```

## Testing

```bash
# Run tests via Nix (sandboxed, reproducible)
nix build .#checks.x86_64-linux.tests

# Quick tests during development (from dev shell)
go test ./...
go test ./internal/timerange/...
go test -run TestParseToday ./internal/timerange/
```

## Checks

`nix flake check` runs all of these:

| Check | What it does |
|---|---|
| `checks.build` | Builds the binary via `buildGoModule` |
| `checks.tests` | Runs `go test ./...` in sandbox |
| `checks.lint` | Runs `golangci-lint` in sandbox |
| `checks.pre-commit` | gofmt, typos, alejandra |

Run a single check:
```bash
nix build .#checks.x86_64-linux.lint
```

## CI/CD

GitHub Actions runs `nix flake check` on every push and PR. See `.github/workflows/ci.yml`.

Releases use GoReleaser via `.github/workflows/release.yml` (independent of the Nix build system).

## Cross-Compilation

From the dev shell:
```bash
GOOS=darwin GOARCH=arm64 go build -o bin/anker .   # macOS ARM64
GOOS=linux GOARCH=amd64 go build -o bin/anker .    # Linux AMD64
```
