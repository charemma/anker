# Todo

## Config path: move to ~/.config/anker

Currently all state lives in `~/.anker/` (config.yaml, sources.yaml). Should follow XDG Base Directory spec and use `~/.config/anker/` instead.

Needs:
- Update `internal/paths/GetAnkerHome()` to default to `~/.config/anker`
- Migration: check for `~/.anker/` and move or warn
- Keep `ANKER_HOME` env var override working
- Update docs and CLAUDE.md
