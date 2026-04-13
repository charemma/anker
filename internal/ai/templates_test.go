package ai

import (
	"testing"

	"github.com/charemma/ikno/internal/sources"
)

func TestIsValidStyle(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"brief", true},
		{"digest", true},
		{"report", true},
		{"retro", true},
		{"stats", true},
		{"status", true},
		{"unknown", false},
		{"", false},
		{"BRIEF", false}, // case-sensitive
	}
	for _, tt := range tests {
		if got := IsValidStyle(tt.input); got != tt.want {
			t.Errorf("IsValidStyle(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestValidStyleNames(t *testing.T) {
	names := ValidStyleNames()
	if len(names) != 6 {
		t.Fatalf("expected 6 style names, got %d", len(names))
	}
	want := map[string]bool{"brief": true, "digest": true, "report": true, "retro": true, "stats": true, "status": true}
	for _, name := range names {
		if !want[name] {
			t.Errorf("unexpected style name %q", name)
		}
	}
}

func TestDefaultTimespec(t *testing.T) {
	tests := []struct {
		style Style
		want  string
	}{
		{StyleBrief, "yesterday"},
		{StyleDigest, "today"},
		{StyleReport, "today"},
		{StyleRetro, "today"},
		{StyleStatus, "today"},
	}
	for _, tt := range tests {
		if got := DefaultTimespec(tt.style); got != tt.want {
			t.Errorf("DefaultTimespec(%q) = %q, want %q", tt.style, got, tt.want)
		}
	}
}

func TestAllowedSources(t *testing.T) {
	tests := []struct {
		style   Style
		entries []sources.Entry
		wantN   int
	}{
		{
			StyleBrief,
			[]sources.Entry{
				{Source: "git"},
				{Source: "obsidian"},
				{Source: "claude"},
				{Source: "markdown"},
			},
			2, // git + claude
		},
		{
			StyleDigest,
			[]sources.Entry{
				{Source: "git"},
				{Source: "obsidian"},
				{Source: "claude"},
				{Source: "markdown"},
			},
			4, // no filter
		},
		{
			StyleReport,
			[]sources.Entry{
				{Source: "git"},
				{Source: "obsidian"},
				{Source: "claude"},
			},
			3, // no filter
		},
		{
			StyleRetro,
			[]sources.Entry{
				{Source: "git"},
				{Source: "obsidian"},
				{Source: "claude"},
			},
			3, // no filter
		},
		{
			StyleStatus,
			[]sources.Entry{
				{Source: "git"},
				{Source: "obsidian"},
				{Source: "claude"},
			},
			3, // no filter
		},
	}

	for _, tt := range tests {
		allowed := AllowedSources(tt.style)
		var filtered []sources.Entry
		if len(allowed) == 0 {
			filtered = tt.entries
		} else {
			for _, e := range tt.entries {
				for _, a := range allowed {
					if e.Source == a {
						filtered = append(filtered, e)
						break
					}
				}
			}
		}
		if len(filtered) != tt.wantN {
			t.Errorf("AllowedSources(%q): got %d entries, want %d", tt.style, len(filtered), tt.wantN)
		}
	}
}

func TestPromptNotEmpty(t *testing.T) {
	for _, style := range validStyles {
		p := Prompt(style)
		if p == "" {
			t.Errorf("Prompt(%q) returned empty string", style)
		}
	}
}

func TestPromptContainsLanguagePlaceholder(t *testing.T) {
	// All prompts must contain the {language} placeholder so PromptWithLanguage works.
	for _, style := range validStyles {
		p := Prompt(style)
		if !contains(p, "{language}") {
			t.Errorf("Prompt(%q) does not contain {language} placeholder", style)
		}
	}
}

func TestPromptWithLanguage(t *testing.T) {
	for _, style := range validStyles {
		p := PromptWithLanguage(style, "english")
		if contains(p, "{language}") {
			t.Errorf("PromptWithLanguage(%q, english) still contains {language} placeholder", style)
		}
		if !contains(p, "english") {
			t.Errorf("PromptWithLanguage(%q, english) does not contain injected language", style)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
