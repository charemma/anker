package ai

import "slices"

// Style identifies a prompt template by output format.
type Style string

const (
	StyleBrief  Style = "brief"  // Done/Next/Blocker, max 12 lines
	StyleDigest Style = "digest" // thematic overview, all sources, default
	StyleReport Style = "report" // polished formal report, deliveries
	StyleRetro  Style = "retro"  // retrospective: good/bad/learnings
	StyleStatus Style = "status" // progress/blocker/next, progress-focused
)

// validStyles is the exhaustive list of built-in style identifiers.
var validStyles = []Style{
	StyleBrief,
	StyleDigest,
	StyleReport,
	StyleRetro,
	StyleStatus,
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
	if style == StyleBrief {
		// Brief only needs commits and AI sessions -- no vault file changes.
		return []string{"git", "claude"}
	}
	return nil
}

// DefaultTimespec returns the default time specification for a style.
// Most styles default to "today"; brief defaults to "yesterday".
func DefaultTimespec(style Style) string {
	if style == StyleBrief {
		return "yesterday"
	}
	return "today"
}

// Prompt returns the built-in German prompt for the given style.
func Prompt(style Style) string {
	switch style {
	case StyleBrief:
		return promptBrief
	case StyleReport:
		return promptReport
	case StyleRetro:
		return promptRetro
	case StyleStatus:
		return promptStatus
	default:
		return promptDigest
	}
}

const promptBrief = `Summarize the developer's activity log for a daily standup. Keep it tight.

## How to read the input

Each line: DATE SOURCE: CONTENT

claude -- AI session: [project] snippet -- N turns, M min. Skip if < 3 turns or < 5 min.
git -- commit message. Always include.

## Output format

Return exactly three sections. German.

**Done**
2-4 bullets. Completed things only. Format: "<action> in <project/tool>"
One bullet = one concrete thing, not a theme.

**Next**
1-2 bullets. What is clearly next based on open work visible in the data. If nothing obvious, write "Offen -- muss noch entschieden werden."

**Blocker**
0-2 bullets. Only real blockers or missing decisions visible in the data. If none, omit this section entirely.

Rules:
- Max 8 words per bullet
- No timestamps, hashes, or file paths
- Name the project: "in anker", "in nixos-config"
- Active verbs: "implementiert", "gefixt", "dokumentiert" -- not "gearbeitet an"
- Do not pad with filler bullets if the data is sparse
- Language: German`

const promptDigest = `Summarize the developer's activity log as a thematic technical overview.

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
- Skip themes with only 1 low-signal obsidian entry and no commits or claude sessions
- Language: German`

const promptReport = `Write a polished formal report from the developer's activity log.

## How to read the input

Each line: DATE SOURCE: CONTENT

obsidian -- file modified. Decode path for context.
claude -- AI session: [project] snippet -- N turns, M min. Low weight if < 3 turns or < 5 min.
git -- commit message. Translate to outcome language. Merged/shipped work only.

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

const promptStatus = `Write a progress-focused status update from the developer's activity log.

## How to read the input

Each line: DATE SOURCE: CONTENT

obsidian -- file modified. Use the path to infer topic. No content available.
claude -- AI session: [project] snippet -- N turns, M min. Skip short sessions (< 3 turns or < 5 min).
git -- commit message. Always relevant. Group by repo.

## Output format

Write in German. Focus on what is done, what is blocked, and what comes next.

### Fortschritt
3-5 bullets. Completed and in-progress work. Both matter here.
Format: "<was> -- <aktueller Stand: abgeschlossen / in Arbeit>"

### Blocker
0-3 bullets. Unresolved dependencies, open decisions, missing input.
If none visible in the data, omit this section.

### Naechste Schritte
2-4 bullets. Concrete next actions based on open work in the data.
One bullet = one specific action, not a theme.

Rules:
- No commit hashes or file paths
- Keep bullets tight -- no explanations, just facts
- "in Arbeit" means started but not merged/shipped
- Language: German`

const promptRetro = `Write a sprint retrospective from the developer's activity log.

## How to read the input

Each line: DATE SOURCE: CONTENT

obsidian -- file modified. Use the path to infer topic. No content available.
claude -- AI session: [project] snippet -- N turns, M min. Skip short sessions (< 3 turns or < 5 min).
git -- commit message. Always relevant. Group by repo.

## Output format

Write in German. Structured retrospective format.

### Was lief gut
2-4 bullets. Things that went smoothly, clear wins, good decisions made.
Be specific -- name the feature, fix, or approach that worked.

### Was lief schlecht
2-4 bullets. Friction points, repeated back-and-forths, time sinks, wrong turns.
No self-flagellation -- just honest observations.

### Zeitverteilung
2-3 bullets. Where did the time actually go? Use claude session durations and commit density as proxies.
Format: "<topic> -- ca. <N>% der Zeit"

### Learnings
1-3 bullets. What would you do differently? What insight is worth keeping?

Rules:
- No commit hashes or file paths
- Be specific, not generic
- If there is nothing notable for a section, omit it entirely
- Language: German`
