package locales

import "time"

// German month names (full and abbreviated)
func init() {
	RegisterMonthNames(map[string]time.Month{
		// Full names
		"januar":   time.January,
		"februar":  time.February,
		"märz":     time.March,
		"april":    time.April,
		"mai":      time.May,
		"juni":     time.June,
		"juli":     time.July,
		"august":   time.August,
		"september": time.September,
		"oktober":  time.October,
		"november": time.November,
		"dezember": time.December,

		// Abbreviations (where different from full name)
		"jan": time.January,
		"feb": time.February,
		"mär": time.March,
		"okt": time.October,
		"dez": time.December,
	})
}
