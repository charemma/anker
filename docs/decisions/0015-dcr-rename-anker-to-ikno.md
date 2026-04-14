# 0015: Rename anker to ikno

**Date:** 2026-04-12
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

The name "anker" had no connection to what the tool does and conflicted with existing products (Anker electronics). Needed a distinctive name that reflects the tool's purpose of tracing and reconstructing workdays.

## Options Considered

**Keep "anker":**
- Good: No migration work
- Bad: Name collision with Anker electronics brand
- Bad: No semantic connection to the tool's purpose

**Rename to "ikno":**
- Good: From Greek "ichnos" (trace/imprint), fits the tool's purpose
- Good: Short, easy to type, available as a name
- Good: Personal connection (Greek heritage)
- Bad: Requires migration of module path, config paths, env vars, docs

**Other names:**
- Bad: Nothing else combined brevity, meaning, and availability

## Decision

We chose **ikno**.

Why: The name connects to the tool's core concept (tracing your workday), is short and distinctive, and avoids brand conflicts.

### What changed

- Go module: `charemma/anker` to `charemma/ikno`
- Binary: `anker` to `ikno`
- Config path: `~/.anker/` to `~/.config/ikno/` (XDG-compliant)
- Env var: `ANKER_HOME` to `IKNO_HOME`
- Auto-migration from `~/.anker/` on first run
