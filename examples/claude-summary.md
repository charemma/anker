# Generating work summaries with AI

ikno uses AI by default to generate recap summaries. Just run `ikno recap today` and get a formatted summary.

## Built-in styles

```bash
ikno recap today                        # default style (digest)
ikno recap today --style brief          # quick 3-5 bullets
ikno recap thisweek --style status      # standup format
ikno recap thisweek --style report      # professional report
ikno recap yesterday --style retro      # retrospective
ikno recap thisweek --style stats       # work statistics
```

Control the language with `--lang`:

```bash
ikno recap today --lang english
ikno recap today --lang deutsch
```

Set defaults in config:

```yaml
# ~/.config/ikno/config.yaml
ai_default_style: digest
ai_language: deutsch
```

## CLI backend

If you have a Claude Pro/Max subscription or another LLM CLI tool, use the CLI backend to avoid API costs:

```yaml
# ~/.config/ikno/config.yaml
ai_backend: cli
ai_cli_command: claude -p     # default
```

This pipes your recap data via stdin to the CLI tool with the prompt as the last argument.

Other CLI tools work too:

```yaml
ai_cli_command: llm -s        # Simon Willison's llm
ai_cli_command: sgpt -s        # shell-gpt
```

## API backend

For direct API access:

```yaml
ai_backend: api
ai_base_url: https://api.anthropic.com/v1/
ai_model: claude-sonnet-4-20250514
ai_api_key: sk-...
```

Supports any OpenAI-compatible endpoint (Anthropic, OpenAI, ollama, vllm).

API key resolution: `--api-key` flag > `AI_API_KEY` env var > config file.

## Piping to an external LLM CLI

You can also bypass the built-in AI and pipe raw output:

```bash
ikno recap today --raw | claude -p 'Summarize my workday.'
ikno recap today --raw | llm 'Summarize my workday.'
```

## Custom prompts

Override the built-in prompt templates:

```yaml
ai_prompt: |
  Summarize my workday. Group by topic, not chronologically.
  Write in German. Keep it concise.
```

Custom prompt templates can also be stored as `.md` files.

## Prompt examples

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

**German:**
```
Fasse meinen Arbeitstag zusammen.
Gruppiere nach Themen. Ueberspringe triviale Aenderungen.
Hebe getroffene Entscheidungen, Erkenntnisse und offene Punkte hervor.
Schreib auf Deutsch, kurz und praegnant.
```

## Obsidian integration

Set `ai_prompt` to produce Obsidian-friendly output:

```yaml
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
  - Add a line with tags: #recap #ikno
  - Link mentioned projects and tools as [[Wikilinks]]
```

## Claude Code rules

The `examples/claude-rules/` directory contains example rules you can copy to `.claude/rules/` to give Claude Code context about ikno when it processes your recaps.

```bash
cp examples/claude-rules/ikno-context.md .claude/rules/
```

## Shell alias

```bash
ikno-summary() {
  ikno recap "${1:-today}"
}
```

Then just run `ikno-summary` or `ikno-summary thisweek`.
