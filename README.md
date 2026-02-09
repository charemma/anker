# anker

> a fixpoint for your work

**anker** is a local CLI tool that helps you remember what you actually did — without time tracking, productivity metrics, or background agents.

**About this project:** anker is built in collaboration with Claude AI as a demonstration of modern AI-assisted development practices. The goal is to explore how AI can accelerate software development while producing genuinely useful tools.

## The Problem

Work happens across multiple git repositories, scattered notes, meetings, and unplanned tasks. At the end of the day, the hard part isn't doing the work — it's **explaining and remembering what actually happened**.

anker helps you reconstruct your workday after the fact.

## Philosophy

**anker does not try to optimize you.**

It does not tell you how productive you were, how focused you stayed, or how your time was spent.

**You cannot plan everything in advance.** Knowledge work is fundamentally unpredictable — production incidents happen, requirements change mid-sprint, bugs emerge from nowhere, colleagues need urgent help. Sometimes the best solutions come from unexpected detours.

Detailed time-blocking and rigid schedules ignore this reality.

**anker accepts the chaos.**

It exists to help you retain orientation after the fact — to explain your work to yourself or others, not to judge it.

### Core Principles

- **Deferred analysis** — work first, summarize later
- **Explicit over implicit** — nothing is tracked automatically
- **Local & transparent** — all data stays on your machine
- **Text-first** — human-readable storage

## Getting Started

### Installation

```bash
go install github.com/charemma/anker@latest
```

Or build from source:

```bash
git clone https://github.com/charemma/anker
cd anker
just build
```

### Quick Start

```bash
# Track your git repositories (one-time setup)
cd ~/code/my-project
anker source add git .

# Add other data sources
anker source add markdown ~/notes --tags work,done
anker source add obsidian ~/Obsidian/MyVault

# Generate a report
anker recap today
```

## Usage

### Tracking Sources

anker analyzes data from sources you explicitly configure.

**Track a git repository:**
```bash
anker source add git ~/code/my-project
anker source add git .  # current directory
```

**Add markdown notes:**
```bash
anker source add markdown ~/Obsidian/Daily --tags work,done
anker source add markdown ~/notes --headings "## Work,## Done"
```

**Track Obsidian vault:**
```bash
anker source add obsidian ~/Obsidian/MyVault
anker source add obsidian ~/Documents/"Second Brain"
```

**List and remove sources:**
```bash
anker source list
anker source remove ~/code/my-project
anker source remove git ~/code/my-project  # if path is ambiguous
```

### Generating Reports

Create summaries for any time period:

```bash
anker recap today
anker recap yesterday
anker recap thisweek
anker recap lastweek
anker recap 2025-12-01..2025-12-31
anker recap "last 7 days"
anker recap "week 52"
```

**Output Formats:**

```bash
anker recap today --format simple      # default: bullet list
anker recap today --format detailed    # with timestamps and metadata
anker recap today --format json        # structured data
anker recap today --format markdown    # full context with diffs (for AI/docs)
```

**Integration with AI and Tools:**

```bash
# Analyze with Claude CLI
anker recap lastweek --format markdown | claude -p "Summarize my work"

# Generate standup notes
anker recap yesterday --format markdown | claude -p "Create concise standup notes"

# Create weekly report for manager
anker recap thisweek --format markdown | claude -p "Write a professional weekly status report"

# Generate release notes from commits
anker recap "last 2 weeks" --format markdown | claude -p "Create release notes from these commits"

# Pretty terminal rendering
anker recap thisweek --format markdown | glow -
anker recap today --format markdown | bat -l markdown

# Interactive pager with glow
anker recap thisweek --format markdown | glow -p

# Save and process later
anker recap "December 2025" --format markdown > monthly-report.md
cat monthly-report.md | glow -p
```

### Configuration

anker reads your git config for author filtering:

```bash
# By default, reports only show commits by you
git config --global user.email  # used for filtering

# Override in ~/.anker/config.yaml
week_start: monday        # or sunday
author_email: you@work.com
```

**Custom configuration directory:**

```bash
# Set ANKER_HOME to use a different directory
export ANKER_HOME=/path/to/custom/config
anker recap today  # uses /path/to/custom/config instead of ~/.anker
```

## Privacy & Data

**anker has no default sources.**

It does not monitor your system and does not collect data automatically. All sources must be explicitly configured by the user. If a source exists, it is because you asked for it.

**Your data stays local:**
- No telemetry, no analytics, no cloud sync
- All storage in plain text files (`~/.anker/`)
- Human-readable YAML and Markdown

**Data storage:**
```
~/.anker/                  # or $ANKER_HOME if set
  ├── config.yaml          # your preferences
  ├── sources.yaml         # tracked repos and sources
  └── entries/             # (planned) manual work notes
```

## Supported Sources

- **Git repositories** — commits from tracked repos (filtered by author)
- **Markdown files** — notes with tag or heading filters
- **Obsidian vaults** — lists modified/created markdown files by timestamp
- **More planned** — see [TODO.md](TODO.md) for roadmap

## Development

**Quick start:**
```bash
just build              # build to bin/anker
just test               # run all tests
just coverage           # generate coverage report
just check              # run all checks (test + lint + build)
go run . recap today    # run without building
```

**Requirements:**
- Go 1.21+
- [Just](https://just.systems) - command runner
- [Dagger](https://dagger.io) - for containerized builds (optional for local dev)

**Documentation:**
- [Building and Testing](docs/building-and-testing.md) - build system, Dagger, CI/CD
- [Architecture decisions](docs/decisions/) - design rationale and ADRs
- [TODO.md](TODO.md) - feature roadmap and planned improvements

## What anker is NOT

- Not a time tracker
- Not a productivity optimizer
- Not a background daemon
- Not a cloud service
- Not a monitoring tool

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.
