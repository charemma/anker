# 0005: Report Output Format

**Date:** 2026-01-28
**Status:** Proposed (Under Consideration)
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Different users need different report formats:
- **Freelancers:** Copy-paste into timesheets, invoices, client systems
- **Employees:** Weekly summaries for standups, manager updates
- **Open source maintainers:** Contribution summaries for sponsors
- **Academics:** Research activity logs

Current hardcoded text output doesn't meet these needs. Users have to manually reformat output, which takes time and is error-prone.

**Goal:** Make report generation so good that users can't work without it. Save 30+ minutes per week for freelancers.

## Options Considered

**Go templates (stdlib text/template):**
- Good: No external dependencies
- Good: In standard library
- Good: Type-safe with Go structs
- Good: Good performance
- Bad: Syntax might be unfamiliar ({{ .Field }})
- Bad: Less features than dedicated template engines

**Jinja2-style templates (e.g., pongo2):**
- Good: Very popular (Python users know it)
- Good: Rich feature set (filters, tests, etc.)
- Good: Familiar syntax ({% %}, {{ }})
- Good: Good documentation and examples
- Bad: Need Go implementation library (pongo2)
- Bad: External dependency
- Bad: Might be overkill for simple use cases

**Mustache/Handlebars templates:**
- Good: Logic-less (very simple)
- Good: Multi-language (familiar to many developers)
- Good: Clean, minimal syntax
- Bad: Too limited for complex formatting needs
- Bad: External dependency
- Bad: May need extensions for date/time formatting

**Fixed formats only (text, markdown, JSON, HTML):**
- Good: Simple to implement
- Good: No user learning curve
- Good: No template syntax errors
- Good: Works immediately
- Bad: Can't meet everyone's exact needs
- Bad: Not a differentiating feature
- Bad: Freelancers need client-specific formats
- Bad: Would need many built-in variants

**Embedded scripting (Lua, JavaScript via goja):**
- Good: Most flexible (full programming language)
- Good: Can handle any formatting logic
- Bad: Way too complex for formatting task
- Bad: Security concerns (arbitrary code execution)
- Bad: Overkill for text output
- Bad: Need runtime interpreter

**Simple string interpolation:**
- Good: Very simple for users
- Good: No syntax to learn
- Bad: Can't handle loops or conditionals
- Bad: Limited to single-line patterns
- Bad: Can't format dates/times properly

## Decision

**Under consideration.** Need to evaluate which templating approach best fits user needs.

**Leading candidates:**
1. **Go stdlib templates** - No dependencies, good enough
2. **Jinja2-style (pongo2)** - Familiar to many users
3. **Start with fixed formats** - Simple, validate demand first

**Potential usage:**
```bash
anker report thisweek --format markdown       # built-in
anker report thisweek --template client-timesheet  # user custom
```

**Open questions:**
- Which template syntax will users find easiest?
- Should we start simple (fixed formats) and add templating later?
- What built-in formats should we provide initially?
- How do we handle template errors clearly?
- Where to store templates? (`~/.anker/templates/`?)
- Do we need template validation on add?

This needs user research and validation before deciding on implementation.
