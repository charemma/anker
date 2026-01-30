# Architecture Decision Records

Lightweight documentation of important design decisions.

## Format

Each decision follows: **Problem → Options → Decision**

```markdown
## Problem
What are we solving?

## Options Considered
**Option A:**
- Good: Reason 1
- Bad: Reason 2

**Option B:**
- Good: Reason 3
- Bad: Reason 4

## Decision
We chose Option A.
Why: Brief explanation.
```

See [TEMPLATE.md](TEMPLATE.md) for the full template.

Quick to scan, easy to understand the reasoning.

## Records

### Core Philosophy
- [0001-dcr-source-discovery-strategy.md](0001-dcr-source-discovery-strategy.md) - No automatic discovery
- [0002-dcr-source-type-specification.md](0002-dcr-source-type-specification.md) - `git` as argument, not flag

### Extensibility
- [0003-dcr-community-source-extensions.md](0003-dcr-community-source-extensions.md) - Plugin system for community sources
- [0004-dcr-command-shortcuts.md](0004-dcr-command-shortcuts.md) - User-defined command shortcuts
- [0005-dcr-report-output-format.md](0005-dcr-report-output-format.md) - Customizable report output

### Technical Choices
- [0006-dcr-cli-framework-selection.md](0006-dcr-cli-framework-selection.md) - CLI framework selection
- [0007-dcr-data-storage-strategy.md](0007-dcr-data-storage-strategy.md) - Text files with TOML format
- [0008-dcr-config-file-location.md](0008-dcr-config-file-location.md) - Config location and ANKER_HOME
- [0009-dcr-template-engine-selection.md](0009-dcr-template-engine-selection.md) - Template system for reports
- [0010-dcr-build-tool-selection.md](0010-dcr-build-tool-selection.md) - Task runner (superseded by 0012)
- [0011-dcr-code-quality-enforcement.md](0011-dcr-code-quality-enforcement.md) - CI/CD and quality gates
- [0012-dcr-build-system-architecture.md](0012-dcr-build-system-architecture.md) - Just + Dagger build system

## Status Legend

- **Implemented**: Already in code
- **Planned**: Approved, pending implementation
- **Proposed**: Open for discussion

## Private Decision Records

Business strategy and monetization decisions are kept private and are not included in this public repository.
