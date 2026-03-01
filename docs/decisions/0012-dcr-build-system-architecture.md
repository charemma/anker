# 0012: Build System Architecture

**Date:** 2026-01-29
**Status:** Superseded by 0013
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

A Go CLI project needs a build system that provides:
- Reproducible builds across different developer environments
- Consistent execution in local development and CI/CD
- Simple, discoverable developer interface
- Container-based isolation for tests and builds

How should build tasks be organized? What tools should execute build logic vs. provide developer interface?

## Options Considered

### Option A: Just (thin wrapper) + Dagger (build engine)

**Architecture:**
```
Developer → Just (UX wrapper) → Dagger (containerized build logic)
```

**Just provides:**
- Simple command interface (`just test`, `just build`)
- Shell-like syntax (familiar, minimal abstraction)
- No build logic (pure wrapper)

**Dagger provides:**
- All build logic (test, lint, build)
- Container-based execution (golang:1.24-alpine)
- Reproducible environments
- Composable functions

**Good:**
- Clear separation of concerns (UX vs logic)
- Build logic is containerized (reproducible)
- Local development = CI (same Dagger code)
- Just commands hide verbose Dagger syntax
- Easy to understand (wrapper is transparent)
- No duplication of logic
- Upgradeable (swap Just for another wrapper without touching build logic)

**Bad:**
- Two tools instead of one
- Requires Dagger installation (Docker dependency)
- Initial container download takes time

### Option B: Taskfile only

**Architecture:**
```
Developer → Taskfile (YAML-based task runner)
```

**Good:**
- Single tool
- YAML syntax (declarative)
- Go-native (written in Go)
- Popular in Go ecosystem

**Bad:**
- Build logic lives in Taskfile (not containerized)
- YAML becomes complex for advanced logic
- Duplicates abstractions (Taskfile deps + functions)
- Not reproducible (runs natively, environment-dependent)
- CI and local may diverge

### Option C: Make only

**Architecture:**
```
Developer → Makefile (traditional build system)
```

**Good:**
- Universal (pre-installed)
- Shell-like syntax
- Proven at scale

**Bad:**
- Build logic in Makefile (not containerized)
- Arcane syntax (tabs vs spaces, special variables)
- Not reproducible across environments
- Platform differences (GNU Make vs BSD Make)

### Option D: Dagger only (no wrapper)

**Architecture:**
```
Developer → Dagger directly
```

**Example:**
```bash
dagger call test --source=.
dagger call build --source=. --export-path=./bin/anker
```

**Good:**
- Single tool
- No abstraction layer
- Most direct

**Bad:**
- Verbose (`--source=.` on every command)
- Poor discoverability (must run `dagger functions`)
- Not ergonomic for daily use
- Higher barrier to entry

### Option E: Docker Compose for builds

**Architecture:**
```
Developer → docker-compose run <service>
```

**Good:**
- Container-based
- Familiar to many developers

**Bad:**
- YAML configuration complexity
- Not designed for build pipelines
- Limited composability
- Awkward for single-command workflows

## Decision

**Option A: Just + Dagger**

### Architecture

**Just (Developer Interface):**
```just
# justfile
test:
    dagger call test --source=.

build-osx:
    @mkdir -p bin
    dagger call build --source=. --goos=darwin --goarch=arm64 export --path=./bin/anker

build-linux:
    @mkdir -p bin
    dagger call build --source=. --goos=linux --goarch=amd64 export --path=./bin/anker

check:
    dagger call check --source=.
```

**Dagger (Build Logic):**
```go
// ci/main.go
func (m *Anker) Test(ctx context.Context, source *dagger.Directory) error {
    _, err := dag.Container().
        From("golang:1.24-alpine").
        WithExec([]string{"apk", "add", "--no-cache", "git"}).
        WithDirectory("/src", source).
        WithWorkdir("/src").
        WithExec([]string{"go", "test", "-v", "./..."}).
        Sync(ctx)
    return err
}

func (m *Anker) Build(ctx context.Context, source *dagger.Directory, goos string, goarch string) *dagger.File {
    return dag.Container().
        From("golang:1.24-alpine").
        WithDirectory("/src", source).
        WithWorkdir("/src").
        WithEnvVariable("GOOS", goos).
        WithEnvVariable("GOARCH", goarch).
        WithExec([]string{"go", "build", "-o", "anker", "."}).
        File("/src/anker")
}
```

### Why This Architecture

**Separation of Concerns:**

| Layer | Responsibility | Technology |
|-------|---------------|------------|
| Developer UX | Command discoverability, simple invocation | Just |
| Build Logic | Tests, linting, compilation, containerization | Dagger |

**UX Layer (Just):**
- Provides familiar, shell-like commands
- Hides Dagger's verbose syntax
- No logic, pure delegation
- Easy to replace if needed

**Build Engine (Dagger):**
- All logic lives in Go code (type-safe, debuggable)
- Container-based (golang:1.24-alpine guarantees consistency)
- Cross-compilation support (build for any platform from any platform)
- Composable functions (Test, Lint, Build, Check)
- Same code runs locally and in CI
- No "works on my machine" - all builds containerized

**Why Just over Taskfile:**

| Aspect | Just | Taskfile |
|--------|------|----------|
| Syntax | Shell-like, minimal | YAML, more abstraction |
| Purpose | Command wrapper | Task runner with logic |
| Temptation | Stay thin (shell commands only) | Add logic (YAML deps, conditions) |
| Overhead | ~1 line per command | More configuration |
| Philosophy | Explicit delegation | Self-contained tasks |

**Key distinction:** Taskfile encourages putting build logic in YAML (deps, multi-step tasks). Just encourages staying thin and delegating.

**Why Just over Make:**

Just is preferred for syntax clarity, ergonomics, and cross-platform installation:
- No tab-vs-space issues
- Better error messages
- Simpler recipe syntax
- Still shell-first philosophy (like Make)
- Easy installation on all platforms (Windows/macOS/Linux)
  - Make (GNU) has inconsistent availability and versions across platforms
  - Just provides single binary installation via package managers everywhere

Make would work but requires more care with portability and installation.

**Why Dagger over native builds:**

| Aspect | Dagger | Native (go test, go build) |
|--------|--------|---------------------------|
| Environment | Container (isolated) | Host machine |
| Reproducibility | Same everywhere | Varies by Go version, OS |
| CI parity | Identical code | Different scripts |
| Dependencies | Declared in code | Implicit (installed tools) |
| Debugging | Go code | Shell scripts |

**Workflow:**

```bash
# Developer (daily use)
just test           # Containerized tests
just build-osx      # Containerized build for macOS ARM64
just build-linux    # Containerized build for Linux
just check          # Full validation (tests + lint + build)

# CI (GitHub Actions)
dagger call check --source=.   # Same underlying code (builds for Linux)
```

**Cross-compilation:**
- Local macOS development: `--goos=darwin --goarch=arm64`
- CI/production Linux: `--goos=linux --goarch=amd64`
- All builds happen in containers (no native builds)
- Eliminates "works on my machine" completely

**Maintenance benefits:**

1. **Change build logic:** Edit `ci/main.go` (Go code, type-checked)
2. **Change UX:** Edit `justfile` (shell commands, simple)
3. **Swap wrapper:** Replace Just with another tool, Dagger unchanged
4. **Add CI provider:** Same Dagger code works everywhere

### Trade-offs Accepted

**Additional complexity:**
- Two tools (Just + Dagger) vs one (Taskfile or Make)
- Requires Docker for Dagger

**Why acceptable:**
- Clear separation prevents logic drift
- Container benefits outweigh installation overhead
- Just wrapper is simple enough to understand in minutes

**Build speed:**
- Container overhead vs native execution
- Mitigated by Dagger's caching

**Developer onboarding:**
- Must install Dagger
- Mitigated by `just setup` checking prerequisites

## Implementation

**Project structure:**
```
anker/
├── justfile              # Developer commands (UX)
├── ci/
│   └── main.go          # Dagger pipeline (logic)
├── .github/workflows/
│   └── ci.yml           # Uses Dagger directly
```

**Example justfile:**
```just
test:
    dagger call test --source=.

build:
    dagger call build --source=.

check:
    dagger call check --source=.
```

**Example CI:**
```yaml
# .github/workflows/ci.yml
- uses: dagger/dagger-for-github@v5
- run: dagger call check --source=.
```

**Key principle:** justfile never implements build logic. All logic in Dagger.

## Alternative Considered

Dagger provides a Python SDK that could eliminate the wrapper entirely by generating a CLI. However:
- Adds Python dependency to Go project
- Less control over UX
- More complex than simple Just wrapper

## Future Considerations

If Just becomes limiting:
- Wrapper can be replaced (Make, Task, custom script)
- Dagger code remains unchanged
- No rewrite of build logic required

If Dagger becomes too complex:
- Unlikely given current simplicity (5 functions)
- Could extract common patterns to helper functions
- Go code is easier to refactor than shell/YAML
