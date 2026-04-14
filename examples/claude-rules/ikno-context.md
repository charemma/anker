## ikno context

ikno is a CLI tool for reconstructing your workday after the fact. It collects activity data from local sources (git commits, markdown notes, Obsidian vaults, Claude Code sessions) and produces recap summaries.

### How ikno works

- Sources are explicitly registered: `ikno source add git ~/code/project`
- `ikno recap thisweek` collects entries from all sources and generates an AI summary
- Output modes: AI summary (default), `--raw` for plain text, `--json` for structured data
- `--style` selects the AI prompt style (brief, digest, status, report, retro, stats)
- `--lang` controls the output language

### Interpreting recap data

When you receive ikno recap data (e.g. via piped `--raw` input):

- Entries are grouped by source location (repository path, vault path, etc.)
- Each entry has a timestamp, content (commit message, note text, etc.), and metadata
- Git entries include author, commit hash, and optionally full diffs
- Claude entries include session slugs
- Entries are sorted newest-first

### What makes a good summary

- Group by topic/theme, not chronologically
- Skip trivial changes (typo fixes, formatting) unless part of a larger effort
- Highlight decisions, trade-offs, and open threads when present in the data
- Don't invent information that isn't in the recap data
- Match the language the user asks for
