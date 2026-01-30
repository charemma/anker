# 0001: Source Discovery Strategy

**Date:** 2026-01-28
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Users work across multiple repos/sources. How should anker know what to track?

## Options Considered

**Explicit registration** (`anker source add git .`):
- Good: User controls exactly what's tracked (privacy)
- Good: Predictable (no surprises)
- Good: Clear mental model ("if tracked, I added it")
- Good: Only analyzes relevant repos (performance)
- Good: No background agents
- Bad: Requires one-time setup per repo
- Bad: New users need explicit onboarding

**Auto-discovery** (scan ~/code for repos):
- Good: No setup needed
- Bad: Violates privacy principle
- Bad: May track personal/work unintentionally
- Bad: Unclear which repos are relevant
- Bad: Implicit behavior

**Config file scanning** (`.ankerrc` lists dirs):
- Good: One config for all
- Bad: Still requires explicit configuration
- Bad: Command-based is more discoverable

## Decision

We chose **explicit registration**.

Why: Privacy, predictability, aligns with "explicit over implicit" principle.
