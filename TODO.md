# TODO

Feature ideas and improvements for future releases.

## Planned Features

### Command Alias System
User-defined shortcuts for frequent operations.

```toml
# ~/.anker/config.toml
[alias]
track = "source add git"
today = "report today"
```

**Priority:** High - improves daily UX
See [docs/decisions/0004-dcr-command-shortcuts.md](docs/decisions/0004-dcr-command-shortcuts.md)

### Template-Based Report Output
Customizable report formatting for freelancers and professionals.

```bash
anker report thisweek --format markdown
anker report thisweek --template client-timesheet
```

Built-in formats: text, markdown, html, json
User templates: `~/.anker/templates/*.tmpl`

**Priority:** High - differentiating feature
See [docs/decisions/0005-dcr-report-output-format.md](docs/decisions/0005-dcr-report-output-format.md)

### anker note Command
Implement the note command for one-off work entries.

**Features:**
- Add notes with explicit timestamps (for retroactive entries)
- Default to current time if not specified
- Support same timespec format as report command

**Examples:**
```bash
# Current time (default)
anker note "Customer call about feature X"

# Specific timestamp
anker note "Fixed production bug" --at "2025-12-15 14:30"
anker note "Team meeting" --at yesterday
anker note "Code review" --at "2 hours ago"

# Date only (uses start of day)
anker note "Started new project" --date 2025-12-01
```

**Storage:**
```
~/.anker/entries/2025/12/2025-12-15.md
---
14:30 - Customer call about feature X
16:45 - Fixed production bug
```

**Use case:**
- Reconstructing forgotten activities
- Adding context to commits
- Meetings, calls, research that doesn't result in commits

### Browser History Source Provider
Track work-related browsing activity from browser history databases.

**Implementation:**
- Read SQLite history from Chrome, Firefox, Safari
- Filter by domain whitelist (github.com, confluence, jira, etc.)
- Extract: URL, page title, visit time
- Use case: Track PR reviews, documentation reading, research

**Example:**
```bash
anker source add browser chrome --domains github.com,jira.atlassian.net
```

**Privacy considerations:**
- Strict domain filtering
- Optional time-based filtering (work hours only)
- No sensitive URLs (password managers, private sites)

### Per-Repository Configuration
Support `.anker` config file in individual repositories.

**Use case:**
- Different author emails per project/client
- Per-repo source settings
- Override global defaults

**Format:**
```yaml
# .anker in repo root
author_email: client@company.com
include_branches: [main, develop]
exclude_patterns: [WIP, fixup]
```

### Calendar Integration
Track meetings and events as work activities.

**Sources:**
- Google Calendar API
- iCal files
- Outlook calendar

**Example:**
```bash
anker source add calendar google --calendar work@company.com
```

### Issue Tracker Integration
Pull ticket activity from issue trackers.

**Providers:**
- Jira
- Linear
- GitHub Issues
- GitLab Issues

**Track:**
- Issues created/closed
- Comments
- Status changes
- Time spent

### Enhanced Reporting

**Weekly summaries:**
```bash
anker report thisweek --format markdown > weekly-report.md
```

**Grouping options:**
- By repository
- By source type
- By day
- By project/tag

**Output formats:**
- Plain text (current)
- Markdown
- HTML
- JSON (for further processing)

### AI-Powered Summaries
Generate natural language summaries of work done.

**Use case:**
- Standup preparation
- Weekly reports for management
- Sprint retrospectives

**Requires:**
- Optional OpenAI/Claude API integration
- Local LLM support (ollama)

### Recursive Repository Discovery
Scan directories for git repositories automatically.

**Example:**
```bash
anker scan ~/code --recursive
```

**Features:**
- Detect all git repos in directory tree
- Batch track with single command
- Respect .gitignore patterns

### Source Management Improvements

**Features:**
```bash
anker source add git . --author custom@email.com
anker source add git . --no-author  # track all commits
anker source edit git /path/to/repo --author new@email.com
```

### Time Tracking Enhancement
Better time range specifications.

**Additional formats:**
- "this month" / "last month"
- "Q4 2025" (quarters)
- "2025" (full year)
- "last 2 weeks"

### Configuration Management
```bash
anker config set author_email user@company.com
anker config set week_start sunday
anker config get author_email
anker config list
```

## Technical Improvements

### Migrate to TOML Configuration
Replace YAML with TOML for better human-readability.

**Changes:**
- `config.yaml` → `config.toml`
- `sources.yaml` → `sources.toml`
- Backward compatibility during migration

**Priority:** Medium - improves UX
See [docs/decisions/0007-dcr-data-storage-strategy.md](docs/decisions/0007-dcr-data-storage-strategy.md)

## Technical Debt

### Test Coverage Roadmap

**Current: 49%**

Incremental coverage improvement plan:

1. **Target: 60%** (Priority: High)
   - Add tests for cmd/source.go (loadConfig, add/list/remove commands)
   - Add tests for internal/config package (Load, Save, GetTimerangeConfig)
   - Test git config helpers (GetAuthorEmail, GetAuthorName)

2. **Target: 70%** (Priority: High)
   - Add integration tests for recap command (all output formats)
   - Test markdown/obsidian Type() and Location() methods
   - Test storage AddSource/GetSources/RemoveSource methods

3. **Target: 75%** (Priority: Medium)
   - Test git diff functionality (GetDiff, EnrichWithDiffs)
   - Test timerange locale system (ParseMonth, RegisterMonthNames)
   - Edge cases for existing tests

4. **Target: 80%** (Priority: Medium)
   - Full integration tests for all commands
   - Error path testing
   - Test with real Obsidian vault structures

5. **Target: 85%** (Priority: Low)
   - Complete CLI integration tests
   - Test all error scenarios
   - Performance regression tests

6. **Target: 90%+ (Stretch goal)**
   - Exhaustive edge case coverage
   - Fuzz testing for parsers
   - Property-based testing

### Documentation
- User guide
- Source provider development guide
- Configuration reference

### Performance
- Cache git log results
- Parallel source processing
- Lazy loading for large histories

## Nice to Have

### Shell Completion
Generate shell completions for bash/zsh/fish.

### TUI (Terminal UI)
Interactive terminal interface for:
- Browsing tracked sources
- Quick date selection
- Real-time preview

### Plugin System for Community Sources

Allow community-contributed source providers (Jira, Slack, Calendar, etc.) while keeping built-in sources (git, markdown) in core.

See [docs/decisions/0003-dcr-community-source-extensions.md](docs/decisions/0003-dcr-community-source-extensions.md) for design.

### Export Formats
- PDF reports
- CSV export for spreadsheets
- Integration with time tracking tools (Toggl, Harvest)

### Activity Heatmap
Visual representation of work activity over time.
