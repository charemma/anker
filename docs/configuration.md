# Configuration

ikno stores configuration in `~/.ikno/` (or `$IKNO_HOME` if set).

## Configuration File

Create `~/.ikno/config.yaml` to customize behavior:

```yaml
# Week start day (default: monday)
week_start: monday  # or sunday

# Override git author email for filtering commits
# By default, uses: git config --global user.email
author_email: you@work.com
```

## Custom Configuration Directory

Set `IKNO_HOME` to use a different directory:

```bash
export IKNO_HOME=/path/to/custom/config
ikno recap today  # uses /path/to/custom/config instead of ~/.ikno
```

Add to your shell profile to make it permanent:

```bash
# ~/.zshrc or ~/.bashrc
export IKNO_HOME=/path/to/custom/config
```

## Data Storage

```
~/.ikno/                  # or $IKNO_HOME if set
  ├── config.yaml          # your preferences
  ├── sources.yaml         # tracked repos and sources
  └── entries/             # (planned) manual work notes
```

### sources.yaml Format

This file is managed by `ikno source add/remove` commands, but you can inspect it:

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

ikno reads your git config to filter commits by author:

```bash
# Check your git identity
git config --global user.name
git config --global user.email

# ikno uses user.email to filter commits by default
```

**Author email priority:**
1. `--author` flag when adding a git source
2. `author_email` in `~/.ikno/config.yaml`
3. `git config --global user.email` (automatic fallback)

If none of these are set, ikno will track ALL commits in the repository (with a warning).

## Privacy

All data stays local:
- No telemetry or analytics
- No cloud sync
- Human-readable YAML and Markdown
- No background processes

ikno only reads from locations you explicitly configure with `ikno source add`.
