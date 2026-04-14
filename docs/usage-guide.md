# Usage Guide

Complete guide to using ikno's features.

## Managing Data Sources

### Adding Sources

**Git repositories:**
```bash
ikno source add git ~/code/my-project
ikno source add git .  # current directory

# Filter by author (default: git config user.email)
ikno source add git . --author you@work.com
ikno source add git . --author foo@work.com --author bar@personal.com
```

By default, ikno uses your `git config --global user.email` to filter commits. You can override this with `--author` or set `author_email` in `~/.config/ikno/config.yaml`.

**Markdown notes:**
```bash
# Filter by tags
ikno source add markdown ~/Obsidian/Daily --tags work,done

# Filter by headings
ikno source add markdown ~/notes --headings "## Work,## Done"
```

**Obsidian vault:**
```bash
ikno source add obsidian ~/Obsidian/MyVault
ikno source add obsidian ~/Documents/"Second Brain"
```

### Interactive Setup

```bash
ikno init
```

The init wizard scans your system for git repos, Claude Code sessions, and Obsidian vaults, then lets you select which to register.

### Managing Sources

**List all sources:**
```bash
ikno source list
```

**Remove sources:**
```bash
ikno source remove ~/code/my-project
ikno source remove git ~/code/my-project  # if path is ambiguous
```

## Generating Recaps

### Time Specifications

**Relative:**
- `today` - Today's work
- `yesterday` - Yesterday's work
- `thisweek` - Current week (Monday-Sunday)
- `lastweek` - Previous week
- `thismonth` - Current month
- `lastmonth` - Previous month

**Specific dates:**
```bash
ikno recap 2025-12-01                    # Single day
ikno recap 2025-12-01..2025-12-31        # Date range
```

**Named periods:**
```bash
ikno recap "last 7 days"
ikno recap "week 52"
ikno recap "october 2025"
ikno recap "dezember 2025"  # German month names supported
```

### Output Modes

**AI summary (default):**
```bash
ikno recap today
ikno recap thisweek --style brief
ikno recap yesterday --style status --lang english
```

AI-generated summary using your configured backend. Styles: brief, digest (default), status, report, retro, stats.

**Raw activity log:**
```bash
ikno recap today --raw
```

Plain text list of all activities with timestamps.

**JSON:**
```bash
ikno recap today --json
```

Structured data for further processing.

## AI Configuration

### Styles

Built-in prompt styles control how the AI formats your recap:

- `brief` - Quick 3-5 bullet summary
- `digest` - Grouped by theme with highlights (default)
- `status` - Standup-ready format
- `report` - Professional weekly report
- `retro` - Retrospective with lessons learned
- `stats` - Work statistics with ASCII charts

Select with `--style` flag or set a default in config:

```yaml
# ~/.config/ikno/config.yaml
ai_default_style: digest
ai_language: deutsch
```

### Backends

**CLI backend (recommended for Claude Pro/Max subscribers):**
```yaml
ai_backend: cli
ai_cli_command: claude -p     # default
```

No API costs -- uses your existing subscription.

**API backend:**
```yaml
ai_backend: api
ai_base_url: https://api.anthropic.com/v1/
ai_model: claude-sonnet-4-20250514
ai_api_key: sk-...
```

Supports any OpenAI-compatible endpoint (Anthropic, OpenAI, ollama, vllm).

### Custom Prompts

Override the built-in prompt:
```yaml
ai_prompt: |
  Summarize my workday. Group by topic, not chronologically.
  Write in German. Use ## headings and bullet points.
```

Custom prompt templates as `.md` files are also supported.

## Integration Examples

### Piping to External Tools

```bash
# Pipe raw output to any LLM CLI
ikno recap today --raw | claude -p "Create standup notes"
ikno recap thisweek --raw | llm "Write a weekly report"

# Render with glow
ikno recap today | glow -

# Save to file
ikno recap "December 2025" > monthly-report.md
```

### Obsidian Integration

Set `ai_prompt` in config to produce Obsidian-friendly output with wikilinks and tags:

```yaml
ai_prompt: |
  Summarize my workday. The output will be stored in my Obsidian vault.
  Group by topic, use ## headings, bullet points.
  Add tags: #recap #ikno
  Link mentioned projects as [[Wikilinks]]
```

## Advanced Usage

### Environment Variables

**Custom config directory:**
```bash
export IKNO_HOME=/path/to/custom/config
ikno recap today  # uses /path/to/custom/config instead of ~/.config/ikno
```

### Configuration

```bash
ikno config set ai_default_style brief
ikno config set ai_language english
ikno config set ai_backend cli
ikno config get ai_default_style
ikno config list
```

## Troubleshooting

### No entries found

**Check your sources:**
```bash
ikno source list
```

**Verify git authorship:**
```bash
git config --global user.email
# Should match commits in your repos
```

**Check time range:**
```bash
# Make sure you have activity in the specified period
ikno recap "last 30 days"  # Broader range
```

### Source validation errors

**Git repository:**
- Make sure the path contains a `.git` directory
- Repository must have commits

**Obsidian vault:**
- Directory must contain `.obsidian` folder
- Vault must be initialized

**Markdown source:**
- Directory must exist and contain `.md` files
- Check tag/heading filters are correct

## Tips

- Run `ikno init` to get started quickly
- Use `thisweek` on Monday mornings for standup prep
- Try different `--style` options to find your preferred format
- Set `ai_language` in config to get recaps in your preferred language
