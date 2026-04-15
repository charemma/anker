# ikno

> Your workday, reconstructed.

Work leaves traces -- git commits, session logs, timestamped notes. ikno finds them and turns them into something you can read, share, or think with.

[Website](https://ikno.charemma.de) | [Install](#install) | [Quick start](#quick-start)

---

```
$ ikno recap thisweek --style stats

# 2026-04-07 to 2026-04-13 -- 73 activities

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

## Summary

Full rebrand day: renamed anker to ikno, built a prompt template
system with 6 styles, and iterated on logo/naming.
```

---

## Why ikno

Context switches make it hard to remember what you did. You jump between repos, tickets, chats, and coding sessions all day. By the end, half of it is gone.

Planning tools like GTD or Zettelkasten help you decide what comes next. But looking back at what actually happened? That was always a gap you had to fill manually.

ikno closes that gap. Your existing tools become input sources -- Obsidian vault, git repos, Claude Code sessions. ikno reads them all and reconstructs what happened. No logging, no discipline. You work, then you ask.

---

## Install

**Quick install** (installs to `~/.local/bin`):
```bash
curl -fsSL https://ikno.charemma.de/install.sh | sh
```

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

`ikno init` scans your home directory for git repos, Obsidian vaults, and Claude Code sessions. Select what to track, and you're done.

---

## Sources

ikno reconstructs your day from the data you already generate:

- **Git** -- commits from any tracked repo, with diff stats
- **Markdown** -- tagged lines or sections from any `.md` file
- **Obsidian** -- files modified or created in your vault
- **Claude Code** -- AI coding sessions from `~/.claude/projects/`

More sources are planned (Jira, Slack, calendar, browser history). The architecture is extensible -- adding a new source type doesn't require changing core code.

```bash
ikno source add git ~/code/my-project
ikno source add obsidian ~/Documents/Notes
ikno source list
```

---

## Styles

6 built-in report styles. Pick the one that fits:

```bash
ikno recap thisweek --style brief     # standup-ready (5-10 lines)
ikno recap thisweek                   # digest -- full overview (default)
ikno recap thisweek --style status    # progress / blockers / next
ikno recap thisweek --style report    # polished prose for stakeholders
ikno recap thisweek --style retro     # what went well / badly / learnings
ikno recap thisweek --style stats     # category breakdown with ASCII charts
```

## Any language

```bash
ikno recap thisweek --lang english
ikno recap thisweek --lang deutsch
ikno config set ai_language english    # set default
```

## Custom templates

Create your own style as a `.md` file:

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

---

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

## Raw output and pipes

```bash
ikno recap thisweek --raw > this-week.txt
ikno recap thisweek --raw | grep "feat"
ikno recap thisweek --json
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

## License

Apache 2.0 -- see [LICENSE](LICENSE) for details.
