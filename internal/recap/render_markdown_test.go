package recap

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/charemma/ikno/internal/sources"
	"github.com/charemma/ikno/internal/timerange"
)

func TestRenderForAI_CondensesClaudeSessions(t *testing.T) {
	now := time.Now()
	tr := &timerange.TimeRange{From: now.AddDate(0, 0, -7), To: now}

	longPrompt := strings.Repeat("this is a verbose prompt about refactoring ", 20)

	result := &RecapResult{
		TimeRange: tr,
		Timespec:  "lastweek",
		Entries: []sources.Entry{
			{
				Timestamp: now.Add(-2 * time.Hour),
				Source:    "claude",
				Location:  "/home/user/.claude",
				Content:   "[ikno] " + longPrompt + " -- 15 turns, 45 min",
				Metadata: map[string]string{
					"project_name":     "ikno",
					"turn_count":       "15",
					"duration_minutes": "45",
					"git_branch":       "feat/refactor",
					"first_prompt":     longPrompt,
					"model":            "claude-sonnet-4-20250514",
					"tools_used":       "Read,Edit,Bash",
				},
			},
			{
				Timestamp: now.Add(-1 * time.Hour),
				Source:    "claude",
				Location:  "/home/user/.claude",
				Content:   "[dotfiles] check zshrc config -- 2 turns, 5 min",
				Metadata: map[string]string{
					"project_name":     "dotfiles",
					"turn_count":       "2",
					"duration_minutes": "5",
					"first_prompt":     "check zshrc config",
					"model":            "claude-sonnet-4-20250514",
				},
			},
			{
				Timestamp: now.Add(-3 * time.Hour),
				Source:    "git",
				Location:  "/home/user/code/ikno",
				Content:   "feat: add recap command",
				Metadata: map[string]string{
					"author": "user",
					"hash":   "abc1234",
					"diff":   "+func Recap() {}\n-// todo",
				},
			},
		},
	}

	// Render with the old (full) method
	var fullBuf bytes.Buffer
	if err := renderMarkdownRaw(&fullBuf, result); err != nil {
		t.Fatalf("renderMarkdownRaw: %v", err)
	}

	// Render with the AI-condensed method
	var aiBuf bytes.Buffer
	if err := RenderForAI(&aiBuf, result); err != nil {
		t.Fatalf("RenderForAI: %v", err)
	}

	fullSize := fullBuf.Len()
	aiSize := aiBuf.Len()

	t.Logf("Full render: %d bytes", fullSize)
	t.Logf("AI render:   %d bytes", aiSize)
	t.Logf("Reduction:   %.0f%%", float64(fullSize-aiSize)/float64(fullSize)*100)

	// AI output should be smaller
	if aiSize >= fullSize {
		t.Errorf("AI render (%d bytes) should be smaller than full render (%d bytes)", aiSize, fullSize)
	}

	aiText := aiBuf.String()

	// Should contain condensed Claude section header
	if !strings.Contains(aiText, "Claude Sessions (2 sessions)") {
		t.Error("missing condensed Claude section header")
	}

	// Should contain truncated prompt (not the full 800+ char prompt)
	if strings.Contains(aiText, longPrompt) {
		t.Error("AI render should truncate long prompts")
	}

	// Git entries should still render fully
	if !strings.Contains(aiText, "```diff") {
		t.Error("git entries should still have diffs in AI render")
	}
	if !strings.Contains(aiText, "feat: add recap command") {
		t.Error("git entries should still have full content in AI render")
	}
}

func TestRenderMarkdownRaw_Unchanged(t *testing.T) {
	now := time.Now()
	tr := &timerange.TimeRange{From: now.AddDate(0, 0, -7), To: now}

	result := &RecapResult{
		TimeRange: tr,
		Timespec:  "lastweek",
		Entries: []sources.Entry{
			{
				Timestamp: now,
				Source:    "claude",
				Location:  "/home/user/.claude",
				Content:   "[ikno] do something -- 5 turns, 10 min",
				Metadata: map[string]string{
					"first_prompt": "do something",
					"turn_count":   "5",
				},
			},
			{
				Timestamp: now,
				Source:    "git",
				Location:  "/home/user/code/repo",
				Content:   "fix: something",
				Metadata: map[string]string{
					"hash": "abc",
					"diff": "+line",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := renderMarkdownRaw(&buf, result); err != nil {
		t.Fatalf("renderMarkdownRaw: %v", err)
	}

	text := buf.String()

	// renderMarkdownRaw should still use the full format with numbered entries
	if !strings.Contains(text, "### 1.") {
		t.Error("renderMarkdownRaw should have numbered entries")
	}
	if !strings.Contains(text, "**Message:** [ikno] do something") {
		t.Error("renderMarkdownRaw should have full message content")
	}
	if !strings.Contains(text, "*(No diff available)*") {
		t.Error("renderMarkdownRaw should show 'No diff available' for claude entries")
	}
}

func TestTruncatePrompt(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"empty", "", 100, ""},
		{"short", "hello world", 100, "hello world"},
		{"exact", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello..."},
		{"newlines", "line1\nline2\nline3", 100, "line1 line2 line3"},
		{"multi_spaces", "hello   world", 100, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePrompt(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncatePrompt(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
