# 0002: Source Type Specification

**Date:** 2026-01-28
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Users need to specify source type when adding sources. How should this be expressed in the CLI?

## Options Considered

**Positional argument:** `ikno source add git .`
- Good: Explicit (type is required, shows it)
- Good: Concise syntax
- Good: Follows CLI conventions (kubectl, docker)
- Good: Easy to extend (no new flag per type)
- Good: Type-specific flags still work (`--author`)
- Bad: Users must know type upfront

**Flag-based:** `ikno source add --git .`
- Good: Familiar pattern for some users
- Bad: Implies optionality (flags = optional)
- Bad: Verbose
- Bad: Hard to extend (each type needs new flag)
- Bad: Type is not optional, shouldn't look optional

**Auto-detection:** `ikno source add .` (detects git repo)
- Good: Shortest syntax
- Bad: Violates "explicit over implicit"
- Bad: Ambiguous when multiple types apply
- Bad: Error-prone
- Bad: Inconsistent with philosophy

## Decision

We chose **positional argument**: `ikno source add git .`

Why: Makes required parameter explicit, follows CLI conventions, extensible.

## Amendment (2026-04-12)

Type is now an optional positional argument. When omitted, type is inferred from detection signals (`.git/`, `.obsidian/`, `.md` files, `.claude/projects/`). Explicit `ikno source add git <path>` syntax remains fully supported and is not deprecated.

New invocation forms:
- `ikno source add` -- detect cwd
- `ikno source add ~/path` -- detect ~/path or scan its children
- `ikno source add ~/path --type git` -- force type on a path

The `--type` flag provides an explicit override when auto-detection is ambiguous or wrong.
