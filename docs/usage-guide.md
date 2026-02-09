# Usage Guide

Complete guide to using anker's features.

## Managing Data Sources

### Adding Sources

**Git repositories:**
```bash
anker source add git ~/code/my-project
anker source add git .  # current directory
```

**Markdown notes:**
```bash
# Filter by tags
anker source add markdown ~/Obsidian/Daily --tags work,done

# Filter by headings
anker source add markdown ~/notes --headings "## Work,## Done"
```

**Obsidian vault:**
```bash
anker source add obsidian ~/Obsidian/MyVault
anker source add obsidian ~/Documents/"Second Brain"
```

### Managing Sources

**List all sources:**
```bash
anker source list
```

**Remove sources:**
```bash
anker source remove ~/code/my-project
anker source remove git ~/code/my-project  # if path is ambiguous
```

## Generating Reports

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
anker recap 2025-12-01                    # Single day
anker recap 2025-12-01..2025-12-31        # Date range
```

**Named periods:**
```bash
anker recap "last 7 days"
anker recap "week 52"
anker recap "october 2025"
anker recap "dezember 2025"  # German month names supported
```

### Output Formats

**Simple (default):**
```bash
anker recap today
```
Bullet list of activities.

**Detailed:**
```bash
anker recap today --format detailed
```
Includes timestamps and metadata.

**JSON:**
```bash
anker recap today --format json
```
Structured data for further processing.

**Markdown:**
```bash
anker recap today --format markdown
```
Full context with git diffs, perfect for AI processing or documentation.

## Integration Examples

### Claude CLI

```bash
# Generate standup notes
anker recap yesterday --format markdown | claude -p "Create concise standup notes"

# Weekly report
anker recap thisweek --format markdown | claude -p "Write a professional weekly status report"

# Release notes
anker recap "last 2 weeks" --format markdown | claude -p "Create release notes from these commits"

# Full pipeline: analyze → summarize → render
anker recap thisweek --format markdown | claude -p "Summarize my week" | glow -p
```

### Pretty Terminal Output

```bash
# Render with glow
anker recap thisweek --format markdown | glow -

# Interactive pager
anker recap thisweek --format markdown | glow -p

# Syntax highlighting with bat
anker recap today --format markdown | bat -l markdown
```

### Save and Process

```bash
# Save to file
anker recap "December 2025" --format markdown > monthly-report.md

# View later
glow monthly-report.md

# Process with AI
cat monthly-report.md | claude -p "Create release notes"
```

## Advanced Usage

### Environment Variables

**Custom config directory:**
```bash
export ANKER_HOME=/path/to/custom/config
anker recap today  # uses /path/to/custom/config instead of ~/.anker
```

### Filtering Git Commits

By default, anker filters commits by your git user.email. You can override this:

**Global override:**
```yaml
# ~/.anker/config.yaml
author_email: you@work.com
week_start: monday  # or sunday
```

**Check current git config:**
```bash
git config --global user.email
```

## Troubleshooting

### No entries found

**Check your sources:**
```bash
anker source list
```

**Verify git authorship:**
```bash
git config --global user.email
# Should match commits in your repos
```

**Check time range:**
```bash
# Make sure you have activity in the specified period
anker recap "last 30 days"  # Broader range
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

- Start by tracking your main git repositories
- Add markdown sources for meeting notes or daily logs
- Use `thisweek` on Monday mornings for standup prep
- Pipe to `glow -p` for a nice reading experience
- Use `--format markdown` when working with AI tools
