package timerange

import (
	"testing"
	"time"
)

func TestParser_ParseToday(t *testing.T) {
	now := time.Date(2025, 6, 15, 14, 30, 0, 0, time.Local)
	parser := &Parser{
		config: DefaultConfig(),
		now:    now,
	}

	tr, err := parser.Parse("today")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedStart := time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local)
	expectedEnd := time.Date(2025, 6, 15, 23, 59, 59, 999999999, time.Local)

	if !tr.From.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, tr.From)
	}

	if !tr.To.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, tr.To)
	}
}

func TestParser_ParseYesterday(t *testing.T) {
	now := time.Date(2025, 6, 15, 14, 30, 0, 0, time.Local)
	parser := &Parser{
		config: DefaultConfig(),
		now:    now,
	}

	tr, err := parser.Parse("yesterday")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedStart := time.Date(2025, 6, 14, 0, 0, 0, 0, time.Local)
	expectedEnd := time.Date(2025, 6, 14, 23, 59, 59, 999999999, time.Local)

	if !tr.From.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, tr.From)
	}

	if !tr.To.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, tr.To)
	}
}

func TestParser_ParseThisWeek_Monday(t *testing.T) {
	// Wednesday, June 18, 2025
	now := time.Date(2025, 6, 18, 14, 30, 0, 0, time.Local)
	parser := &Parser{
		config: &Config{WeekStart: Monday},
		now:    now,
	}

	tr, err := parser.Parse("thisweek")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Week should start on Monday, June 16
	expectedStart := time.Date(2025, 6, 16, 0, 0, 0, 0, time.Local)
	// And end on Sunday, June 22
	expectedEnd := time.Date(2025, 6, 22, 23, 59, 59, 999999999, time.Local)

	if !tr.From.Equal(expectedStart) {
		t.Errorf("expected start %v (Monday), got %v", expectedStart, tr.From)
	}

	if !tr.To.Equal(expectedEnd) {
		t.Errorf("expected end %v (Sunday), got %v", expectedEnd, tr.To)
	}
}

func TestParser_ParseThisWeek_Sunday(t *testing.T) {
	// Wednesday, June 18, 2025
	now := time.Date(2025, 6, 18, 14, 30, 0, 0, time.Local)
	parser := &Parser{
		config: &Config{WeekStart: Sunday},
		now:    now,
	}

	tr, err := parser.Parse("thisweek")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Week should start on Sunday, June 15
	expectedStart := time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local)
	// And end on Saturday, June 21
	expectedEnd := time.Date(2025, 6, 21, 23, 59, 59, 999999999, time.Local)

	if !tr.From.Equal(expectedStart) {
		t.Errorf("expected start %v (Sunday), got %v", expectedStart, tr.From)
	}

	if !tr.To.Equal(expectedEnd) {
		t.Errorf("expected end %v (Saturday), got %v", expectedEnd, tr.To)
	}
}

func TestParser_ParseLastWeek(t *testing.T) {
	// Wednesday, June 18, 2025
	now := time.Date(2025, 6, 18, 14, 30, 0, 0, time.Local)
	parser := &Parser{
		config: &Config{WeekStart: Monday},
		now:    now,
	}

	tr, err := parser.Parse("lastweek")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Last week: Monday, June 9 to Sunday, June 15
	expectedStart := time.Date(2025, 6, 9, 0, 0, 0, 0, time.Local)
	expectedEnd := time.Date(2025, 6, 15, 23, 59, 59, 999999999, time.Local)

	if !tr.From.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, tr.From)
	}

	if !tr.To.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, tr.To)
	}
}

func TestParser_ParseWeekNumber(t *testing.T) {
	now := time.Date(2025, 6, 15, 14, 30, 0, 0, time.Local)
	parser := &Parser{
		config: &Config{WeekStart: Monday},
		now:    now,
	}

	tests := []struct {
		spec string
		want bool
	}{
		{"week 1", true},
		{"week 25", true},
		{"week 52", true},
		{"week 1 2024", true},
		{"week 0", false},
		{"week 54", false},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			tr, err := parser.Parse(tt.spec)
			if tt.want && err != nil {
				t.Errorf("expected success, got error: %v", err)
			}
			if !tt.want && err == nil {
				t.Errorf("expected error, got success with range: %v", tr)
			}
			if tt.want && tr == nil {
				t.Error("expected time range, got nil")
			}
		})
	}
}

func TestParser_ParseSingleDate(t *testing.T) {
	parser := NewParser(nil)

	tr, err := parser.Parse("2025-12-02")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedStart := time.Date(2025, 12, 2, 0, 0, 0, 0, time.Local)
	expectedEnd := time.Date(2025, 12, 2, 23, 59, 59, 999999999, time.Local)

	if !tr.From.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, tr.From)
	}

	if !tr.To.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, tr.To)
	}
}

func TestParser_ParseDateRange(t *testing.T) {
	parser := NewParser(nil)

	tr, err := parser.Parse("2025-12-01..2025-12-31")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedStart := time.Date(2025, 12, 1, 0, 0, 0, 0, time.Local)
	expectedEnd := time.Date(2025, 12, 31, 23, 59, 59, 999999999, time.Local)

	if !tr.From.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, tr.From)
	}

	if !tr.To.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, tr.To)
	}
}

func TestParser_ParseRelative(t *testing.T) {
	now := time.Date(2025, 6, 15, 14, 30, 0, 0, time.Local)
	parser := &Parser{
		config: DefaultConfig(),
		now:    now,
	}

	tests := []struct {
		spec         string
		expectedDays int
	}{
		{"last 7 days", 7},
		{"last 1 day", 1},
		{"last 30 days", 30},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			tr, err := parser.Parse(tt.spec)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Calculate expected range
			expectedEnd := time.Date(2025, 6, 15, 23, 59, 59, 999999999, time.Local)
			expectedStart := time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local).AddDate(0, 0, -tt.expectedDays+1)

			if !tr.From.Equal(expectedStart) {
				t.Errorf("expected start %v, got %v", expectedStart, tr.From)
			}

			if !tr.To.Equal(expectedEnd) {
				t.Errorf("expected end %v, got %v", expectedEnd, tr.To)
			}
		})
	}
}

func TestParser_ParseInvalid(t *testing.T) {
	parser := NewParser(nil)

	invalid := []string{
		"invalid",
		"week",
		"2025-13-01",      // invalid month
		"2025-12-01...31", // triple dots
		"last week",       // should be "lastweek"
		"this week",       // should be "thisweek"
	}

	for _, spec := range invalid {
		t.Run(spec, func(t *testing.T) {
			_, err := parser.Parse(spec)
			if err == nil {
				t.Errorf("expected error for spec %q, got nil", spec)
			}
		})
	}
}

func TestParser_CaseInsensitive(t *testing.T) {
	parser := NewParser(nil)

	specs := []string{
		"TODAY",
		"Today",
		"YESTERDAY",
		"Yesterday",
		"THISWEEK",
		"ThisWeek",
	}

	for _, spec := range specs {
		t.Run(spec, func(t *testing.T) {
			_, err := parser.Parse(spec)
			if err != nil {
				t.Errorf("expected success for %q, got error: %v", spec, err)
			}
		})
	}
}

func TestParser_ParseMonthYear(t *testing.T) {
	parser := NewParser(nil)

	tests := []struct {
		spec          string
		expectedMonth time.Month
		expectedYear  int
	}{
		{"october 2025", time.October, 2025},
		{"October 2025", time.October, 2025},
		{"OCTOBER 2025", time.October, 2025},
		{"oct 2025", time.October, 2025},
		{"Oct 2025", time.October, 2025},
		{"2025 october", time.October, 2025},
		{"2025 October", time.October, 2025},
		{"2025 oct", time.October, 2025},
		{"january 2024", time.January, 2024},
		{"jan 2024", time.January, 2024},
		{"december 2025", time.December, 2025},
		{"dec 2025", time.December, 2025},
		{"september 2025", time.September, 2025},
		{"sep 2025", time.September, 2025},
		{"sept 2025", time.September, 2025},
		// German
		{"januar 2025", time.January, 2025},
		{"februar 2025", time.February, 2025},
		{"märz 2025", time.March, 2025},
		{"mai 2025", time.May, 2025},
		{"juni 2025", time.June, 2025},
		{"juli 2025", time.July, 2025},
		{"oktober 2025", time.October, 2025},
		{"dezember 2025", time.December, 2025},
		{"mär 2025", time.March, 2025},
		{"okt 2025", time.October, 2025},
		{"dez 2025", time.December, 2025},
		// German year-first
		{"2025 oktober", time.October, 2025},
		{"2025 März", time.March, 2025},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			tr, err := parser.Parse(tt.spec)
			if err != nil {
				t.Fatalf("Parse failed for %q: %v", tt.spec, err)
			}

			// First day of month at 00:00:00
			expectedStart := time.Date(tt.expectedYear, tt.expectedMonth, 1, 0, 0, 0, 0, time.Local)
			// Last day of month at 23:59:59.999999999
			expectedEnd := time.Date(tt.expectedYear, tt.expectedMonth+1, 0, 23, 59, 59, 999999999, time.Local)

			if !tr.From.Equal(expectedStart) {
				t.Errorf("expected start %v, got %v", expectedStart, tr.From)
			}

			if !tr.To.Equal(expectedEnd) {
				t.Errorf("expected end %v, got %v", expectedEnd, tr.To)
			}

			// Verify it spans the entire month
			if tr.From.Month() != tt.expectedMonth {
				t.Errorf("expected month %v, got %v", tt.expectedMonth, tr.From.Month())
			}
			if tr.To.Month() != tt.expectedMonth {
				t.Errorf("expected month %v, got %v", tt.expectedMonth, tr.To.Month())
			}
		})
	}
}

func TestParser_ParseMonthYear_Invalid(t *testing.T) {
	parser := NewParser(nil)

	invalid := []string{
		"invalidmonth 2025",
		"oct",              // missing year
		"2025",             // missing month
		"october 25",       // year too short
		"13 2025",          // numeric month not supported here
		"oct-2025",         // wrong separator
		"october/2025",     // wrong separator
	}

	for _, spec := range invalid {
		t.Run(spec, func(t *testing.T) {
			_, err := parser.Parse(spec)
			if err == nil {
				t.Errorf("expected error for spec %q, got nil", spec)
			}
		})
	}
}
