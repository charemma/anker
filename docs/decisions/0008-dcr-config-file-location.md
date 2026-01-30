# 0008: Config File Location

**Date:** 2026-01-28
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Where should anker read its configuration? Need a predictable location that's also flexible for testing and multi-environment setups.

## Options Considered

**`~/.anker/config.toml` with `ANKER_HOME` override:**
- Good: Predictable default location
- Good: Environment variable for flexibility
- Good: Follows XDG-like pattern
- Good: Easy for tests (set ANKER_HOME to temp dir)
- Good: Multi-environment support
- Good: Like git's GIT_CONFIG_GLOBAL pattern
- Bad: Need to document env var

**Fixed `~/.anker/config.toml` only:**
- Good: Simple, predictable
- Bad: No test isolation
- Bad: No multi-environment support
- Bad: Hard to override

**XDG Base Directory (`~/.config/anker/`):**
- Good: Standards-compliant (XDG)
- Bad: Less discoverable than ~/.anker
- Bad: More directories to navigate
- Bad: Not all users familiar with XDG

**Per-directory `.anker` config:**
- Good: Project-specific settings
- Bad: Config scattered everywhere
- Bad: Harder to manage
- Bad: Not a global config solution

## Decision

We chose **`~/.anker/config.toml`** with **`ANKER_HOME`** environment variable override.

Why: Predictable default, flexible when needed, follows git pattern.

**Default:** `~/.anker/config.toml`
**Override:** `export ANKER_HOME=/custom/path` → `/custom/path/config.toml`
**Structure:**
```
~/.anker/           # or $ANKER_HOME
  config.toml       # main config
  sources.toml      # tracked sources
  templates/        # user templates
  entries/          # manual notes
```

Like git's `~/.gitconfig`, but with override capability.
