# 0006: CLI Framework Selection

**Date:** 2026-01-27
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Need a CLI framework for subcommands. Standard library, Cobra, urfave/cli, or custom?

## Options Considered

**Cobra:**
- Good: Easy subcommand structure
- Good: Automatic help generation
- Good: Command suggestions on typos
- Good: Shell completion support
- Good: Flag inheritance
- Good: Well-documented, active
- Good: Used by kubectl, gh, hugo
- Good: Small API surface
- Bad: Adds dependency (but stable)
- Bad: Shapes command structure (but fits)

**Standard library flag:**
- Good: No dependencies
- Bad: Tedious subcommand implementation
- Bad: No automatic help
- Bad: Would reinvent Cobra

**urfave/cli:**
- Good: Another option
- Bad: Less intuitive API
- Bad: Smaller community
- Bad: No strong advantage

**Custom parser:**
- Good: Total control
- Bad: Time-consuming
- Bad: Easy to get wrong
- Bad: Maintenance burden

## Decision

We chose **Cobra**.

Why: Easy subcommands, automatic help, proven at scale (kubectl, gh, hugo).
