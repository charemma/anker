## anker context

anker is a CLI tool for reconstructing your workday after the fact. It collects activity data from local sources (git commits, markdown notes, Obsidian vaults, Claude Code sessions) and produces recap summaries.

### How anker works

- Sources are explicitly registered: `anker source add git ~/code/project`
- `anker recap today` collects entries from all sources for the given time period
- Output formats: simple (default), detailed, json, markdown, ai
- The `ai` format sends the recap to an LLM for summarization

### Interpreting recap data

When you receive anker recap data (e.g. via `--format ai` or piped input):

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
