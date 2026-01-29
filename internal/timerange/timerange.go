package timerange

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charemma/anker/internal/timerange/locales"
)

// TimeRange represents a time interval with a start and end time.
type TimeRange struct {
	From time.Time
	To   time.Time
}

// WeekStart defines which day the week starts on.
type WeekStart int

const (
	Sunday WeekStart = iota
	Monday
)

// Config holds configuration for time range parsing.
type Config struct {
	WeekStart WeekStart
}

// DefaultConfig returns the default configuration (Monday start).
func DefaultConfig() *Config {
	return &Config{
		WeekStart: Monday,
	}
}

// Parser handles parsing of time specifications.
type Parser struct {
	config *Config
	now    time.Time
}

// NewParser creates a new time range parser.
func NewParser(config *Config) *Parser {
	if config == nil {
		config = DefaultConfig()
	}
	return &Parser{
		config: config,
		now:    time.Now(),
	}
}

// Parse parses a time specification string into a TimeRange.
// Supported formats:
//   - "today" - current day
//   - "yesterday" - previous day
//   - "thisweek" - current week
//   - "lastweek" - previous week
//   - "october 2025" - specific month (multilingual, case-insensitive)
//   - "2025 october" - year first format
//   - "week 32" - specific calendar week
//   - "2025-12-02" - specific date
//   - "2025-12-01..2025-12-31" - date range
//   - "last 7 days" - relative range
//
// Month names are supported in multiple languages (English, German, etc.).
// See internal/timerange/locales/ for available languages and how to add more.
func (p *Parser) Parse(spec string) (*TimeRange, error) {
	spec = strings.TrimSpace(strings.ToLower(spec))

	switch spec {
	case "today":
		return p.parseToday(), nil
	case "yesterday":
		return p.parseYesterday(), nil
	case "thisweek":
		return p.parseThisWeek(), nil
	case "lastweek":
		return p.parseLastWeek(), nil
	}

	// Try month and year: "october 2025", "2025 october"
	if tr, ok := p.tryParseMonthYear(spec); ok {
		return tr, nil
	}

	// Try week number: "week 32"
	if tr, ok := p.tryParseWeekNumber(spec); ok {
		return tr, nil
	}

	// Try date range: "2025-12-01..2025-12-31"
	if tr, ok := p.tryParseDateRange(spec); ok {
		return tr, nil
	}

	// Try single date: "2025-12-02"
	if tr, ok := p.tryParseSingleDate(spec); ok {
		return tr, nil
	}

	// Try relative: "last 7 days"
	if tr, ok := p.tryParseRelative(spec); ok {
		return tr, nil
	}

	return nil, fmt.Errorf("unsupported time specification: %s", spec)
}

func (p *Parser) parseToday() *TimeRange {
	start := startOfDay(p.now)
	end := endOfDay(p.now)
	return &TimeRange{From: start, To: end}
}

func (p *Parser) parseYesterday() *TimeRange {
	yesterday := p.now.AddDate(0, 0, -1)
	start := startOfDay(yesterday)
	end := endOfDay(yesterday)
	return &TimeRange{From: start, To: end}
}

func (p *Parser) parseThisWeek() *TimeRange {
	start := p.startOfWeek(p.now)
	end := p.endOfWeek(p.now)
	return &TimeRange{From: start, To: end}
}

func (p *Parser) parseLastWeek() *TimeRange {
	lastWeek := p.now.AddDate(0, 0, -7)
	start := p.startOfWeek(lastWeek)
	end := p.endOfWeek(lastWeek)
	return &TimeRange{From: start, To: end}
}

func (p *Parser) tryParseMonthYear(spec string) (*TimeRange, bool) {
	// Try "october 2025" or "oct 2025" (support unicode letters for non-English months)
	re1 := regexp.MustCompile(`^(\p{L}+)\s+(\d{4})$`)
	matches := re1.FindStringSubmatch(spec)

	if matches == nil {
		// Try "2025 october" or "2025 oct"
		re2 := regexp.MustCompile(`^(\d{4})\s+(\p{L}+)$`)
		matches = re2.FindStringSubmatch(spec)
		if matches != nil {
			// Swap order: matches[1] is year, matches[2] is month
			matches = []string{matches[0], matches[2], matches[1]}
		}
	}

	if matches == nil {
		return nil, false
	}

	monthStr := matches[1]
	yearStr := matches[2]

	// Parse month name (full or abbreviated) using registered locales
	month, ok := locales.ParseMonth(monthStr)
	if !ok {
		return nil, false
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil, false
	}

	// First day of the month
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	// Last day of the month
	end := time.Date(year, month+1, 0, 23, 59, 59, 999999999, time.Local)

	return &TimeRange{From: start, To: end}, true
}

func (p *Parser) tryParseWeekNumber(spec string) (*TimeRange, bool) {
	// Format: "week 32" or "week 32 2025"
	re := regexp.MustCompile(`^week\s+(\d+)(?:\s+(\d{4}))?$`)
	matches := re.FindStringSubmatch(spec)
	if matches == nil {
		return nil, false
	}

	weekNum, _ := strconv.Atoi(matches[1])

	year := p.now.Year()
	if matches[2] != "" {
		year, _ = strconv.Atoi(matches[2])
	}

	if weekNum < 1 || weekNum > 53 {
		return nil, false
	}

	start, end := p.weekBounds(year, weekNum)
	return &TimeRange{From: start, To: end}, true
}

func (p *Parser) tryParseDateRange(spec string) (*TimeRange, bool) {
	// Format: "2025-12-01..2025-12-31"
	parts := strings.Split(spec, "..")
	if len(parts) != 2 {
		return nil, false
	}

	from, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(parts[0]), time.Local)
	if err != nil {
		return nil, false
	}

	to, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(parts[1]), time.Local)
	if err != nil {
		return nil, false
	}

	return &TimeRange{
		From: startOfDay(from),
		To:   endOfDay(to),
	}, true
}

func (p *Parser) tryParseSingleDate(spec string) (*TimeRange, bool) {
	// Format: "2025-12-02"
	date, err := time.ParseInLocation("2006-01-02", spec, time.Local)
	if err != nil {
		return nil, false
	}

	return &TimeRange{
		From: startOfDay(date),
		To:   endOfDay(date),
	}, true
}

func (p *Parser) tryParseRelative(spec string) (*TimeRange, bool) {
	// Format: "last 7 days"
	re := regexp.MustCompile(`^last\s+(\d+)\s+days?$`)
	matches := re.FindStringSubmatch(spec)
	if matches == nil {
		return nil, false
	}

	days, _ := strconv.Atoi(matches[1])

	end := endOfDay(p.now)
	start := startOfDay(p.now.AddDate(0, 0, -days+1))

	return &TimeRange{From: start, To: end}, true
}

func (p *Parser) startOfWeek(t time.Time) time.Time {
	// Start of day first
	t = startOfDay(t)

	weekday := int(t.Weekday())
	weekStartDay := int(p.config.WeekStart)

	// Calculate days to subtract
	daysBack := (weekday - weekStartDay + 7) % 7

	return t.AddDate(0, 0, -daysBack)
}

func (p *Parser) endOfWeek(t time.Time) time.Time {
	start := p.startOfWeek(t)
	return endOfDay(start.AddDate(0, 0, 6))
}

func (p *Parser) weekBounds(year, week int) (time.Time, time.Time) {
	// Find the first day of the year
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)

	// Find the first week start day of the year
	firstWeekStart := p.startOfWeek(jan1)

	// If the first week start is in the previous year, add a week
	if firstWeekStart.Year() < year {
		firstWeekStart = firstWeekStart.AddDate(0, 0, 7)
	}

	// Calculate the target week
	targetWeekStart := firstWeekStart.AddDate(0, 0, (week-1)*7)
	targetWeekEnd := p.endOfWeek(targetWeekStart)

	return targetWeekStart, targetWeekEnd
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}
