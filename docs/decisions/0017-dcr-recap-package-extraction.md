# 0017: Recap Package Extraction

**Date:** 2026-04-12
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

The `cmd/recap.go` file had grown to 472 lines, mixing Cobra command wiring with business logic (entry collection, grouping, rendering). This made the code hard to test, hard to extend with new output formats, and violated separation of concerns.

## Options Considered

**Keep everything in cmd/recap.go:**
- Good: No refactoring work
- Bad: Untestable business logic (coupled to Cobra)
- Bad: Single file doing collection, grouping, and rendering
- Bad: Adding a new renderer means touching command code

**Extract to internal/recap/ package:**
- Good: Business logic testable independently
- Good: Clear separation: cmd/ handles CLI, internal/recap/ handles logic
- Good: Each renderer in its own file
- Good: cmd/recap.go shrinks to ~80 lines (just command wiring)
- Bad: More files to navigate

**Extract to pkg/ (public API):**
- Good: Reusable by other tools
- Bad: Premature -- no external consumers exist
- Bad: Public API is a maintenance commitment

## Decision

We chose **internal/recap/ package**.

Why: Clean separation without the commitment of a public API. The cmd layer stays thin (flag parsing, calling into internal/recap/), and each concern gets its own file.

### Package structure

```
internal/recap/
  collect.go           -- BuildRecap: collects entries from all sources
  group.go             -- GroupByRepo, SourceLabel helpers
  result.go            -- RecapResult type definition
  render_markdown.go   -- markdown/plain text renderer (--raw)
  render_json.go       -- JSON renderer (--json)
```

The old `render_simple.go`, `render_detailed.go`, and `render_summary.go` were removed as part of the AI-default output change (ADR 0016).
