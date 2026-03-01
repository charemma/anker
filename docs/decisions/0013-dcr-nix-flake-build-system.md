# 0013: Nix Flake Build System

**Date:** 2026-03-01
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai
**Supersedes:** 0012

## Problem

The build system used Just + Dagger (ADR 0012). Dagger requires Docker, adds container overhead for a simple Go CLI, and doesn't integrate with the Nix ecosystem the developer already uses across all systems (NixOS, home-manager, flakes in multiple repos).

## Options Considered

**Option A:** Keep Dagger
- Good: Already working, containerized reproducibility
- Bad: Requires Docker daemon, doesn't fit Nix workflow, extra layer of indirection

**Option B:** Nix flakes (replace Dagger and Just)
- Good: Native to existing toolchain, no Docker dependency, reproducible via Nix store
- Good: `nix flake check` covers tests + lint + build in one command
- Good: `nix develop` / direnv replaces manual tool installation
- Good: No wrapper layer needed -- Nix commands are direct enough
- Bad: Cross-compilation to macOS from Linux requires extra Nix config (or just use GOOS/GOARCH)

**Option C:** Hybrid (Dagger for CI, Nix for local)
- Good: Best of both
- Bad: Two build systems to maintain, defeats the purpose

## Decision

We chose **Option B**.

Nix flakes replace both Dagger and Just. The Nix CLI (`nix build`, `nix flake check`, `nix develop`) is direct enough that a wrapper layer adds no value. The flake provides:

- `packages.default` -- anker binary via `buildGoModule`
- `checks.tests` -- `go test ./...`
- `checks.lint` -- golangci-lint
- `checks.build` -- the package itself
- `checks.pre-commit` -- gofmt, typos, alejandra
- `devShells.default` -- go, gopls, golangci-lint + pre-commit hooks
- `formatter` -- alejandra

CI runs `nix flake check` instead of Dagger. GoReleaser release workflow stays unchanged.

Cross-compilation uses `GOOS`/`GOARCH` from the devShell rather than Nix cross-compilation, which keeps things simple.
