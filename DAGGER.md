# Dagger CI/CD

anker uses Dagger for reproducible, containerized builds.

For daily development, use the Just wrapper (see README.md).

This document is for direct Dagger usage (CI, debugging, or advanced use cases).

## Installation

```bash
# macOS/Linux
curl -L https://dl.dagger.io/dagger/install.sh | sh

# Or via Homebrew (macOS)
brew install dagger/tap/dagger

# Verify installation
dagger version
```

## Direct Usage

For daily development, use `just` commands instead (simpler).

Direct Dagger usage (all commands run from project root):

### List available functions
```bash
dagger functions

# Or with more details
dagger call --help
```

### Run all checks (test + lint + build)
```bash
# Default action
dagger call default --source=.

# Or explicitly
dagger call check --source=.
```

### Individual commands

**Run tests:**
```bash
dagger call test --source=.
```

**Run linter:**
```bash
dagger call lint --source=.
```

**Build binary (exports to ./bin/anker):**
```bash
dagger call build --source=.

# Custom output path
dagger call build --source=. --output=./anker
```

**Run tests with coverage:**
```bash
dagger call coverage --source=.
```

## What Dagger Does

**Reproducibility:**
- Runs in Docker containers (same environment everywhere)
- Local development = CI (no "works on my machine")
- Versioned container images (golang:1.21, golangci-lint:v1.55)

**Speed:**
- Caching (rebuilds only what changed)
- Parallel execution where possible

**Simplicity:**
- One tool for local and CI
- No shell scripts, no YAML
- Go code you can debug

## Development Workflow

```bash
# Quick iteration during development
go test ./...               # Fast, native
go build -o bin/anker .     # Fast, native

# Before merging to main
dagger call check --source=.  # Full CI checks (Docker)
```

## CI/CD (GitHub Actions)

```yaml
# .github/workflows/ci.yml
- uses: dagger/dagger-for-github@v5
- run: dagger call check --source=.
```

## Why Dagger?

**vs Taskfile/Make:**
- Reproducible (Docker containers)
- Not shell-dependent (works everywhere)
- Cacheable, faster CI

**vs Docker commands:**
- No complex docker run commands
- Composable functions
- Better caching

**vs Custom CI:**
- Local = CI (test locally before push)
- Portable across CI providers
- Version controlled (Go code)

## Troubleshooting

**Dagger not found:**
```bash
# Add to PATH (after install)
export PATH="$HOME/.local/bin:$PATH"
```

**Docker errors:**
```bash
# Dagger requires Docker
docker --version  # Should work
```

**Slow first run:**
- Dagger downloads container images on first run
- Subsequent runs are cached and fast
