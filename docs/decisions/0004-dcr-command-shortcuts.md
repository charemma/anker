# 0004: Command Shortcuts

**Date:** 2026-01-28
**Status:** Planned
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Users have different workflows and preferences. Some want short commands, others want descriptive ones. Core can't provide shortcuts for everyone without bloating. How can users customize anker to fit their workflow?

## Options Considered

**Git-style aliases** (in config file):
- Good: Power users create shortcuts
- Good: Keeps core clean
- Good: Familiar pattern (git)
- Good: Portable across shells
- Good: Part of anker config (shareable)
- Good: Context-aware (future)
- Good: Discoverable
- Bad: Another thing to learn
- Bad: Potential conflicts (but built-in wins)
- Bad: No recursion

**Built-in convenience commands:**
- Good: No config needed
- Bad: Command bloat (where do we stop?)
- Bad: Not personalizable
- Bad: Harder to maintain

**Shell aliases only:**
- Good: Standard approach
- Bad: Not portable across shells
- Bad: Not part of anker config
- Bad: Can't be context-aware
- Bad: Lost when switching machines

**Plugins for aliases:**
- Good: Very flexible
- Bad: Too complex for text substitution
- Bad: Security concerns
- Bad: Overkill

## Decision

We chose **git-style aliases**.

Why: Personalized shortcuts, keeps core clean, familiar pattern.

```toml
[alias]
track = "source add git"
today = "report today"
```

Usage: `anker track` → `anker source add git .`
