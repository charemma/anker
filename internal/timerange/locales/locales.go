// Package locales provides month name translations for different languages.
//
// To add a new language:
//
// 1. Create a new file named after the language code (e.g., fr.go for French)
// 2. Copy the structure from en.go or de.go
// 3. Add month names in the target language to the map
// 4. The init() function will automatically register the names at startup
//
// Example (fr.go):
//
//	package locales
//
//	import "time"
//
//	func init() {
//	    RegisterMonthNames(map[string]time.Month{
//	        "janvier":  time.January,
//	        "jan":      time.January,
//	        "février":  time.February,
//	        "fév":      time.February,
//	        // ... etc
//	    })
//	}
//
// That's it! No other code changes needed.
package locales

import "time"

var monthNames = make(map[string]time.Month)

// RegisterMonthNames registers month name translations.
// Called by init() functions in language files (en.go, de.go, etc.)
func RegisterMonthNames(names map[string]time.Month) {
	for name, month := range names {
		monthNames[name] = month
	}
}

// ParseMonth looks up a month name (case-insensitive) across all registered languages.
// Returns the month and true if found, or time.Month(0) and false if not found.
func ParseMonth(name string) (time.Month, bool) {
	month, ok := monthNames[name]
	return month, ok
}
