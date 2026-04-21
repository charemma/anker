package cmd

import "testing"

func TestResolveLanguage(t *testing.T) {
	tests := []struct {
		name          string
		flagValue     string
		configDefault string
		want          string
	}{
		{
			name:          "flag takes precedence",
			flagValue:     "english",
			configDefault: "deutsch",
			want:          "english",
		},
		{
			name:          "config used when flag empty",
			flagValue:     "",
			configDefault: "greek",
			want:          "greek",
		},
		{
			name:          "default when both empty",
			flagValue:     "",
			configDefault: "",
			want:          "deutsch",
		},
		{
			name:          "flag overrides empty config",
			flagValue:     "spanish",
			configDefault: "",
			want:          "spanish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveLanguage(tt.flagValue, tt.configDefault)
			if got != tt.want {
				t.Errorf("resolveLanguage(%q, %q) = %q, want %q",
					tt.flagValue, tt.configDefault, got, tt.want)
			}
		})
	}
}

func TestParseTemplateFile(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantDesc    string
		wantBodyPfx string // first non-empty line of body
	}{
		{
			name:        "with frontmatter and prompt heading",
			input:       "---\ndescription: My custom style\n---\n\n## Prompt\n\nWrite something.\n",
			wantDesc:    "My custom style",
			wantBodyPfx: "Write something.",
		},
		{
			name:        "with frontmatter no prompt heading",
			input:       "---\ndescription: Simple style\n---\n\nWrite something.\n",
			wantDesc:    "Simple style",
			wantBodyPfx: "Write something.",
		},
		{
			name:        "no frontmatter",
			input:       "Write something.\n",
			wantDesc:    "",
			wantBodyPfx: "Write something.",
		},
		{
			name:        "quoted description",
			input:       "---\ndescription: \"Quoted description\"\n---\n\nPrompt body.\n",
			wantDesc:    "Quoted description",
			wantBodyPfx: "Prompt body.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTemplateFile([]byte(tt.input))
			if got.Description != tt.wantDesc {
				t.Errorf("description = %q, want %q", got.Description, tt.wantDesc)
			}
			if tt.wantBodyPfx != "" && len(got.Body) < len(tt.wantBodyPfx) {
				t.Errorf("body too short: %q", got.Body)
			} else if tt.wantBodyPfx != "" && got.Body[:len(tt.wantBodyPfx)] != tt.wantBodyPfx {
				t.Errorf("body starts with %q, want %q", got.Body[:len(tt.wantBodyPfx)], tt.wantBodyPfx)
			}
		})
	}
}
