# Configuration

anker stores configuration in `~/.anker/` (or `$ANKER_HOME` if set).

## Configuration File

Create `~/.anker/config.yaml` to customize behavior:

```yaml
# Week start day (default: monday)
week_start: monday  # or sunday

# Override git author email for filtering commits
# By default, uses: git config --global user.email
author_email: you@work.com
```

## Custom Configuration Directory

Set `ANKER_HOME` to use a different directory:

```bash
export ANKER_HOME=/path/to/custom/config
anker recap today  # uses /path/to/custom/config instead of ~/.anker
```

Add to your shell profile to make it permanent:

```bash
# ~/.zshrc or ~/.bashrc
export ANKER_HOME=/path/to/custom/config
```

## Data Storage

```
~/.anker/                  # or $ANKER_HOME if set
  ├── config.yaml          # your preferences
  ├── sources.yaml         # tracked repos and sources
  └── entries/             # (planned) manual work notes
```

### sources.yaml Format

This file is managed by `anker source add/remove` commands, but you can inspect it:

```yaml
sources:
  - type: git
    path: /path/to/repo
    added: 2026-01-27T12:14:01+02:00

  - type: markdown
    path: /path/to/notes
    added: 2026-01-27T15:00:00+02:00
    metadata:
      tags: work,done
      headings: "## Work,## Done"

  - type: obsidian
    path: /Users/you/Obsidian/Second Brain
    added: 2026-01-29T10:00:00+02:00
```

## Git Configuration

anker reads your git config to filter commits by author:

```bash
# Check your git identity
git config --global user.name
git config --global user.email

# anker uses user.email to filter commits
# Override in ~/.anker/config.yaml if needed
```

## Privacy

All data stays local:
- No telemetry or analytics
- No cloud sync
- Human-readable YAML and Markdown
- No background processes

anker only reads from locations you explicitly configure with `anker source add`.
