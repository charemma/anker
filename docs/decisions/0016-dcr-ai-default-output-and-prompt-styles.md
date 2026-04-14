# 0016: AI-default Output and Prompt Style System

**Date:** 2026-04-12
**Status:** Implemented
**Participants:** Charalambos Emmanouilidis, Claude.ai
**Supersedes:** Parts of ADR 0005 (report output format), ADR 0009 (template engine selection)

## Problem

The recap command had multiple output formats (`--format simple|detailed|markdown|json|ai`) which made the AI-generated output feel like an optional extra. In practice, the AI recap was the most useful output by far, and the plain text formats were rarely used. The `--format` flag approach also didn't address how to control the AI prompt style (executive summary vs. standup notes vs. retrospective).

## Options Considered

**Keep --format flag with AI as one option:**
- Good: Backward compatible
- Bad: AI is buried as one of many formats
- Bad: No way to control AI prompt style
- Bad: Multiple plain text renderers to maintain

**AI as default, --raw/--json for alternatives:**
- Good: Best output is the default (zero flags for common case)
- Good: Simple flag surface: `--raw` for plain text, `--json` for structured
- Good: Removes 3 unused renderers (simple, detailed, summary)
- Bad: Breaking change for scripts using `--format`

**AI as default with --style for prompt selection:**
- Good: Everything from option B, plus control over AI output style
- Good: Built-in styles cover common use cases
- Good: Custom templates via .md files for full control
- Bad: More flags to document

## Decision

We chose **AI as default with --style and --lang flags**.

Why: The AI recap is the product's core value. Making it the default removes friction. The `--style` flag gives control over AI output without complicating the basic command. `--raw` and `--json` cover the remaining use cases.

### Output modes

- `ikno recap today` -- AI-generated recap (default)
- `ikno recap today --raw` -- plain text, no AI
- `ikno recap today --json` -- structured JSON

### Prompt style system

- `--style <name>` selects a prompt template (default: digest)
- 6 built-in styles: brief, digest, status, report, retro, stats
- `--lang <code>` controls output language (default: en)
- Style resolution: `--style` flag > config `ai_default_style` > "digest"
- Custom templates: place `.md` files in `~/.config/ikno/templates/`
- `--styles` flag lists all available styles with their prompts

### Impact on earlier ADRs

- ADR 0005 (report output format): the open question about template formats is resolved. AI prompts replaced custom output templates for the primary use case.
- ADR 0009 (template engine selection): Lua templates were never implemented. The prompt template system (.md files with Go template variables) covers the customization need with much less complexity.
