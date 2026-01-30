# 0002: Source Type Specification

**Date:** 2026-01-28
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Users need to specify source type when adding sources. How should this be expressed in the CLI?

## Options Considered

**Positional argument:** `anker source add git .`
- Good: Explicit (type is required, shows it)
- Good: Concise syntax
- Good: Follows CLI conventions (kubectl, docker)
- Good: Easy to extend (no new flag per type)
- Good: Type-specific flags still work (`--author`)
- Bad: Users must know type upfront

**Flag-based:** `anker source add --git .`
- Good: Familiar pattern for some users
- Bad: Implies optionality (flags = optional)
- Bad: Verbose
- Bad: Hard to extend (each type needs new flag)
- Bad: Type is not optional, shouldn't look optional

**Auto-detection:** `anker source add .` (detects git repo)
- Good: Shortest syntax
- Bad: Violates "explicit over implicit"
- Bad: Ambiguous when multiple types apply
- Bad: Error-prone
- Bad: Inconsistent with philosophy

## Decision

We chose **positional argument**: `anker source add git .`

Why: Makes required parameter explicit, follows CLI conventions, extensible.
