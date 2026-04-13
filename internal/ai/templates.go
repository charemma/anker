package ai

import "slices"

// Style identifies a prompt template for a specific audience.
type Style string

const (
	StyleSelf     Style = "self"
	StyleManager  Style = "manager"
	StyleCustomer Style = "customer"
	StyleStandup  Style = "standup"
	StyleRetro    Style = "retro"
)

// validStyles is the exhaustive list of built-in style identifiers.
var validStyles = []Style{
	StyleSelf,
	StyleManager,
	StyleCustomer,
	StyleStandup,
	StyleRetro,
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
	switch style {
	case StyleCustomer:
		// Customer reports show only git commits -- no internal notes or AI sessions.
		return []string{"git"}
	case StyleStandup:
		// Standup includes git and claude sessions; obsidian is low-signal for daily standups.
		return []string{"git", "claude"}
	default:
		return nil
	}
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
	case StyleManager:
		return promptManager
	case StyleCustomer:
		return promptCustomer
	case StyleStandup:
		return promptStandup
	case StyleRetro:
		return promptRetro
	default:
		return promptSelf
	}
}

const promptSelf = `Summarize the developer's activity log as a personal technical recap. No audience -- this is for the developer reading their own notes.

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

const promptManager = `Summarize the developer's activity log to prepare for a 1:1 meeting with a manager.

## How to read the input

Each line: DATE SOURCE: CONTENT

obsidian -- file modified, no content. Decode the path for topic:
  - 1 Projects/ = active project
  - 2 Areas/ = ongoing responsibility
  - 3 Resources/ = research

claude -- AI session: [project] snippet -- N turns, M min. Low weight if < 3 turns or < 5 min.

git -- commit message. High confidence, always relevant.

## Output structure

Write in German. Three sections:

### Was lief gut
2-4 bullets. Completed work with concrete outcome. Format: "<was> -- <ergebnis oder status>"
State outcomes, not just activities ("Feature geliefert", not "an Feature gearbeitet").

### Wo brauche ich Unterstuetzung
1-3 bullets. Open items, blockers, or unresolved decisions visible in the data.
If nothing is visible, write: "Kein aktueller Blocker erkennbar."

### Naechste Woche
1-2 bullets. What is clearly in-progress or started but not finished.
If nothing is obvious from the data, write: "Offen -- muss noch priorisiert werden."

Rules:
- No passive voice. Active verbs only.
- Name the project or tool, not a generic theme
- No commit hashes or file paths in output
- Do not invent blockers or next steps that are not visible in the data
- Language: German`

const promptCustomer = `Summarize the developer's activity log as a professional status update for a customer or external stakeholder.

## How to read the input

Each line: DATE SOURCE: CONTENT

git -- commit message. Translate technical terms into customer-readable language.

## Output structure

Write in German. Professional tone -- no internal jargon, no repo names, no tool names unless customer-facing.

### Geliefert
2-4 bullets. Completed deliverables only. Describe the outcome and value, not the technical implementation.
Format: "<was wurde gebaut/fertiggestellt> -- <warum das wichtig ist>"

### Projektstand
2-3 sentences. Current state of the main project. What phase, what is working, what is in progress.

### Naechste Schritte
2-3 bullets. What comes next. No vague placeholders like "weitere Entwicklung".

Rules:
- Do not mention commit hashes, file paths, or internal tool names (git, Obsidian, Claude)
- Do not mention effort in minutes or commit counts
- Translate technical outcomes: "refactored the renderer" -> "verbesserte Ausgabequalitaet und Stabilitaet"
- If a feature is in-progress but not done, do not list it under "Geliefert"
- Omit research and internal tooling work that has no direct customer impact
- Language: German`

const promptStandup = `Summarize the activity log for a daily standup. Audience: teammates who know the projects. Keep it tight.

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

const promptRetro = `Summarize the developer's activity log for a sprint retrospective.
Goal: honest reflection, not a highlight reel.

## How to read the input

Each line: DATE SOURCE: CONTENT

obsidian -- file modified. Decode path: 1 Projects/, 3 Resources/, etc. No content available.
claude -- AI session: [project] snippet -- N turns, M min
  - Long sessions (> 60 min) may indicate a hard problem or rabbit hole
  - Short sessions (< 5 min) likely noise

git -- commit message. Pattern: many small commits = iterative; few large commits = batched work.

## Output structure

Write in German. Four sections:

### Was gut lief
2-4 bullets. Things completed, decisions made, good patterns visible in the data.
Format: "<was> -- <warum das gut war>"

### Was nicht so gut lief
2-4 bullets. Rough spots visible in the data: repeated fixes, long unresolved sessions, context switching.
If nothing negative is visible, write 1-2 bullets about what is still uncertain or unfinished.
Do not invent problems that are not signaled in the data.

### Zeitverteilung
1-2 sentences. Where did the bulk of time actually go? Be honest -- if research dominated over delivery, say so.
If a claude session > 200 min, name the topic explicitly.

### Learnings / Naechstes Mal
2-3 bullets. Concrete takeaways based on the patterns above. Actionable, not generic ("besser kommunizieren").
Format: "Beim naechsten Mal: <was genau>"

Rules:
- Active voice throughout
- "Was nicht gut lief" must not be empty -- retrospectives require honest reflection
- Do not mention commit hashes or file paths
- Language: German`
