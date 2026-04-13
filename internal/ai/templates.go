package ai

import "slices"

// Style identifies a prompt template by output format.
type Style string

const (
	StyleSummary   Style = "summary"   // short overview, bullet points
	StyleDetailed  Style = "detailed"  // comprehensive, all topics, more context
	StyleStandup   Style = "standup"   // gestern/heute/blocker, minimal
	StyleNarrative Style = "narrative" // flowing prose, email-ready
	StyleReport    Style = "report"    // formal status report, professional tone
)

// validStyles is the exhaustive list of built-in style identifiers.
var validStyles = []Style{
	StyleSummary,
	StyleDetailed,
	StyleStandup,
	StyleNarrative,
	StyleReport,
}

// IsValidStyle reports whether s is a known built-in style.
func IsValidStyle(s string) bool {
	return slices.Contains(validStyles, Style(s))
}

// ValidStyleNames returns all built-in style names as strings.
func ValidStyleNames() []string {
	names := make([]string, len(validStyles))
	for i, s := range validStyles {
		names[i] = string(s)
	}
	return names
}

// AllowedSources returns the source types included for a given style.
// An empty slice means all sources are passed through unfiltered.
func AllowedSources(style Style) []string {
	if style == StyleStandup {
		// Standup only needs commits and AI sessions -- no vault file changes.
		return []string{"git", "claude"}
	}
	return nil
}

// DefaultTimespec returns the default time specification for a style.
// Most styles default to "today"; standup defaults to "yesterday".
func DefaultTimespec(style Style) string {
	if style == StyleStandup {
		return "yesterday"
	}
	return "today"
}

// Prompt returns the built-in German prompt for the given style.
func Prompt(style Style) string {
	switch style {
	case StyleDetailed:
		return promptDetailed
	case StyleStandup:
		return promptStandup
	case StyleNarrative:
		return promptNarrative
	case StyleReport:
		return promptReport
	default:
		return promptSummary
	}
}

const promptSummary = `Summarize the developer's activity log as a short technical overview.

## How to read the input

Each line: DATE SOURCE: CONTENT

obsidian -- a file was modified; no content available. Read the path:
  - 1 Projects/<name>/ = active project work
  - 2 Areas/<topic>/ = ongoing responsibility
  - 3 Resources/<topic>/ = research (e.g. K8s/, Go/, Nix/)
  - Journal/ = daily journal -- skip, no signal

claude -- an AI session. Format: [project] snippet -- N turns, M min
  - Under 3 turns or under 5 min = likely aborted, skip
  - Duration in minutes = effort proxy

git -- a commit message. High-signal, always include.

## Output format

Write in German. No preamble. Start directly with the first bullet.

Structure:
- One section per theme (no heading, just bullets grouped together)
- Each bullet: active verb + what + where (repo or tool)
- After the bullets: one line "Offen:" followed by open items (or omit if nothing obvious)

Rules:
- Group by theme, not by date or source
- Max 3-4 bullets per theme
- Do not list individual commit hashes or file paths
- No vague qualifiers: "intensiv", "erfolgreich", "verschiedene", "einige"
- If something took > 100 claude minutes, mark with "(groesster Zeitblock)"
- Skip themes with only 1 low-signal obsidian entry and no commits or claude sessions`

const promptDetailed = `Write a comprehensive summary of the developer's activity log. Cover all notable themes with full context.

## How to read the input

Each line: DATE SOURCE: CONTENT

obsidian -- file modified, no content. Decode the path:
  - 1 Projects/<name>/ = active project work
  - 2 Areas/<topic>/ = ongoing responsibility
  - 3 Resources/<topic>/ = research or learning
  - Journal/ = daily journal -- low signal, skip unless dense activity

claude -- AI session: [project] snippet -- N turns, M min
  - Under 3 turns or under 5 min = noise, skip
  - Long sessions (> 60 min) = significant effort, always mention

git -- commit message. Always include. Group by repo.

## Output format

Write in German. Cover every theme with activity -- do not skip minor ones.

Structure per theme:
  ### <Theme name>
  2-5 bullets. What was done, in what context, what the outcome was.
  Add a "Status:" line at the end: fertig / in Arbeit / blockiert

Final section "Offene Punkte" (omit if none):
  List unresolved items or open branches visible from the data.

Rules:
- No vague summaries: be specific about what changed or was decided
- Do not list commit hashes or file paths
- If a claude session dominated (> 100 min), describe the topic and duration
- Language: German`

const promptStandup = `Summarize the activity log for a daily standup. Keep it tight.

## How to read the input

Each line: DATE SOURCE: CONTENT

claude -- AI session: [project] snippet -- N turns, M min. Skip if < 3 turns or < 5 min.
git -- commit message. Always include.

## Output format

Return exactly three sections. German.

**Gestern / Diese Woche**
2-4 bullets. Completed things only. Format: "<action> in <project/tool>"
One bullet = one concrete thing, not a theme.

**Heute / Naechste Schritte**
1-2 bullets. What is clearly next based on open work visible in the data. If nothing obvious, write "Offen -- muss noch entschieden werden."

**Blocker**
0-2 bullets. Only real blockers or missing decisions visible in the data. If none, omit this section entirely.

Rules:
- Max 8 words per bullet
- No timestamps, hashes, or file paths
- Name the project: "in anker", "in nixos-config"
- Active verbs: "implementiert", "gefixt", "dokumentiert" -- not "gearbeitet an"
- Do not pad with filler bullets if the data is sparse`

const promptNarrative = `Write the developer's activity as a flowing prose summary -- readable as an email or message.

## How to read the input

Each line: DATE SOURCE: CONTENT

obsidian -- file modified. Use the path to infer topic. No content available.
claude -- AI session: [project] snippet -- N turns, M min. Skip short sessions (< 3 turns or < 5 min).
git -- commit message. Always relevant.

## Output format

Write in German as connected prose. No bullet lists, no section headings.

Structure:
- Opening sentence: overall theme of the period (1 sentence)
- Body: 2-4 paragraphs, one per major topic. Each paragraph covers what was done, why it matters, and current status.
- Closing: 1-2 sentences on what comes next or what is still open.

Rules:
- Write in first person ("Ich habe...", "Diese Woche...")
- No commit hashes, file paths, or internal tool names in the output
- Translate technical work into readable language -- "einen Bug in der Zeitraum-Erkennung behoben" not "fixed parse edge case"
- Do not list everything -- synthesize into a coherent narrative
- Length: 150-250 words
- Language: German`

const promptReport = `Write a formal status report from the developer's activity log.

## How to read the input

Each line: DATE SOURCE: CONTENT

obsidian -- file modified. Decode path for context.
claude -- AI session: [project] snippet -- N turns, M min. Low weight if < 3 turns or < 5 min.
git -- commit message. Translate to non-technical outcome language.

## Output structure

Write in German. Formal tone. No internal jargon, no tool names unless relevant.

### Fortschritt
2-4 bullets. Completed items only. Format: "<was abgeschlossen> -- <Mehrwert oder Ergebnis>"
Describe outcomes and value, not implementation details.

### Aktueller Stand
2-3 sentences. Current project state: what phase, what works, what is in progress.

### Naechste Schritte
2-3 bullets. Concrete next items. No vague placeholders.

Rules:
- No commit hashes, file paths, or internal tool names (git, Obsidian, Claude)
- No effort metrics (minutes, commit counts)
- In-progress work does not appear under "Fortschritt"
- Translate implementation work: "Renderer umgebaut" -> "verbesserte Ausgabequalitaet"
- Language: German`
