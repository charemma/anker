package recap

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

var germanWeekdays = []string{
	"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag",
}

var germanMonths = []string{
	"", "Januar", "Februar", "März", "April", "Mai", "Juni",
	"Juli", "August", "September", "Oktober", "November", "Dezember",
}

// germanDate formats a time.Time as "Montag, 7. April" or "7. April".
func germanDate(t time.Time, includeWeekday bool) string {
	month := germanMonths[int(t.Month())]
	if includeWeekday {
		return fmt.Sprintf("%s, %d. %s", germanWeekdays[int(t.Weekday())], t.Day(), month)
	}
	return fmt.Sprintf("%d. %s", t.Day(), month)
}

// DayGroup holds entries for a single calendar day.
type DayGroup struct {
	Date    time.Time
	Entries []sources.Entry
}

// GroupByDay buckets entries into per-day groups covering every day from from to to.
// Groups are sorted oldest-first. Entries within each day are also sorted oldest-first.
func GroupByDay(entries []sources.Entry, from, to time.Time) []DayGroup {
	byDay := make(map[string][]sources.Entry)
	for _, e := range entries {
		key := e.Timestamp.Format("2006-01-02")
		byDay[key] = append(byDay[key], e)
	}

	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())

	var groups []DayGroup
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		dayEntries := byDay[key]
		sort.Slice(dayEntries, func(i, j int) bool {
			return dayEntries[i].Timestamp.Before(dayEntries[j].Timestamp)
		})
		groups = append(groups, DayGroup{Date: d, Entries: dayEntries})
	}
	return groups
}

// entrySourceLabel returns the display label for an entry.
// git: "git/<repo-name>", all others: source type.
func entrySourceLabel(e sources.Entry) string {
	if e.Source == "git" {
		name := e.Location
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		return "git/" + name
	}
	return e.Source
}

// footerKey strips the "git/" prefix so the footer shows "anker" not "git/anker".
func footerKey(label string) string {
	return strings.TrimPrefix(label, "git/")
}

var ccTypeRE = regexp.MustCompile(`^([a-z]+)(?:\([^)]*\))?!?:`)

func parseCCType(msg string) string {
	m := ccTypeRE.FindStringSubmatch(msg)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// entryDisplayContent returns a clean display string for an entry.
// For obsidian it strips the trailing " (modified)" / " (created)" action suffix.
func entryDisplayContent(e sources.Entry) string {
	if e.Source == "obsidian" {
		if idx := strings.LastIndex(e.Content, " ("); idx >= 0 {
			return e.Content[:idx]
		}
	}
	return e.Content
}

// sourceAgg aggregates entries from one source label for a single day.
type sourceAgg struct {
	label      string
	srcType    string
	count      int
	ccTypes    map[string]int
	allContent []string // display content for every entry, oldest first
}

// aggregateBySource groups a day's entries by source label.
// The returned slice preserves insertion order (first-seen).
func aggregateBySource(entries []sources.Entry) []sourceAgg {
	var order []string
	aggs := make(map[string]*sourceAgg)

	for _, e := range entries {
		lbl := entrySourceLabel(e)
		agg, exists := aggs[lbl]
		if !exists {
			agg = &sourceAgg{
				label:   lbl,
				srcType: e.Source,
				ccTypes: make(map[string]int),
			}
			aggs[lbl] = agg
			order = append(order, lbl)
		}
		agg.count++
		agg.allContent = append(agg.allContent, entryDisplayContent(e))
		if e.Source == "git" {
			if t := parseCCType(e.Content); t != "" {
				agg.ccTypes[t]++
			}
		}
	}

	result := make([]sourceAgg, 0, len(order))
	for _, lbl := range order {
		result = append(result, *aggs[lbl])
	}
	return result
}

// unitWord returns the German plural/singular unit for a source type.
func unitWord(srcType string, n int) string {
	switch srcType {
	case "git":
		if n == 1 {
			return "commit"
		}
		return "commits"
	case "claude":
		if n == 1 {
			return "session"
		}
		return "sessions"
	case "obsidian", "markdown":
		if n == 1 {
			return "Notiz"
		}
		return "Notizen"
	default:
		if n == 1 {
			return "Eintrag"
		}
		return "Einträge"
	}
}

// ccSummary formats CC type counts as "(2 feat, 1 fix)".
func ccSummary(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}
	type kv struct {
		k string
		v int
	}
	pairs := make([]kv, 0, len(counts))
	for k, v := range counts {
		pairs = append(pairs, kv{k, v})
	}
	// Sort by count desc, then name asc for stability.
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].v != pairs[j].v {
			return pairs[i].v > pairs[j].v
		}
		return pairs[i].k < pairs[j].k
	})
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = fmt.Sprintf("%d %s", p.v, p.k)
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// buildFooter returns "N entries · label1 (n1), label2 (n2), ..."
func buildFooter(entries []sources.Entry) string {
	order := []string{}
	counts := map[string]int{}
	for _, e := range entries {
		key := footerKey(entrySourceLabel(e))
		if _, seen := counts[key]; !seen {
			order = append(order, key)
		}
		counts[key]++
	}
	parts := make([]string, len(order))
	for i, k := range order {
		parts[i] = fmt.Sprintf("%s (%d)", k, counts[k])
	}
	total := fmt.Sprintf("%d entries", len(entries))
	if len(parts) == 0 {
		return total
	}
	return total + " · " + strings.Join(parts, ", ")
}

// showTopN is the maximum number of entry lines shown per source in week view.
const showTopN = 3

// rangeIsShort reports whether the time range is short enough to warrant
// per-entry detail lines (≤ 7 days).
func rangeIsShort(result *RecapResult) bool {
	days := result.TimeRange.To.Sub(result.TimeRange.From).Hours() / 24
	return days <= 7
}

// RenderSummary is the default renderer. For timespec "today" it shows all
// entries individually with a time column. For short ranges (≤7 days) it shows
// a day-by-day view with top-3 entry lines per source. For longer ranges it
// shows aggregate numbers only. When plain is true, no ANSI codes are emitted.
func RenderSummary(w io.Writer, result *RecapResult, plain bool) error {
	if result.Timespec == "today" {
		return renderToday(w, result, plain)
	}
	return renderWeek(w, result, plain)
}

func renderToday(w io.Writer, result *RecapResult, plain bool) error {
	// Collect entries sorted oldest first
	entries := make([]sources.Entry, len(result.Entries))
	copy(entries, result.Entries)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	if plain {
		for _, e := range entries {
			label := entrySourceLabel(e)
			_, _ = fmt.Fprintf(w, "%s %s: %s\n", e.Timestamp.Format("2006-01-02"), label, e.Content)
		}
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, buildFooter(entries))
		return nil
	}

	// Compute max label width for column alignment.
	maxLabelW := 0
	for _, e := range entries {
		if l := len(entrySourceLabel(e)); l > maxLabelW {
			maxLabelW = l
		}
	}

	// Date header
	date := result.TimeRange.From
	if len(entries) > 0 {
		date = entries[0].Timestamp
	}
	dayHeader := germanDate(date, true)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, ui.StyleDay.Render(dayHeader))
	_, _ = fmt.Fprintln(w)

	for _, e := range entries {
		timeStr := e.Timestamp.Format("15:04")
		label := entrySourceLabel(e)

		timeStyled := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(timeStr)
		labelStyled := lipgloss.NewStyle().Foreground(ui.SourceColor(e.Source)).
			Width(maxLabelW).Render(label)
		contentStyled := lipgloss.NewStyle().Foreground(ui.ColorNormal).Render(e.Content)

		_, _ = fmt.Fprintf(w, "  %s  %s  %s\n", timeStyled, labelStyled, contentStyled)
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, ui.StyleMuted.Render(buildFooter(entries)))
	_, _ = fmt.Fprintln(w, ui.StyleMuted.Render("For a narrative summary: anker recap today --format ai"))
	return nil
}

func renderWeek(w io.Writer, result *RecapResult, plain bool) error {
	// Collect entries sorted oldest first for plain output.
	entries := make([]sources.Entry, len(result.Entries))
	copy(entries, result.Entries)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	if plain {
		for _, e := range entries {
			label := entrySourceLabel(e)
			_, _ = fmt.Fprintf(w, "%s %s: %s\n", e.Timestamp.Format("2006-01-02"), label, e.Content)
		}
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "%d entries (%s to %s)\n",
			len(entries),
			result.TimeRange.From.Format("2006-01-02"),
			result.TimeRange.To.Format("2006-01-02"))
		return nil
	}

	showDetails := rangeIsShort(result)
	days := GroupByDay(entries, result.TimeRange.From, result.TimeRange.To)

	// Header
	_, _ = fmt.Fprintln(w)
	weekHeader := buildWeekHeader(result)
	_, _ = fmt.Fprintln(w, lipgloss.NewStyle().Bold(true).Render(weekHeader))

	for _, day := range days {
		_, _ = fmt.Fprintln(w)
		dayLabel := germanDate(day.Date, true)

		if len(day.Entries) == 0 {
			emptyLine := fmt.Sprintf("%-24s -", dayLabel)
			_, _ = fmt.Fprintln(w, ui.StyleMuted.Render(emptyLine))
			continue
		}

		// Day header with activity count
		countStr := fmt.Sprintf("%d Aktivitäten", len(day.Entries))
		dayLine := fmt.Sprintf("%-24s %s", dayLabel, countStr)
		_, _ = fmt.Fprintln(w, ui.StyleDay.Render(dayLine))
		_, _ = fmt.Fprintln(w)

		aggs := aggregateBySource(day.Entries)
		maxLabelW := 0
		for _, a := range aggs {
			if l := len(a.label); l > maxLabelW {
				maxLabelW = l
			}
		}

		for _, agg := range aggs {
			labelStyled := lipgloss.NewStyle().Foreground(ui.SourceColor(agg.srcType)).
				Width(maxLabelW).Render(agg.label)

			unit := unitWord(agg.srcType, agg.count)
			countPart := fmt.Sprintf("%d %s", agg.count, unit)

			line := fmt.Sprintf("  %s  %s", labelStyled, countPart)

			// CC type summary for git
			if agg.srcType == "git" && len(agg.ccTypes) > 0 {
				cc := ccSummary(agg.ccTypes)
				line += "  " + ui.StyleMuted.Render(cc)
			}

			_, _ = fmt.Fprintln(w, line)

			// For short ranges: show top-N entry lines below the summary row.
			if showDetails && agg.count > 0 {
				indent := strings.Repeat(" ", 2+maxLabelW+2)
				shown := agg.allContent
				rest := 0
				if len(shown) > showTopN {
					rest = len(shown) - showTopN
					shown = shown[:showTopN]
				}
				for _, c := range shown {
					_, _ = fmt.Fprintln(w, ui.StyleMuted.Render(indent+c))
				}
				if rest > 0 {
					_, _ = fmt.Fprintln(w, ui.StyleMuted.Render(
						fmt.Sprintf("%s[+%d weitere]", indent, rest)))
				}
			}
		}
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, ui.StyleMuted.Render(buildFooter(entries)))
	_, _ = fmt.Fprintln(w, ui.StyleMuted.Render("For a narrative summary: anker recap "+result.Timespec+" --format ai"))
	return nil
}

// buildWeekHeader returns a header like "Woche 15  (7.-11. April)   24 Aktivitäten in 3 Quellen".
func buildWeekHeader(result *RecapResult) string {
	from := result.TimeRange.From
	to := result.TimeRange.To

	_, week := from.ISOWeek()

	// Date range abbreviation
	var dateRange string
	if from.Month() == to.Month() {
		dateRange = fmt.Sprintf("%d.-%d. %s", from.Day(), to.Day(), germanMonths[int(from.Month())])
	} else {
		dateRange = fmt.Sprintf("%d. %s - %d. %s",
			from.Day(), germanMonths[int(from.Month())],
			to.Day(), germanMonths[int(to.Month())])
	}

	// Count distinct source labels
	labels := map[string]bool{}
	for _, e := range result.Entries {
		labels[entrySourceLabel(e)] = true
	}

	return fmt.Sprintf("Woche %d  (%s)   %d Aktivitäten in %d Quellen",
		week, dateRange, len(result.Entries), len(labels))
}
