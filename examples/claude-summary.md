# Generating work summaries with AI

anker collects raw activity data. Use `--format ai` to get LLM-generated summaries directly, or pipe output into any LLM CLI.

## Built-in AI summary

The simplest way -- anker calls the LLM API directly and streams the response:

```bash
anker recap today --format ai
anker recap thisweek --format ai
anker recap yesterday --format ai --prompt "Write a standup update."
```

This requires an API key and endpoint configured in `~/.anker/config.yaml` or via environment variables. Run `anker config` to set it up. See the config file comments for supported providers (Anthropic, OpenAI, ollama, etc.).

API key resolution: `--api-key` flag > `AI_API_KEY` env var > config file.

## CLI backend

If you have a Claude Pro/Max subscription or another LLM CLI tool, you can use `--format ai` without API credits by switching to the CLI backend:

```yaml
# ~/.anker/config.yaml
ai_backend: cli
ai_cli_command: claude -p     # default
```

This pipes your recap data via stdin to the CLI tool with the prompt as the last argument. The tool's output goes straight to your terminal. `--prompt` works as usual, `--api-key` is ignored.

Other CLI tools work too:

```yaml
ai_cli_command: llm -s        # Simon Willison's llm
ai_cli_command: sgpt -s        # shell-gpt
```

## Piping to an external LLM CLI

You can also pipe anker output into any LLM CLI tool:

```bash
anker recap today --format detailed | claude -p 'Summarize my workday.'
anker recap today --format detailed | llm 'Summarize my workday.'
```

The `detailed` format works best here -- it includes timestamps and metadata that help the model understand the sequence and context of your work.

The `markdown` format includes full git diffs for richer context but uses more tokens:

```bash
anker recap today --format markdown | claude -p 'Summarize what I worked on. Describe the actual changes, not just commit messages.'
```

## Prompt examples

These work with both `--format ai --prompt "..."` and piped usage.

**Standup format:**
```
Summarize my work from yesterday as a standup update.
Format: what I did, what I plan to do, any blockers.
Keep it to 3-4 sentences.
```

**Weekly report:**
```
Write a weekly summary of my work.
Group by project or area. For each, list key accomplishments and open threads.
Keep it professional but not overly formal.
```

**With decisions and insights:**
```
Summarize my workday. Group by topic, not chronologically.
For each topic, highlight decisions made (and why), key insights or lessons learned,
and open threads. Only include these if actually present in the data.
```

**German:**
```
Fasse meinen Arbeitstag zusammen.
Gruppiere nach Themen. Ueberspringe triviale Aenderungen.
Hebe getroffene Entscheidungen, Erkenntnisse und offene Punkte hervor.
Schreib auf Deutsch, kurz und praegnant.
```

**Detailed changelog:**
```
Create a changelog from this activity log.
List every meaningful change with a one-line description.
Group by repository or project.
```

## Obsidian integration

Set `ai_prompt` in your config to produce Obsidian-friendly output with wikilinks and tags. This way every recap becomes a connected note in your vault.

```yaml
# ~/.anker/config.yaml
ai_backend: cli
ai_cli_command: claude -p

ai_prompt: |
  Summarize my workday. The output will be stored in my Obsidian vault.

  Formatting:
  - Group by topic, not chronologically
  - Use ## for topic headings
  - Bullet points for details, no nested lists
  - Skip trivial changes (typos, formatting)
  - Highlight decisions, insights, and open threads
  - Keep it concise

  At the end of the document:
  - Add a line with tags: #recap #anker
  - Link mentioned projects and tools as [[Wikilinks]]
```

The wikilinks (`[[LPIC-1]]`, `[[Pentest]]`, `[[Nix]]`) automatically connect your recaps to existing notes. Tags like `#recap` make them easy to find and filter.

Prompt resolution order: `--prompt` flag > `ai_prompt` in config > built-in default.

## Claude Code rules

The `examples/claude-rules/` directory contains example rules you can copy to `.claude/rules/` to give Claude Code context about anker when it processes your recaps.

```bash
cp examples/claude-rules/anker-context.md .claude/rules/
```

## Shell alias

If you use the piped approach regularly:

```bash
anker-summary() {
  anker recap "${1:-today}" --format ai
}
```

Then just run `anker-summary` or `anker-summary thisweek`.
