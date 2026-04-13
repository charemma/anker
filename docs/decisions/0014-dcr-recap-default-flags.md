# 0014: Recap Default Flags and Output File

**Date:** 2026-03-01
**Status:** Proposed
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Running `ikno recap today --format ai --output ~/path/to/vault/recap.md` every time is tedious. Users need a way to set persistent defaults for recap flags, including an output path with date-based filenames.

## Options Considered

**Option A:** Nested config section
```yaml
recap:
  format: ai
  output: ~/Documents/Notes/.../Recaps/
```
- Good: Clean structure, scales to other commands
- Bad: Different syntax than CLI, user has to learn field names

**Option B:** Flag string in config (like FZF_DEFAULT_OPTS)
```yaml
recap_default_flags: "--format ai --output ~/Documents/Notes/.../Recaps/{date}_recap.md"
```
- Good: Same syntax as CLI, nothing new to learn
- Good: `--help` documents everything, config just reuses it
- Bad: Needs shlex-style parsing for quoted paths with spaces

**Option C:** Individual flat fields (recap_format, recap_output)
- Good: Simple to implement
- Bad: Gets messy with more fields, naming inconsistent with existing config

## Decision

We chose **Option B**.

Why: CLI syntax consistency is the strongest argument. Users already know the flags from `--help`. A shlex library handles the quoting edge case. FZF has proven this pattern works well.

## Implementation Notes

- Add `--output` / `-o` flag to recap command
- Support placeholders in output path, using Go's time format tokens:
  - `{YYYY}` -- year (2026)
  - `{MM}` -- zero-padded month (03)
  - `{MMMM}` -- full month name (March)
  - `{DD}` -- zero-padded day (01)
  - `{dddd}` -- full weekday name (Sunday)
  - `{date}` -- shortcut for `YYYY-MM-DD`, resolves to range `2026-02-23_2026-03-01` for multi-day periods
  - `{timespec}` -- raw input (today, lastweek, etc.)
- Obsidian Daily Notes compatible path example:
  `--output "~/Documents/Notes/Journal/{YYYY}/{MM-MMMM}/{YYYY-MM-DD-dddd}_recap.md"`
  resolves to `Journal/2026/03-March/2026-03-01-Sunday_recap.md`
- CLI flags override `recap_default_flags`
- Shlex-style splitting for the flag string (handle quoted paths with spaces)
- Obsidian formatting is handled via `ai_prompt`, not via flags
- `--split` flag: when used with a multi-day timespec (e.g. `lastweek`), generates one file per day instead of a single report. Each file uses the day's date as filename. Without `--split`, everything goes into one report.
