# 0010: Build Tool Selection

**Date:** 2026-01-28
**Status:** Superseded by [0012-dcr-build-system-architecture](0012-dcr-build-system-architecture.md)
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Need a tool to run common development tasks (test, lint, build) consistently across local development and CI/CD.

**Requirements:**
- Run tests, linting, builds
- Simple to use (`task test`, `task build`)
- Cross-platform (macOS, Linux, Windows)
- Works locally and in CI
- Ideally Go-native or neutral (no weird dependencies)

## Options Considered

### Option A: Taskfile (Task)

Modern task runner with YAML config:

```yaml
# Taskfile.yml
version: '3'
tasks:
  test:
    cmds: [go test ./...]
  lint:
    cmds: [golangci-lint run]
  build:
    cmds: [go build -o bin/anker .]
```

**Good:**
- Written in Go (fits Go project)
- Simple YAML syntax
- Cross-platform (single binary)
- Fast, parallel execution
- Good documentation
- Modern, active development
- Popular in Go community

**Bad:**
- Another dependency (but small, ~5MB)
- Less universal than Make

### Option B: Make (Makefile)

Traditional Unix build tool:

```makefile
# Makefile
test:
	go test ./...
lint:
	golangci-lint run
build:
	go build -o bin/anker .
```

**Good:**
- Universal, pre-installed on most systems
- Very established (decades old)
- No additional dependency

**Bad:**
- Arcane syntax (tabs vs spaces issues)
- Platform inconsistencies (GNU Make vs BSD Make)
- Less ergonomic than modern tools
- Not Go-specific
- Harder to read/maintain

### Option C: Justfile (Just)

Command runner similar to Make but modern:

```just
# justfile
test:
    go test ./...
lint:
    golangci-lint run
build:
    go build -o bin/anker .
```

**Good:**
- Modern, good syntax
- Fast
- Better than Make

**Bad:**
- Written in Rust (Rust dependency for Go project feels wrong)
- Less adoption in Go ecosystem
- User in platform-engineering-demo uses Just, but for Go project Task fits better

### Option D: Custom Shell Scripts

```bash
#!/bin/bash
# scripts/test.sh
go test ./...
```

**Good:**
- No dependencies
- Complete control
- Simple

**Bad:**
- Need separate scripts for each task
- No dependency management between tasks
- Platform differences (bash vs sh vs cmd)
- Reinventing the wheel
- Hard to compose tasks

### Option E: Dagger.io

Programmable CI/CD in Go with containers:

```go
// ci/main.go
func (m *Anker) Test(ctx context.Context) error {
    return dag.Container().
        From("golang:1.21").
        WithDirectory("/src", m.Source).
        WithExec([]string{"go", "test", "./..."}).
        Sync(ctx)
}
```

**Good:**
- Programmable in Go (great fit)
- Container-based (reproducible)
- Local = CI (truly identical)
- Modern approach
- Great for complex builds

**Bad:**
- Overkill for simple Go CLI
- Docker dependency (heavyweight)
- Learning curve
- More complexity than needed
- Better suited for multi-service, complex builds
- Can add later if needed

### Option F: pre-commit Framework

Python-based hook manager:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    hooks:
      - id: golangci-lint
```

**Good:**
- Comprehensive hook management
- Many pre-built hooks

**Bad:**
- Python dependency for Go project (weird)
- Focused on git hooks, not general task running
- Not common in Go community
- Overkill for simple test/build tasks

## Decision

**Use Taskfile (Task) for task orchestration.**

**Why:**
- ✅ Go-native (Go project uses Go tool)
- ✅ Simple YAML syntax (easy to read/write)
- ✅ Cross-platform (works everywhere)
- ✅ Fast, parallel execution
- ✅ Good fit for Go ecosystem
- ✅ Modern, actively maintained
- ❌ Not Make (more consistent, better UX)
- ❌ Not Dagger (too complex for simple CLI)
- ❌ Not pre-commit (Python dependency)

**Taskfile.yml:**
```yaml
version: '3'

tasks:
  test:
    desc: Run all tests
    cmds:
      - go test -v ./...

  test-coverage:
    desc: Run tests with coverage report
    cmds:
      - go test -coverprofile=coverage.out ./...
      - go tool cover -html=coverage.out

  lint:
    desc: Run golangci-lint
    cmds:
      - golangci-lint run

  build:
    desc: Build anker binary
    cmds:
      - go build -o bin/anker .

  check:
    desc: Run all quality checks (tests, lint, build)
    deps: [test, lint, build]

  dev:
    desc: Quick dev check (test + build, skip lint for speed)
    deps: [test, build]

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -rf bin/ dist/ coverage.out
```

**Usage:**
```bash
task test           # Run tests
task lint           # Run linter
task build          # Build binary
task check          # Run all checks
task --list         # Show all available tasks
```

**In CI (GitHub Actions):**
```yaml
# Separate jobs for better visibility
jobs:
  test:
    steps:
      - run: task test
  lint:
    steps:
      - run: task lint
  build:
    steps:
      - run: task build
```

## When to Reconsider

**Switch to Dagger if:**
- Build becomes complex (multi-stage Docker, multiple services)
- Need true environment reproducibility (container isolation)
- Building for multiple platforms with different requirements
- Team grows and needs more sophisticated CI/CD

**For now:** Taskfile is perfect. Simple, fast, Go-native.

## Installation

```bash
# macOS
brew install go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Windows
choco install go-task

# Or: go install (always works)
go install github.com/go-task/task/v3/cmd/task@latest
```

## Alternative Considered

User uses Just in platform-engineering-demo, but for Go project Task is better fit (Go-native vs Rust-native).
