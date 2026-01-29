# Timerange Locales

This directory contains month name translations for different languages.

## Supported Languages

- **English** (en.go) - Full names and abbreviations
- **German** (de.go) - Full names and abbreviations

## Adding a New Language

Adding support for a new language is simple - just create a new file:

### Step 1: Create the file

Create a file named after the language code (e.g., `fr.go` for French, `es.go` for Spanish, `gr.go` for Greek).

### Step 2: Copy the template

Use this template:

```go
package locales

import "time"

// [Language] month names (full and abbreviated)
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

		// Abbreviations (if different from full name)
		"jan": time.January,
		"feb": time.February,
		"mar": time.March,
		// ... etc
	})
}
```

### Step 3: Replace with your language

Replace the month names with translations in your target language.

### Example: French (fr.go)

```go
package locales

import "time"

// French month names (full and abbreviated)
func init() {
	RegisterMonthNames(map[string]time.Month{
		// Full names
		"janvier":   time.January,
		"février":   time.February,
		"mars":      time.March,
		"avril":     time.April,
		"mai":       time.May,
		"juin":      time.June,
		"juillet":   time.July,
		"août":      time.August,
		"septembre": time.September,
		"octobre":   time.October,
		"novembre":  time.November,
		"décembre":  time.December,

		// Abbreviations
		"jan":  time.January,
		"fév":  time.February,
		"mar":  time.March,
		"avr":  time.April,
		"jui":  time.June,
		"juil": time.July,
		"aoû":  time.August,
		"sep":  time.September,
		"oct":  time.October,
		"nov":  time.November,
		"déc":  time.December,
	})
}
```

### That's it!

No other code changes needed. The `init()` function automatically registers your translations when the program starts.

## Usage

After adding a language file, users can use month names in that language:

```bash
anker report "octobre 2025"      # French
anker report "2025 décembre"     # French, year-first
anker report "οκτώβριος 2025"    # Greek (if gr.go exists)
```

All formats are case-insensitive and support both "month year" and "year month" order.

## Notes

- Month names are case-insensitive (automatically lowercased)
- All files in this directory are automatically loaded
- Duplicate month names across languages are fine (e.g., "mai" in both EN and DE)
- Use Unicode characters freely (ä, ö, ü, é, à, etc.) - they're fully supported
