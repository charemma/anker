package locales

import "time"

// English month names (full and abbreviated)
func init() {
	RegisterMonthNames(map[string]time.Month{
		// Full names
		"january":   time.January,
		"february":  time.February,
		"march":     time.March,
		"april":     time.April,
		"may":       time.May,
		"june":      time.June,
		"july":      time.July,
		"august":    time.August,
		"september": time.September,
		"october":   time.October,
		"november":  time.November,
		"december":  time.December,

		// Abbreviations
		"jan":  time.January,
		"feb":  time.February,
		"mar":  time.March,
		"apr":  time.April,
		"jun":  time.June,
		"jul":  time.July,
		"aug":  time.August,
		"sep":  time.September,
		"sept": time.September,
		"oct":  time.October,
		"nov":  time.November,
		"dec":  time.December,
	})
}
