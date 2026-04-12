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

func TestProjectNameFromDir(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-home-user-code-anker", "anker"},
		{"-home-user-code-my-project", "project"},
		{"-Users-charemma-code-monorepo", "monorepo"},
		{"standalone", "standalone"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := projectNameFromDir(tt.input)
			if got != tt.expected {
				t.Errorf("projectNameFromDir(%q) = %q, want %q", tt.input, got, tt.expected)
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
	lastWeek := now.Add(-7 * 24 * time.Hour)

	sessionFile := filepath.Join(claudeHome, "projects", "-home-user-code-anker", "session1.jsonl")
	appendSessionLine(t, sessionFile, userLine("implement the claude source", now.Add(-1*time.Hour), false))
	appendSessionLine(t, sessionFile, userLine("fix the test", now.Add(-30*time.Minute), false))
	appendSessionLine(t, sessionFile, userLine("old message outside range", lastWeek, false))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(yesterday, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
		for _, e := range entries {
			t.Logf("  - %s (%s)", e.Content, e.Timestamp)
		}
	}

	for _, entry := range entries {
		if entry.Source != "claude" {
			t.Errorf("expected source 'claude', got '%s'", entry.Source)
		}
		if entry.Location != claudeHome {
			t.Errorf("expected location '%s', got '%s'", claudeHome, entry.Location)
		}
		if entry.Metadata["project"] != "-home-user-code-anker" {
			t.Errorf("expected project '-home-user-code-anker', got '%s'", entry.Metadata["project"])
		}
		if entry.Metadata["session_file"] != "session1.jsonl" {
			t.Errorf("expected session_file 'session1.jsonl', got '%s'", entry.Metadata["session_file"])
		}
		// Content should be prefixed with project name
		if !strings.HasPrefix(entry.Content, "[anker] ") {
			t.Errorf("expected content to start with '[anker] ', got '%s'", entry.Content)
		}
	}
}

func TestClaudeSource_GetEntries_SkipsToolResults(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	toolResultLine := fmt.Sprintf(`{"type":"user","timestamp":"%s","sessionId":"s1","isMeta":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"abc","content":"file contents here"}]}}`,
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
		t.Errorf("expected 1 entry (tool_result should be skipped), got %d", len(entries))
	}
	if len(entries) > 0 && entries[0].Content != "[project1] real user message" {
		t.Errorf("expected '[project1] real user message', got '%s'", entries[0].Content)
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
		t.Errorf("expected 1 entry (meta should be skipped), got %d", len(entries))
	}
	if len(entries) > 0 && entries[0].Content != "[project1] real message" {
		t.Errorf("expected '[project1] real message', got '%s'", entries[0].Content)
	}
}

func TestClaudeSource_GetEntries_PreservesFullContent(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	longText := strings.Repeat("x", 500)

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

	expected := "[project1] " + longText
	if entries[0].Content != expected {
		t.Errorf("expected full content preserved (len %d), got len %d", len(expected), len(entries[0].Content))
	}
}

func TestClaudeSource_GetEntries_MultipleProjects(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "-home-user-code-foo", "-home-user-code-bar")
	now := time.Now().UTC()

	writeSessionLine(t,
		filepath.Join(claudeHome, "projects", "-home-user-code-foo", "session1.jsonl"),
		userLine("message from foo", now, false))
	writeSessionLine(t,
		filepath.Join(claudeHome, "projects", "-home-user-code-bar", "session1.jsonl"),
		userLine("message from bar", now.Add(1*time.Minute), false))

	source := NewClaudeSource(claudeHome)
	entries, err := source.GetEntries(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries from 2 projects, got %d", len(entries))
	}

	// Check that project metadata is set correctly
	projects := map[string]bool{}
	for _, e := range entries {
		projects[e.Metadata["project"]] = true
	}
	if !projects["-home-user-code-foo"] || !projects["-home-user-code-bar"] {
		t.Errorf("expected entries from both projects, got: %v", projects)
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
		t.Errorf("expected 1 entry (interrupt should be skipped), got %d", len(entries))
	}
}

func TestClaudeSource_GetEntries_LogsParseErrors(t *testing.T) {
	claudeHome := setupTestClaudeHome(t, "project1")
	now := time.Now().UTC()

	sessionFile := filepath.Join(claudeHome, "projects", "project1", "session1.jsonl")
	appendSessionLine(t, sessionFile, "this is not valid json")
	appendSessionLine(t, sessionFile, `{"type":"user","timestamp":"not-a-timestamp","sessionId":"s1","message":{"role":"user","content":"hello"}}`)
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
func userLine(text string, ts time.Time, isMeta bool) string {
	return userLineWithOpts(text, ts, isMeta, "test-session-id", "/tmp/test/project", "main")
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
