# anker

**anker** is a local, text-first CLI tool that helps you remember what you actually did —  
without time tracking, productivity metrics, or background agents.

> anker — a fixpoint for your work

## What problem does anker solve?

Work happens across:
- multiple git repositories
- investigations, meetings, admin work
- unplanned, ad-hoc tasks

At the end of the day, the hard part is not *doing* the work —  
it's **explaining and remembering what actually happened**.

anker helps you reconstruct your workday **after the fact**.

No timers.  
No tracking.  
No cloud.

## Core ideas

- **Deferred analysis**: work first, summarize later
- **Explicit over implicit**: nothing is tracked automatically
- **Local & transparent**: all data stays on your machine
- **Text-first**: human-readable storage

## Basic workflow

```bash
cd my/repo
anker track

# later
anker today
```

For work outside repositories:

```bash
anker note "Invoice written"
anker note "Customer call"
```

## Commands

| Command | Status | Description |
|------|--------|-------------|
| `anker track` | ✓ | Mark current repository for later analysis |
| `anker source add` | ✓ | Add data sources (markdown notes, etc.) |
| `anker source list` | ✓ | List all configured sources |
| `anker today` | planned | Generate a summary for today |
| `anker note` | planned | Add a one-off work note |

## Non-goals

- No time tracking
- No productivity scoring
- No background daemon
- No IDE plugins

## Storage

```text
~/.anker/
  ├── sources.yaml       - tracked git repos, markdown dirs, etc.
  ├── entries/           - (planned) work notes
  └── 2026/01/           - (planned) generated summaries
```

anker uses an extensible source system. Currently supported:
- Git repositories (via `anker track`)
- Markdown files (via `anker source add markdown`)
- More sources planned: calendar, Jira, GitHub activity

## Philosophy

anker is not a productivity tool.

It does not tell you how productive you were.  
It helps you **explain your work** — to yourself or others.

## Building

```bash
task build          # builds to bin/anker
task test           # run tests
go install .        # install to $GOPATH/bin
```

Requires Go 1.21+ and [Task](https://taskfile.dev) for build automation.

---

Inspired by calm, explicit CLI tools like `git`, `fzf`, and `chezmoi`.

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
