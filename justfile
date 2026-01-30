# anker development commands
# Thin wrapper around Dagger build system



_default:
    @just --list

# Run all tests
test:
    dagger call test --source=.

# Run linter
lint:
    dagger call lint --source=.

# Build binary for current platform (macOS ARM64)
build-osx:
    @mkdir -p bin
    dagger call build --source=. --goos=darwin --goarch=arm64 export --path=./bin/anker

# Build binary for Linux (CI/production)
build-linux:
    @mkdir -p bin
    dagger call build --source=. --goos=linux --goarch=amd64 export --path=./bin/anker

# Run all quality checks (test + lint + build)
check:
    dagger call check --source=.

# Run tests with coverage report
coverage:
    dagger call coverage --source=.

# Show available Dagger functions
functions:
    dagger functions

# Clean build artifacts
clean:
    rm -rf bin/ dist/ coverage.out

# Install development dependencies and git hooks
setup:
    @echo "Checking prerequisites..."
    @if ! command -v dagger >/dev/null 2>&1; then \
        echo "❌ Dagger not found. Install from https://docs.dagger.io/install"; \
        exit 1; \
    fi
    @echo "✓ Dagger installed: $$(dagger version)"
    @echo ""
    @echo "Installing git hooks..."
    @mkdir -p .git/hooks
    @echo '#!/bin/bash\necho "🔍 Running pre-push checks..."\njust check\nif [ $$? -ne 0 ]; then\n  echo "❌ Checks failed. Fix issues or skip with: git push --no-verify"\n  exit 1\nfi\necho "✅ All checks passed!"' > .git/hooks/pre-push
    @chmod +x .git/hooks/pre-push
    @echo "✓ Git hooks installed"
    @echo ""
    @echo "Ready! Run 'just check' to verify everything works."
