# 0003: Community Source Extensions

**Date:** 2026-01-28
**Status:** Proposed
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Community users want to track data from sources beyond git/markdown (Jira, Slack, calendar, browser history). Should these be built into core or extensible?

## Options Considered

**Executable plugins** (`anker-source-<type>` binaries):
- Good: Language-agnostic (Python, Go, Rust, shell)
- Good: Isolated (crashes don't affect anker)
- Good: Simple JSON protocol
- Good: No version conflicts
- Good: No dependency hell
- Good: Explicit installation (security)
- Good: Low barrier to entry
- Bad: External process overhead
- Bad: Protocol stability maintenance
- Bad: Discovery complexity

**Library plugins** (Go modules):
- Good: In-process (faster)
- Bad: Locks ecosystem into Go
- Bad: Complex versioning
- Bad: Crashes affect core
- Bad: Dependency conflicts

**All sources in core:**
- Good: Integrated, no plugin system
- Bad: Core becomes massive
- Bad: Can't support all integrations
- Bad: Slow update cycles

**Config-based sources** (YAML connectors):
- Good: No code needed
- Bad: Too limited for auth/pagination
- Bad: Security risk (embedded credentials)

## Decision

We chose **executable plugins**.

Why: Language-agnostic, isolated, simple protocol, extensible.

**Naming:** `anker-source-jira`, `anker-source-slack`
**Protocol:** JSON via stdout
**Discovery:** Auto-discover in `$PATH`

**Built-in** (core): git, markdown
**Community** (plugins): jira, slack, calendar, etc.
