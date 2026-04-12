package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProjectNameFromCWD(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/home/user/code/anker", "anker"},
		{"/home/user/code/my-project", "my-project"},
		{"/Users/charemma/code/nixos-config", "nixos-config"},
		{"/home/charemma/Documents/Notes/1 Projects/anker-evolution-strategy", "anker-evolution-strategy"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := projectNameFromCWD(tt.input)
			if got != tt.expected {
				t.Errorf("projectNameFromCWD(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestClaudeSource_Validate(t *testing.T) {
	t.Run("valid directory with session files", func(t *testing.T) {
		claudeHome := setupTestClaudeHome(t, "myproject")
		writeSessionLine(t,
			filepath.Join(claudeHome, "projects", "myproject", "session1.jsonl"),
			userLine("hello", time.Now(), false))

		source := NewClaudeSource(claudeHome)
		if err := source.Validate(); err != nil {
			t.Errorf("expected validation to pass, got: %v", err)
		}
	})

	t.Run("missing directory", func(t *testing.T) {
		source := NewClaudeSource("/nonexistent/.claude")
		if err := source.Validate(); err == nil {
			t.Error("expected validation to fail for missing directory")
		}
	})

	t.Run("empty projects directory", func(t *testing.T) {
		claudeHome := t.TempDir()
		if err := os.MkdirAll(filepath.Join(claudeHome, "projects"), 0755); err != nil {
			t.Fatal(err)
		}

		source := NewClaudeSource(claudeHome)
		if err := source.Validate(); err == nil {
			t.Error("expected validation to fail for empty projects directory")
		}
	})
}

func TestClaudeSource_GetEntries(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "-home-user-code-anker")

	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)
	cwd := "/home/user/code/anker"
	sid := "session-1"

	sessionFile := filepath.Join(claudeHome, "projects", "-home-user-code-anker", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLineWithOpts("implement the claude source", now.Add(-1*time.Hour), false, sid, cwd, "main"))
	appendSessionLine(t, sessionFile, assistantLine(now.Add(-55*time.Minute), sid, cwd, "main", "claude-opus-4-6", nil))
	appendSessionLine(t, sessionFile, userLineWithOpts("fix the test", now.Add(-30*time.Minute), false, sid, cwd, "main"))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(yesterday, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	// Session-level: one entry per session
	if len(entries) != 1 {
		t.Fatalf("expected 1 session entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Source != "claude" {
		t.Errorf("expected source 'claude', got '%s'", entry.Source)
	}
	if entry.Location != claudeHome {
		t.Errorf("expected location '%s', got '%s'", claudeHome, entry.Location)
	}
	if entry.Metadata["project"] != "-home-user-code-anker" {
		t.Errorf("expected project '-home-user-code-anker', got '%s'", entry.Metadata["project"])
	}
	if entry.Metadata["project_name"] != "anker" {
		t.Errorf("expected project_name 'anker', got '%s'", entry.Metadata["project_name"])
	}
	if !strings.HasPrefix(entry.Content, "[anker] implement the claude source") {
		t.Errorf("expected content to start with '[anker] implement the claude source', got '%s'", entry.Content)
	}
	if entry.Metadata["turn_count"] != "2" {
		t.Errorf("expected turn_count '2', got '%s'", entry.Metadata["turn_count"])
	}
}

func TestClaudeSource_GetEntries_SessionOutsideRange(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project")

	now := time.Now().UTC()
	lastWeek := now.Add(-7 * 24 * time.Hour)

	sessionFile := filepath.Join(claudeHome, "projects", "project", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLine("old message", lastWeek, false))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-24*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries for out-of-range session, got %d", len(entries))
	}
}

func TestClaudeSource_GetEntries_SkipsToolResults(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	toolResultLine := fmt.Sprintf(`{"type":"user","timestamp":"%s","sessionId":"s1","isMeta":false,"cwd":"/tmp/test/project1","gitBranch":"main","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"abc","content":"file contents here"}]}}`,
		now.Format(time.RFC3339Nano))

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLine("real user message", now, false))
	appendSessionLine(t, sessionFile, toolResultLine)

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 session entry, got %d", len(entries))
	}

	// tool_result lines don't count as user turns
	if entries[0].Metadata["turn_count"] != "1" {
		t.Errorf("expected turn_count '1' (tool_result excluded), got '%s'", entries[0].Metadata["turn_count"])
	}
}

func TestClaudeSource_GetEntries_SkipsMetaMessages(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLine("real message", now, false))
	appendSessionLine(t, sessionFile, userLine("meta message", now.Add(1*time.Minute), true))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 session entry, got %d", len(entries))
	}
	if entries[0].Metadata["turn_count"] != "1" {
		t.Errorf("expected turn_count '1' (meta excluded), got '%s'", entries[0].Metadata["turn_count"])
	}
}

func TestClaudeSource_GetEntries_SkipsSystemInterrupts(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLine("real message", now, false))
	appendSessionLine(t, sessionFile, userLine("[Request interrupted by user]", now.Add(1*time.Minute), false))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 session entry, got %d", len(entries))
	}
	// System interrupt should not become the first prompt and shouldn't count as a turn
	if entries[0].Metadata["turn_count"] != "1" {
		t.Errorf("expected turn_count '1', got '%s'", entries[0].Metadata["turn_count"])
	}
	if !strings.Contains(entries[0].Content, "real message") {
		t.Errorf("expected first prompt to be 'real message', got '%s'", entries[0].Content)
	}
}

func TestClaudeSource_GetEntries_MultipleProjects(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "-home-user-code-foo", "-home-user-code-bar")
	now := time.Now().UTC()

	writeSessionLine(t,
		filepath.Join(claudeHome, "projects", "-home-user-code-foo", "session1.jsonl"),
		userLineWithOpts("message from foo", now, false, "s-foo", "/home/user/code/foo", "main"))
	writeSessionLine(t,
		filepath.Join(claudeHome, "projects", "-home-user-code-bar", "session1.jsonl"),
		userLineWithOpts("message from bar", now.Add(1*time.Minute), false, "s-bar", "/home/user/code/bar", "main"))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries from 2 projects, got %d", len(entries))
	}

	projects := map[string]bool{}
	for _, e := range entries {
		projects[e.Metadata["project"]] = true
	}
	if !projects["-home-user-code-foo"] || !projects["-home-user-code-bar"] {
		t.Errorf("expected entries from both projects, got: %v", projects)
	}
}

func TestClaudeSource_GetEntries_SessionMetadata(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "-home-user-code-anker")
	now := time.Now().UTC()
	sid := "abc-123"
	cwd := "/home/user/code/anker"

	sessionFile := filepath.Join(claudeHome, "projects", "-home-user-code-anker", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLineWithOpts("implement feature", now, false, sid, cwd, "feat/test"))
	appendSessionLine(t, sessionFile, assistantLine(now.Add(5*time.Second), sid, cwd, "feat/test", "claude-opus-4-6",
		[]map[string]any{
			{"type": "tool_use", "id": "t1", "name": "Read", "input": map[string]any{"file_path": "/home/user/code/anker/main.go"}},
		}))
	appendSessionLine(t, sessionFile, userLineWithOpts("looks good", now.Add(10*time.Minute), false, sid, cwd, "feat/test"))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	meta := entries[0].Metadata
	checks := map[string]string{
		"session_id":       sid,
		"slug":             "test-slug",
		"project":          "-home-user-code-anker",
		"project_name":     "anker",
		"cwd":              cwd,
		"git_branch":       "feat/test",
		"model":            "claude-opus-4-6",
		"turn_count":       "2",
		"duration_minutes": "10",
		"tools_used":       "Read",
		"first_prompt":     "implement feature",
	}

	for key, want := range checks {
		got := meta[key]
		if got != want {
			t.Errorf("metadata[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestClaudeSource_GetEntries_ToolUseExtraction(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()
	sid := "tool-session"

	tools := []map[string]any{
		{"type": "tool_use", "id": "t1", "name": "Read", "input": map[string]any{"file_path": "/code/main.go"}},
		{"type": "tool_use", "id": "t2", "name": "Edit", "input": map[string]any{"file_path": "/code/main.go"}},
		{"type": "tool_use", "id": "t3", "name": "Bash", "input": map[string]any{"command": "go test ./..."}},
		{"type": "text", "text": "some explanation"},
	}

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLineWithOpts("do something", now, false, sid, "/tmp/test/project1", "main"))
	appendSessionLine(t, sessionFile, assistantLine(now.Add(5*time.Second), sid, "/tmp/test/project1", "main", "claude-opus-4-6", tools))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	toolsUsed := entries[0].Metadata["tools_used"]
	// Should be sorted alphabetically and deduplicated
	if toolsUsed != "Bash,Edit,Read" {
		t.Errorf("expected tools_used 'Bash,Edit,Read', got '%s'", toolsUsed)
	}
}

func TestClaudeSource_GetEntries_SlashCommandSkip(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	// First message is a slash command (from isMeta=false but starts with /)
	appendSessionLine(t, sessionFile, userLine("/init", now, false))
	appendSessionLine(t, sessionFile, userLine("implement the feature", now.Add(1*time.Minute), false))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if !strings.Contains(entries[0].Content, "implement the feature") {
		t.Errorf("expected first prompt to skip /init, got '%s'", entries[0].Content)
	}
	// Both messages count as turns
	if entries[0].Metadata["turn_count"] != "2" {
		t.Errorf("expected turn_count '2', got '%s'", entries[0].Metadata["turn_count"])
	}
}

func TestClaudeSource_GetEntries_FirstPromptCap(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	longText := strings.Repeat("x", 700)

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLine(longText, now, false))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	// Content should be capped at 500 chars with "..." before the suffix
	if !strings.Contains(entries[0].Content, "...") {
		t.Error("expected truncation indicator '...' in content")
	}

	// Full prompt should be in metadata
	if entries[0].Metadata["first_prompt"] != longText {
		t.Errorf("expected full prompt in metadata, got len %d", len(entries[0].Metadata["first_prompt"]))
	}
}

func TestClaudeSource_GetEntries_SessionDuration(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	start := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLineWithOpts("start", start, false, "s1", "/tmp/test/project1", "main"))
	appendSessionLine(t, sessionFile, userLineWithOpts("end", start.Add(47*time.Minute+30*time.Second), false, "s1", "/tmp/test/project1", "main"))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(start.Add(-1*time.Hour), start.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	// 47m30s rounds to 48
	if entries[0].Metadata["duration_minutes"] != "48" {
		t.Errorf("expected duration_minutes '48', got '%s'", entries[0].Metadata["duration_minutes"])
	}
	if !strings.Contains(entries[0].Content, "48 min") {
		t.Errorf("expected '48 min' in content, got '%s'", entries[0].Content)
	}
}

func TestClaudeSource_GetEntries_LogsParseErrors(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, "this is not valid json")
	appendSessionLine(t, sessionFile, `{"type":"user","timestamp":"not-a-timestamp","sessionId":"s1","cwd":"/x","message":{"role":"user","content":"hello"}}`)
	appendSessionLine(t, sessionFile, userLine("valid message", now, false))

	var buf bytes.Buffer
	source := NewClaudeSource(claudeHome)
	source.warn = &buf

	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 valid entry, got %d", len(entries))
	}

	warnings := buf.String()
	if !strings.Contains(warnings, "skipping line 1") {
		t.Errorf("expected warning for line 1 (bad json), got: %s", warnings)
	}
	if !strings.Contains(warnings, "skipping line 2") {
		t.Errorf("expected warning for line 2 (bad timestamp), got: %s", warnings)
	}
}

func TestClaudeSource_GetEntries_PreservesFullContent(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	// Text under 500 chars should not be truncated in Content
	text := strings.Repeat("x", 400)

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLine(text, now, false))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if strings.Contains(entries[0].Content, "...") {
		t.Error("expected no truncation for text under 500 chars")
	}
	if !strings.Contains(entries[0].Content, text) {
		t.Error("expected full text in content")
	}
}

func TestClaudeSource_GetEntries_EmptySession(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	// Session with only meta and system messages -- no real user turns
	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLine("meta only", now, true))
	appendSessionLine(t, sessionFile, userLine("[Request interrupted by user]", now.Add(1*time.Minute), false))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries for session with no real turns, got %d", len(entries))
	}
}

func TestClaudeSource_GetEntries_NoAssistantMessages(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLine("interrupted before response", now, false))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	// No model should be set
	if entries[0].Metadata["model"] != "" {
		t.Errorf("expected empty model, got '%s'", entries[0].Metadata["model"])
	}
	if entries[0].Metadata["turn_count"] != "1" {
		t.Errorf("expected turn_count '1', got '%s'", entries[0].Metadata["turn_count"])
	}
}

func TestClaudeSource_GetEntries_MultiSessionFile(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	// Two different sessions in the same file (defensive test)
	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLineWithOpts("session A", now, false, "session-a", "/tmp/test/project1", "main"))
	appendSessionLine(t, sessionFile, userLineWithOpts("session B", now.Add(1*time.Minute), false, "session-b", "/tmp/test/project1", "main"))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries from 2 sessions, got %d", len(entries))
	}

	sessionIDs := map[string]bool{}
	for _, e := range entries {
		sessionIDs[e.Metadata["session_id"]] = true
	}
	if !sessionIDs["session-a"] || !sessionIDs["session-b"] {
		t.Errorf("expected both sessions, got: %v", sessionIDs)
	}
}

func TestClaudeSource_GetEntries_AllLinesFail(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, "not json 1")
	appendSessionLine(t, sessionFile, "not json 2")
	appendSessionLine(t, sessionFile, "not json 3")

	var buf bytes.Buffer
	source := NewClaudeSource(claudeHome)
	source.warn = &buf

	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}

	warnings := buf.String()
	if !strings.Contains(warnings, "all 3 parsed lines failed") {
		t.Errorf("expected summary warning about all lines failing, got: %s", warnings)
	}
}

func TestExtractUserText(t *testing.T) {
	t.Run("string content", func(t *testing.T) {
		raw := json.RawMessage(`"hello world"`)
		got := extractUserText(raw)
		if got != "hello world" {
			t.Errorf("expected 'hello world', got '%s'", got)
		}
	})

	t.Run("array with text blocks", func(t *testing.T) {
		raw := json.RawMessage(`[{"type":"text","text":"implement this"},{"type":"text","text":"please"}]`)
		got := extractUserText(raw)
		if got != "implement this please" {
			t.Errorf("expected 'implement this please', got '%s'", got)
		}
	})

	t.Run("array with mixed blocks", func(t *testing.T) {
		raw := json.RawMessage(`[{"type":"text","text":"check this"},{"type":"tool_result","tool_use_id":"abc","content":"file data"}]`)
		got := extractUserText(raw)
		if got != "check this" {
			t.Errorf("expected 'check this', got '%s'", got)
		}
	})

	t.Run("array with only tool_result blocks", func(t *testing.T) {
		raw := json.RawMessage(`[{"type":"tool_result","tool_use_id":"abc","content":"data"}]`)
		got := extractUserText(raw)
		if got != "" {
			t.Errorf("expected empty string, got '%s'", got)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		raw := json.RawMessage(`""`)
		got := extractUserText(raw)
		if got != "" {
			t.Errorf("expected empty string, got '%s'", got)
		}
	})
}

func TestExtractToolUses(t *testing.T) {
	t.Run("tool_use blocks with file paths", func(t *testing.T) {
		raw := json.RawMessage(`[{"type":"tool_use","name":"Read","input":{"file_path":"/code/main.go"}},{"type":"tool_use","name":"Bash","input":{"command":"ls"}},{"type":"text","text":"explanation"}]`)
		got := extractToolUses(raw)
		if len(got) != 2 {
			t.Fatalf("expected 2 tool invocations, got %d", len(got))
		}
		if got[0].Name != "Read" || got[0].Path != "/code/main.go" {
			t.Errorf("expected Read with path, got %+v", got[0])
		}
		if got[1].Name != "Bash" || got[1].Path != "" {
			t.Errorf("expected Bash without path, got %+v", got[1])
		}
	})

	t.Run("no tool_use blocks", func(t *testing.T) {
		raw := json.RawMessage(`[{"type":"text","text":"just text"}]`)
		got := extractToolUses(raw)
		if len(got) != 0 {
			t.Errorf("expected 0 invocations, got %d", len(got))
		}
	})

	t.Run("string content", func(t *testing.T) {
		raw := json.RawMessage(`"just a string"`)
		got := extractToolUses(raw)
		if len(got) != 0 {
			t.Errorf("expected 0 invocations for string content, got %d", len(got))
		}
	})
}

// --- helpers ---

// setupTestClaudeHome creates a temp directory structure mimicking ~/.claude/projects/<dirs>/
func setupTestClaudeHome(t *testing.T, projectDirs ...string) string {
	t.Helper()
	claudeHome := filepath.Join(t.TempDir(), ".claude")
	for _, dir := range projectDirs {
		if err := os.MkdirAll(filepath.Join(claudeHome, "projects", dir), 0755); err != nil {
			t.Fatalf("failed to create project dir: %v", err)
		}
	}
	return claudeHome
}

// userLine creates a JSONL line for a user message with text content blocks.
// Default cwd is "/tmp/test/project1" so projectNameFromCWD returns "project1".
func userLine(text string, ts time.Time, isMeta bool) string {
	return userLineWithOpts(text, ts, isMeta, "test-session-id", "/tmp/test/project1", "main")
}

// userLineWithOpts creates a JSONL user line with explicit session/cwd/branch values.
func userLineWithOpts(text string, ts time.Time, isMeta bool, sessionID, cwd, branch string) string {
	line := map[string]any{
		"type":      "user",
		"timestamp": ts.Format(time.RFC3339Nano),
		"sessionId": sessionID,
		"slug":      "test-slug",
		"isMeta":    isMeta,
		"cwd":       cwd,
		"gitBranch": branch,
		"message": map[string]any{
			"role": "user",
			"content": []map[string]any{
				{"type": "text", "text": text},
			},
		},
	}
	data, _ := json.Marshal(line)
	return string(data)
}

// assistantLine creates a JSONL line for an assistant message with optional tool_use content blocks.
func assistantLine(ts time.Time, sessionID, cwd, branch, model string, toolBlocks []map[string]any) string {
	var content []map[string]any
	if len(toolBlocks) > 0 {
		content = toolBlocks
	} else {
		content = []map[string]any{
			{"type": "text", "text": "assistant response"},
		}
	}

	line := map[string]any{
		"type":      "assistant",
		"timestamp": ts.Format(time.RFC3339Nano),
		"sessionId": sessionID,
		"cwd":       cwd,
		"gitBranch": branch,
		"message": map[string]any{
			"role":    "assistant",
			"model":   model,
			"content": content,
		},
	}
	data, _ := json.Marshal(line)
	return string(data)
}

// writeSessionLine writes a single line to a new session file.
func writeSessionLine(t *testing.T, path, line string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(line+"\n"), 0644); err != nil {
		t.Fatalf("failed to write session file %s: %v", path, err)
	}
}

// appendSessionLine appends a line to a session file (creates if needed).
func appendSessionLine(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open session file %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(line + "\n"); err != nil {
		t.Fatalf("failed to write to session file %s: %v", path, err)
	}
}
