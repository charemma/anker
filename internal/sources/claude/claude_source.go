package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charemma/anker/internal/sources"
)

const maxContentLength = 200

// ClaudeSource implements the Source interface for Claude Code session data.
// It scans all project directories under <claudeHome>/projects/ for JSONL session files.
type ClaudeSource struct {
	claudeHome string // path to ~/.claude
}

// jsonlLine represents a single line from a Claude Code session JSONL file.
type jsonlLine struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"sessionId"`
	Slug      string          `json:"slug"`
	IsMeta    bool            `json:"isMeta"`
	Message   json.RawMessage `json:"message"`
}

// message represents the message field within a JSONL line.
type message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// contentBlock represents a typed content block within a message.
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewClaudeSource creates a new Claude Code session source.
// claudeHome is the path to the .claude directory (typically ~/.claude).
func NewClaudeSource(claudeHome string) *ClaudeSource {
	return &ClaudeSource{claudeHome: claudeHome}
}

// DefaultClaudeHome returns the default ~/.claude path.
func DefaultClaudeHome() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

func (c *ClaudeSource) Type() string {
	return "claude"
}

func (c *ClaudeSource) Location() string {
	return c.claudeHome
}

func (c *ClaudeSource) Validate() error {
	projectsDir := filepath.Join(c.claudeHome, "projects")
	info, err := os.Stat(projectsDir)
	if err != nil {
		return fmt.Errorf("projects directory not found: %s", projectsDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("projects path is not a directory: %s", projectsDir)
	}

	// Check for at least one project dir with .jsonl files
	dirEntries, err := os.ReadDir(projectsDir)
	if err != nil {
		return fmt.Errorf("failed to read projects directory: %w", err)
	}

	for _, d := range dirEntries {
		if !d.IsDir() {
			continue
		}
		matches, _ := filepath.Glob(filepath.Join(projectsDir, d.Name(), "*.jsonl"))
		if len(matches) > 0 {
			return nil
		}
	}

	return fmt.Errorf("no session files found in %s", projectsDir)
}

func (c *ClaudeSource) GetEntries(from, to time.Time) ([]sources.Entry, error) {
	projectsDir := filepath.Join(c.claudeHome, "projects")

	dirEntries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	var entries []sources.Entry

	for _, d := range dirEntries {
		if !d.IsDir() {
			continue
		}

		projectDir := d.Name()
		matches, err := filepath.Glob(filepath.Join(projectsDir, projectDir, "*.jsonl"))
		if err != nil {
			continue
		}

		for _, path := range matches {
			fileEntries, err := c.parseSessionFile(path, from, to, projectDir)
			if err != nil {
				continue
			}
			entries = append(entries, fileEntries...)
		}
	}

	return entries, nil
}

func (c *ClaudeSource) parseSessionFile(path string, from, to time.Time, project string) ([]sources.Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // 10MB max line size

	var entries []sources.Entry
	sessionFile := filepath.Base(path)

	for scanner.Scan() {
		line := scanner.Bytes()

		var jl jsonlLine
		if err := json.Unmarshal(line, &jl); err != nil {
			continue
		}

		if jl.Type != "user" {
			continue
		}

		if jl.IsMeta {
			continue
		}

		ts, err := time.Parse(time.RFC3339Nano, jl.Timestamp)
		if err != nil {
			continue
		}

		if ts.Before(from) || ts.After(to) {
			continue
		}

		var msg message
		if err := json.Unmarshal(jl.Message, &msg); err != nil {
			continue
		}

		if msg.Role != "user" {
			continue
		}

		text := extractUserText(msg.Content)
		if text == "" {
			continue
		}

		if isSystemMessage(text) {
			continue
		}

		if len(text) > maxContentLength {
			text = text[:maxContentLength] + "..."
		}

		// Prefix with project name for context
		projectName := projectNameFromDir(project)
		content := fmt.Sprintf("[%s] %s", projectName, text)

		entry := sources.Entry{
			Timestamp: ts,
			Source:    "claude",
			Location:  c.claudeHome,
			Content:   content,
			Metadata: map[string]string{
				"project":      project,
				"session_file": sessionFile,
			},
		}
		if jl.SessionID != "" {
			entry.Metadata["session_id"] = jl.SessionID
		}
		if jl.Slug != "" {
			entry.Metadata["slug"] = jl.Slug
		}

		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

// extractUserText extracts the text content from a message's content field.
// Content can be either a plain string or an array of content blocks.
func extractUserText(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}

	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}

	var parts []string
	for _, block := range blocks {
		if block.Type == "text" && block.Text != "" {
			parts = append(parts, strings.TrimSpace(block.Text))
		}
	}
	return strings.Join(parts, " ")
}

// projectNameFromDir extracts a human-readable project name from the encoded
// directory name. Takes the last segment, e.g. "-home-user-code-anker" -> "anker".
func projectNameFromDir(encoded string) string {
	encoded = strings.TrimRight(encoded, "-")
	if idx := strings.LastIndex(encoded, "-"); idx != -1 {
		return encoded[idx+1:]
	}
	return encoded
}

// isSystemMessage checks if the text is an auto-generated system message
// rather than actual user input.
func isSystemMessage(text string) bool {
	return strings.HasPrefix(text, "[Request interrupted")
}
