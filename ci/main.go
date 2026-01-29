// Anker CI/CD pipeline
package main

import (
	"context"
	"dagger/anker/internal/dagger"
	"fmt"
)

type Anker struct{}

// Default action: Run all checks
func (m *Anker) Default(
	ctx context.Context,
	// Project source code
	source *dagger.Directory,
) (string, error) {
	return m.Check(ctx, source)
}

// Run all tests
func (m *Anker) Test(
	ctx context.Context,
	// Project source code
	source *dagger.Directory,
) error {
	_, err := dag.Container().
		From("golang:1.24-alpine").
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"go", "test", "-v", "./..."}).
		Sync(ctx)
	return err
}

// Run linter
func (m *Anker) Lint(
	ctx context.Context,
	// Project source code
	source *dagger.Directory,
) error {
	_, err := dag.Container().
		From("golang:1.24-alpine").
		WithExec([]string{"apk", "add", "--no-cache", "git", "gcc", "musl-dev", "binutils"}).
		WithExec([]string{"go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"}).
		WithDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"golangci-lint", "run", "--timeout=5m"}).
		Sync(ctx)
	return err
}

// Build anker binary and return as file
func (m *Anker) Build(
	ctx context.Context,
	// Project source code
	source *dagger.Directory,
	// Target OS (default: linux)
	// +optional
	// +default="linux"
	goos string,
	// Target architecture (default: amd64)
	// +optional
	// +default="amd64"
	goarch string,
) *dagger.File {
	return dag.Container().
		From("golang:1.24-alpine").
		WithDirectory("/src", source).
		WithWorkdir("/src").
		WithEnvVariable("GOOS", goos).
		WithEnvVariable("GOARCH", goarch).
		WithExec([]string{"go", "build", "-o", "anker", "."}).
		File("/src/anker")
}

// Run all quality checks (test + lint + build)
func (m *Anker) Check(
	ctx context.Context,
	// Project source code
	source *dagger.Directory,
) (string, error) {
	// Run tests
	fmt.Println("→ Running tests...")
	if err := m.Test(ctx, source); err != nil {
		return "", fmt.Errorf("tests failed: %w", err)
	}
	fmt.Println("✓ Tests passed")

	// Run linter
	fmt.Println("→ Running linter...")
	if err := m.Lint(ctx, source); err != nil {
		return "", fmt.Errorf("linter failed: %w", err)
	}
	fmt.Println("✓ Linter passed")

	// Build binary (Linux for CI)
	fmt.Println("→ Building binary...")
	binary := m.Build(ctx, source, "linux", "amd64")
	if _, err := binary.Export(ctx, "./bin/anker"); err != nil {
		return "", fmt.Errorf("build failed: %w", err)
	}
	fmt.Println("✓ Build succeeded at ./bin/anker")

	return "All checks passed! Binary at ./bin/anker", nil
}

// Run tests with coverage report
func (m *Anker) Coverage(
	ctx context.Context,
	// Project source code
	source *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("golang:1.24-alpine").
		WithDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"go", "test", "-coverprofile=coverage.out", "./..."}).
		WithExec([]string{"go", "tool", "cover", "-func=coverage.out"}).
		Stdout(ctx)
}
