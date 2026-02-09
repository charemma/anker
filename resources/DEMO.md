# Creating the Demo GIF

This directory contains everything needed to create a demo GIF showing anker in action.

## Quick Start

```bash
cd resources/

# 1. Install dependencies
just install

# 2. Setup test data
just setup

# 3. Create the GIF
just gif

# Or do everything in one step:
just update
```

## What the demo shows

A live terminal recording showing:

1. **Intro** - What is anker? #AntiProductivity
2. **Add sources** - Git repo + markdown notes
3. **List sources** - See configured sources
4. **Simple recap** - Quick daily summary (with glow)
5. **Full context** - Markdown format with git diffs
6. **AI integration** - Example piping to Claude
7. **Outro** - Install instructions and GitHub link

## Files

- `demo-setup.sh` - Creates test notes with today's timestamp
- `demo.tape` - VHS script that records the terminal session
- `demo.gif` - Generated output (not committed to git)

## Customization

**Edit `demo.tape`** to:
- Change theme: Line 15 (`Set Theme`)
- Adjust timing: `Sleep` values
- Modify size: `Set Width/Height` (Lines 13-14)
- Update commands or text

**Edit `demo-setup.sh`** to:
- Change test note content
- Add more test files

## Available Commands

Run from the `resources/` directory:

- `just install` - Install vhs and glow via brew
- `just setup` - Create test data in /tmp/anker-demo/
- `just gif` - Create demo.gif (~60 seconds)
- `just update` - Run setup + gif in one command

## Notes

- Demo uses isolated environment: `ANKER_HOME=/tmp/anker-demo/.anker`
- Your actual anker config remains untouched
- Test notes are created with today's timestamp
- All commands are executed for real (not mocked)
