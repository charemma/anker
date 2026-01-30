# 0007: Data Storage Strategy

**Date:** 2026-01-27
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Need to store configuration and sources. How should data be persisted?

## Options Considered

**Text files (TOML/Markdown):**
- Good: Git-friendly (version control, dotfiles sync)
- Good: Cloud sync works (Dropbox, iCloud)
- Good: Human-readable (grep, cat, edit directly)
- Good: No locking issues
- Good: Transparent (see exactly what's stored)
- Good: No cgo dependency
- Good: Simple mental model
- Bad: Slower than database (acceptable for our scale)
- Bad: No transaction safety (single-user CLI, acceptable)

**SQLite:**
- Good: Fast queries
- Good: ACID guarantees
- Bad: Doesn't merge in git
- Bad: Not human-readable (need tools)
- Bad: Sync problems across machines
- Bad: Cgo dependency (complicates cross-compile)
- Bad: Overkill for <1000 sources

**JSON files:**
- Good: Universal format
- Bad: No comments
- Bad: Noisy syntax
- Bad: Less friendly than TOML

**Key-value stores (boltdb, badger):**
- Good: Fast embedded DB
- Bad: Same sync problems as SQLite
- Bad: Not human-readable
- Bad: Don't need the performance

## Decision

We chose **text files** (TOML format).

Why: Git/cloud sync, human-readable, aligns with transparency principle.

```
~/.anker/
  config.toml
  sources.toml
  entries/*.md
```
