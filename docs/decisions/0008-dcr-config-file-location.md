# 0008: Config File Location

**Date:** 2026-01-28
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Where should ikno read its configuration? Need a predictable location that's also flexible for testing and multi-environment setups.

## Options Considered

**`~/.ikno/config.toml` with `IKNO_HOME` override:**
- Good: Predictable default location
- Good: Environment variable for flexibility
- Good: Follows XDG-like pattern
- Good: Easy for tests (set IKNO_HOME to temp dir)
- Good: Multi-environment support
- Good: Like git's GIT_CONFIG_GLOBAL pattern
- Bad: Need to document env var

**Fixed `~/.ikno/config.toml` only:**
- Good: Simple, predictable
- Bad: No test isolation
- Bad: No multi-environment support
- Bad: Hard to override

**XDG Base Directory (`~/.config/ikno/`):**
- Good: Standards-compliant (XDG)
- Bad: Less discoverable than ~/.ikno
- Bad: More directories to navigate
- Bad: Not all users familiar with XDG

**Per-directory `.ikno` config:**
- Good: Project-specific settings
- Bad: Config scattered everywhere
- Bad: Harder to manage
- Bad: Not a global config solution

## Decision

We chose **`~/.ikno/config.toml`** with **`IKNO_HOME`** environment variable override.

Why: Predictable default, flexible when needed, follows git pattern.

**Default:** `~/.ikno/config.toml`
**Override:** `export IKNO_HOME=/custom/path` → `/custom/path/config.toml`
**Structure:**
```
~/.ikno/           # or $IKNO_HOME
  config.toml       # main config
  sources.toml      # tracked sources
  templates/        # user templates
  entries/          # manual notes
```

Like git's `~/.gitconfig`, but with override capability.
