# ikno

> Your workday, reconstructed.

From git, notes, and AI sessions -- in one command.

---

```
$ ikno recap today --style stats

# 2026-04-13 -- 73 activities

## Categories

Coding               ███████░░░░░░░░░░░░░  37%  (27)
  ikno rename, prompt templates, merge PR

AI/Prompt Design     █████░░░░░░░░░░░░░░░  26%  (19)
  Style testing, template iteration

Branding             ████░░░░░░░░░░░░░░░░  21%  (15)
  Logo design, naming research (ikno)

Documentation        ██░░░░░░░░░░░░░░░░░░  11%  (8)
  README rewrite, ADR updates, vault notes

DevOps               █░░░░░░░░░░░░░░░░░░░   5%  (4)
  CI auto-tag, nix flake fixes

## Work Types

Building             ███████████░░░░░░░░░  55%
Thinking             ███████░░░░░░░░░░░░░  33%
Organizing           ██░░░░░░░░░░░░░░░░░░  12%

## Summary

Full rebrand day: renamed anker to ikno, built a prompt template
system with 6 styles, and iterated on logo/naming.
```

---

## More styles

**Standup (brief):**
```
$ ikno recap today --style brief

Done
- Renamed anker to ikno (full rebrand)
- Added prompt template system with --style/--lang flags
- Implemented stats style with ASCII charts
- Set up auto-tagging CI

Next
- Logo finalization for ikno
- Custom template docs/polish
```

**Retrospective:**
```
$ ikno recap thisweek --style retro

### What went well
- The anker-to-ikno rename landed cleanly across module, CLI, config, docs.
- Prompt template system shipped in one day: 6 built-in styles, custom templates, --lang flag.

### What didn't go well
- Stats output formatting needed three consecutive fix commits.
- Logo design went nowhere after three sessions.

### Learnings
- Test AI output against the terminal renderer before committing.
- Parallel-agent pattern for naming worked well -- worth repeating.
```

---

## Install

**Nix:**
```bash
nix run github:charemma/ikno
```

**Go:**
```bash
go install github.com/charemma/ikno@latest
```

---

## Quick start

```bash
ikno init
ikno recap thisweek
```

`ikno init` detects git repos in your current directory, Claude Code sessions in `~/.claude/projects/`, and an Obsidian vault if one is configured. You can add more sources explicitly at any time.

---

## Sources

ikno reconstructs your day from the data you already generate:

- **Git** -- commits from any tracked repo, with diff stats
- **Markdown** -- tagged lines or sections from any `.md` file
- **Obsidian** -- files modified or created in your vault
- **Claude Code** -- AI coding sessions from `~/.claude/projects/`

Add sources explicitly when auto-detection is not enough:

```bash
ikno source add git ~/code/my-project
ikno source add obsidian ~/Documents/Notes
ikno source list
```

---

## Styles

ikno ships with 6 built-in report styles. Pick the one that fits your audience:

```bash
ikno recap today --style brief       # Done / Next / Blockers (5-10 lines)
ikno recap thisweek                  # digest -- full overview (default)
ikno recap thisweek --style status   # Progress / Blockers / Next
ikno recap thisweek --style report   # Polished prose, deliveries first
ikno recap thisweek --style retro    # What went well / badly / learnings
ikno recap today --style stats       # Category breakdown with ASCII charts
```

See what each style does: `ikno recap --styles`

## Any language

Every report -- headings, bullets, everything -- is written in the language you choose:

```bash
ikno recap thisweek --lang english
ikno recap thisweek --lang deutsch
ikno recap thisweek --lang greek
```

Set a default so you don't have to type it every time:

```bash
ikno config set ai_language english
```

## Custom templates

Need a report for a specific client, a weekly team email, or a format your manager prefers? Create your own style as a `.md` file:

```bash
mkdir -p ~/.config/ikno/templates

cat > ~/.config/ikno/templates/client-acme.md << 'EOF'
---
description: Weekly status report for Acme Corp
---

Write a professional status report for a client.
Focus on deliverables and milestones. No internal jargon.
Write EVERYTHING in {language}.
Max 15 lines.
EOF

ikno recap thisweek --style client-acme
```

List all available styles (built-in + custom):

```bash
ikno recap --styles
```

## AI backend

ikno uses your own AI setup -- no account with us, no data leaves your machine unless you choose:

- **CLI tool** (default): `claude -p`, or any tool that reads from stdin
- **API endpoint**: OpenAI, Anthropic, Ollama, or any compatible API
- **Local model**: Ollama running locally -- fully offline

```bash
ikno config set ai_backend cli
ikno config set ai_cli_command "claude -p"
```

---

## Time ranges

```bash
ikno recap today
ikno recap yesterday
ikno recap thisweek
ikno recap lastweek
ikno recap "april 2025"
ikno recap 2025-04-01..2025-04-30
```

---

## Plain output and pipes

When piped or redirected, ikno produces clean plain text without ANSI codes. Force it explicitly with `--plain`:

```bash
ikno recap thisweek --plain > this-week.txt
ikno recap thisweek --plain | grep "git/myrepo"
```

---

## Configuration

```yaml
# ~/.config/ikno/config.yaml
week_start: monday
author_email: you@example.com
ai_default_style: digest
ai_language: en
```

---

## Philosophy

Most productivity tools ask you to track everything up front -- every minute, every switch, every task. ikno does not.

You work normally. At the end of the day (or week), you run one command and get a readable account of what actually happened. No timers, no categories, no logging discipline required.

The insight behind ikno: the traces are already there. Git commits, session histories, timestamped notes -- your work leaves marks. ikno reads those marks and turns them into something you can read, share, or think with.

No background agents. No cloud sync. Everything stays in `~/.config/ikno/`.

---

## License

Apache 2.0 -- see [LICENSE](LICENSE) for details.
