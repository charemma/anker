# 0011: Code Quality Enforcement

**Date:** 2026-01-28
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

How do we ensure only tested, quality-checked code makes it into the main branch?

**Context:**
- Solo development now, public with contributors later
- main branch should always be production-ready
- Need balance between discipline and pragmatism
- Want local feedback (fast) and CI enforcement (reliable)

**Requirements:**
- Tests must pass before merge
- Code must be linted
- Binary must build successfully
- Process shouldn't slow down solo development
- Must scale to public contributors later

## Options Considered

### Enforcement Strategies

**Option A: Trust-based (Manual)**

Developer manually runs checks:
```bash
task check  # Run before merge, no enforcement
git checkout main
git merge feature/xyz
```

**Good:**
- Simple, no tooling overhead
- Fast iteration during development
- Full control

**Bad:**
- Easy to forget
- No enforcement
- Contributors might skip checks
- Relies on discipline

**Option B: Git Hooks (Local Enforcement)**

Hooks run automatically on git operations:
```bash
# .git/hooks/pre-push
task check || exit 1
```

**Good:**
- Automatic local checks
- Catches issues before push
- Fast feedback

**Bad:**
- Can be bypassed (--no-verify)
- Not in repo (each developer must install)
- Different developers might have different hooks
- Fragile (hooks can break)

**Option C: GitHub Branch Protection (Server Enforcement)**

GitHub enforces checks before merge:
- Required status checks
- CI must pass
- Cannot be bypassed

**Good:**
- Cannot be bypassed
- Works for all contributors
- Centrally managed
- Standard industry practice

**Bad:**
- Only works for public/team repos (not solo private initially)
- Slower feedback (wait for CI)
- Requires GitHub Actions setup

**Option D: Combination (Defense in Depth)**

Hooks (fast, local) + Branch Protection (enforced, server):
- Hooks catch issues early
- Branch Protection ensures nothing slips through
- Best of both worlds

**Good:**
- Early feedback (hooks)
- Final enforcement (branch protection)
- Scales from solo to team

**Bad:**
- More complex setup
- Two systems to maintain

**Note:** Tool selection (Just + Dagger architecture) is covered in [[0012-dcr-build-system-architecture]]. This DCR focuses on enforcement strategy.

## Decision

**Two-Phase Approach: Just + Dagger now, Branch Protection when public**

### Phase 1: Solo Development (Private Repo - Temporary)

**Tools:**
- **Just** for developer interface (thin wrapper)
- **Dagger** for all build logic (containerized)
- **Manual checks** before merge (temporary workaround)
- **No CI yet** (not needed while private and solo)

**Temporary Workflow (only until public):**
```bash
# Feature development
git checkout -b feature/xyz
# ... work ...

# Before merge to main (MANUAL CHECK - temporary only!)
just check

# If all checks pass, merge
git checkout main
git merge feature/xyz --no-ff
git push
```

**Important:** This manual merge workflow is **TEMPORARY** and only acceptable because:
- Repository is private
- Solo developer
- Pre-publication development

Once public, this workflow is **NOT ALLOWED**. Must use proper CI/CD flow (Phase 2).

**justfile (thin wrapper):**
```just
# Run all tests
test:
    dagger call test --source=.

# Run linter
lint:
    dagger call lint --source=.

# Build binary
build:
    dagger call build --source=.

# Run all quality checks (test + lint + build)
check:
    dagger call check --source=.

# Run tests with coverage report
coverage:
    dagger call coverage --source=.
```

**ci/main.go (Dagger pipeline with actual build logic):**
```go
// All build logic lives here in containerized Go code
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
```

**GitHub Actions (.github/workflows/ci.yml):**
```yaml
name: CI
on:
  push:
    branches: ['feature/*']  # NOT main (main only via PR merge)
  pull_request:
    branches: [main]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: dagger/dagger-for-github@v5
      - name: Run all checks
        run: dagger call check --source=.
```

**Important:** GitHub Actions uses Dagger directly (same as local `just check`).

**Why single job:**
- Same `check` function as local development
- Dagger handles parallelization internally
- No duplication of logic between CI and local
- Simpler workflow

**Why this approach:**
- **Pragmatic**: No hooks overhead for solo dev
- **Fast**: Local iteration not slowed down
- **Reproducible**: Containerized builds via Dagger
- **Simple**: Just wraps Dagger, CI uses Dagger directly
- **Scales**: Easy to add hooks/protection later

### Phase 2: Public Release (Proper CI/CD - Required before publication)

**This is the REAL workflow. Phase 1 is just a temporary exception.**

When repository becomes public (or even before):

**Add Git Hooks (optional, helpful):**
```just
# justfile addition
setup:
    @echo "Installing git hooks..."
    @mkdir -p .git/hooks
    @echo '#!/bin/bash\necho "🔍 Running pre-push checks..."\njust check\nif [ $$? -ne 0 ]; then\n  echo "❌ Checks failed. Fix issues or skip with: git push --no-verify"\n  exit 1\nfi\necho "✅ All checks passed!"' > .git/hooks/pre-push
    @chmod +x .git/hooks/pre-push
    @echo "✓ Git hooks installed"
```

**Enable GitHub Branch Protection (REQUIRED):**
- Settings → Branches → Add rule for `main`
- ✅ Require status checks to pass before merging
- ✅ Require pull request before merging
- ✅ Require branches to be up to date before merging
- ✅ Include administrators (even you must pass checks)
- Required status checks: `test`, `lint`, `build`

**Proper Workflow (enforced by branch protection):**
```bash
# Feature development
git checkout -b feature/xyz
# ... work ...
git push origin feature/xyz

# Create Pull Request on GitHub
# → CI runs automatically (test, lint, build)
# → Review changes
# → Merge via GitHub UI (only if CI green)

# Never: git checkout main && git merge feature/xyz
# Always: Merge via Pull Request!
```

**README documentation:**
```markdown
## Development

1. Clone the repository
2. Run `just setup` to install git hooks (optional)
3. Create feature branch: `git checkout -b feature/xyz`
4. Work on your feature, commit changes
5. Push: `git push origin feature/xyz`
6. Create Pull Request on GitHub
7. Wait for CI to pass (containerized checks via Dagger)
8. Merge via GitHub UI (never merge locally to main)
```

**Critical Rules (Phase 2):**
- ❌ **Never merge locally to main** (`git merge` forbidden)
- ✅ **Always via Pull Request** (enforced by branch protection)
- ✅ **CI must be green** before merge possible
- ✅ **No bypass** (not even for repo owner)

**Why this is non-negotiable:**
- Ensures all code is tested before reaching main
- Creates audit trail (PR history)
- Standard industry practice
- Scales to contributors
- Main branch is always deployable

### Tools: Just + Dagger

Using Just (thin wrapper) + Dagger (build engine). See [[0012-dcr-build-system-architecture]] for rationale.

### Checks Included

**just check runs (via Dagger):**
1. **Tests** - All tests in containerized environment
2. **Lint** - Code quality checks
3. **Build** - Binary must compile

All checks run in `golang:1.24-alpine` containers for reproducibility.

## Benefits

**For solo development (Phase 1 - temporary):**
- Fast iteration (no CI overhead)
- Manual discipline with `just check`
- Containerized builds (reproducible)
- Acceptable only while private
- Must upgrade to Phase 2 before publication

**For proper development (Phase 2 - permanent):**
- Pull Request workflow (standard practice)
- CI enforces quality (automatic)
- Branch Protection (no bypass)
- Clear, professional process
- Scales to contributors

**For maintenance:**
- Separation of concerns (Just for UX, Dagger for logic)
- Easy to extend (add more Dagger functions)
- CI = local (same Dagger code)
- Transparent (justfile and ci/main.go in repo)
- Reproducible (container-based)

## Implementation Notes

**CI must run on:**
- Every push to feature branches
- Every pull request to main
- **NOT on direct pushes to main** (because those shouldn't happen!)

**Status badge in README:**
```markdown
[![CI](https://github.com/charemma/anker/workflows/CI/badge.svg)](https://github.com/charemma/anker/actions)
```

Shows visitors the project is well-tested.

## Future Considerations

**Could add later:**
- Coverage requirements (e.g., >80%)
- Security scanning (gosec, govulncheck)
- Dependency updates (dependabot)
- Release automation (goreleaser)
- Cross-platform tests (Linux, macOS, Windows) as additional Dagger functions

**Dagger benefits already realized:**
- Reproducible container-based builds
- Local = CI (same Dagger code everywhere)
- Type-safe build logic in Go
- Easy to extend with new functions
