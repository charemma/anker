package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charemma/anker/internal/sources"
	"github.com/charemma/anker/internal/sources/aisession"
)

const maxContentPromptLength = 500

// ClaudeSource implements the Source interface for Claude Code session data.
// It scans all project directories under <claudeHome>/projects/ for JSONL session files.
type ClaudeSource struct {
	claudeHome string    // path to ~/.claude
	warn       io.Writer // warning output, defaults to os.Stderr
}

// jsonlLine represents a single line from a Claude Code session JSONL file.
type jsonlLine struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"sessionId"`
	Slug      string          `json:"slug"`
	IsMeta    bool            `json:"isMeta"`
	CWD       string          `json:"cwd"`
	GitBranch string          `json:"gitBranch"`
	Version   string          `json:"version"`
	Message   json.RawMessage `json:"message"`
}

// message represents the message field within a JSONL line.
type message struct {
	Role    string          `json:"role"`
	Model   string          `json:"model"`
	Content json.RawMessage `json:"content"`
}

// contentBlock represents a typed content block within a message.
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// toolUseBlock represents a tool_use content block from an assistant message.
type toolUseBlock struct {
	Type  string          `json:"type"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// toolInput captures file_path from tool inputs for files_touched tracking.
type toolInput struct {
	FilePath string `json:"file_path"`
}

// sessionData is the internal accumulator for building a session summary.
type sessionData struct {
	id          string
	slug        string
	firstPrompt string
	userTurns   int
	model       string
	cwd         string
	gitBranch   string
	startTime   time.Time
	endTime     time.Time
	toolSet     map[string]bool
	fileSet     map[string]bool
	projectDir  string
}

func newSessionData(id, projectDir string) *sessionData {
	return &sessionData{
		id:         id,
		projectDir: projectDir,
		toolSet:    make(map[string]bool),
		fileSet:    make(map[string]bool),
	}
}

func (s *sessionData) trackTimestamp(ts time.Time) {
	if s.startTime.IsZero() || ts.Before(s.startTime) {
		s.startTime = ts
	}
	if ts.After(s.endTime) {
		s.endTime = ts
	}
}

func (s *sessionData) toSummary() aisession.SessionSummary {
	tools := make([]aisession.ToolInvocation, 0, len(s.toolSet))
	for name := range s.toolSet {
		tools = append(tools, aisession.ToolInvocation{Name: name})
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })

	// Add file paths to the corresponding tool invocations where possible,
	// but also produce the flat list via fileSet
	return aisession.SessionSummary{
		SessionID:   s.id,
		Slug:        s.slug,
		Project:     projectNameFromCWD(s.cwd),
		ProjectDir:  s.projectDir,
		FirstPrompt: s.firstPrompt,
		TurnCount:   s.userTurns,
		Model:       s.model,
		CWD:         s.cwd,
		GitBranch:   s.gitBranch,
		StartTime:   s.startTime,
		EndTime:     s.endTime,
		ToolsUsed:   tools,
	}
}

// NewClaudeSource creates a new Claude Code session source.
// claudeHome is the path to the .claude directory (typically ~/.claude).
func NewClaudeSource(claudeHome string) *ClaudeSource {
	return &ClaudeSource{claudeHome: claudeHome, warn: os.Stderr}
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
			fmt.Fprintf(c.warn, "warning: %s: failed to glob session files: %v\n", projectDir, err)
			continue
		}

		for _, path := range matches {
			fileEntries, err := c.parseSessions(path, from, to, projectDir)
			if err != nil {
				fmt.Fprintf(c.warn, "warning: %s: %v\n", filepath.Base(path), err)
				continue
			}
			entries = append(entries, fileEntries...)
		}
	}

	return entries, nil
}

func (c *ClaudeSource) parseSessions(path string, from, to time.Time, projectDir string) ([]sources.Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	sessionFile := filepath.Base(path)
	sessions := make(map[string]*sessionData)
	lineNum := 0
	parseErrors := 0

	for scanner.Scan() {
		line := scanner.Bytes()
		lineNum++

		var jl jsonlLine
		if err := json.Unmarshal(line, &jl); err != nil {
			fmt.Fprintf(c.warn, "warning: %s: skipping line %d: %v\n", sessionFile, lineNum, err)
			parseErrors++
			continue
		}

		// Only process user and assistant lines
		if jl.Type != "user" && jl.Type != "assistant" {
			continue
		}

		ts, err := time.Parse(time.RFC3339Nano, jl.Timestamp)
		if err != nil {
			fmt.Fprintf(c.warn, "warning: %s: skipping line %d: invalid timestamp: %v\n", sessionFile, lineNum, err)
			parseErrors++
			continue
		}

		sid := jl.SessionID
		if sid == "" {
			sid = sessionFile
		}

		sess, ok := sessions[sid]
		if !ok {
			sess = newSessionData(sid, projectDir)
			sessions[sid] = sess
		}

		sess.trackTimestamp(ts)

		if sess.slug == "" && jl.Slug != "" {
			sess.slug = jl.Slug
		}
		if sess.cwd == "" && jl.CWD != "" {
			sess.cwd = jl.CWD
		}
		if sess.gitBranch == "" && jl.GitBranch != "" {
			sess.gitBranch = jl.GitBranch
		}

		var msg message
		if err := json.Unmarshal(jl.Message, &msg); err != nil {
			fmt.Fprintf(c.warn, "warning: %s: skipping line %d: invalid message: %v\n", sessionFile, lineNum, err)
			parseErrors++
			continue
		}

		switch jl.Type {
		case "user":
			c.processUserLine(sess, &jl, &msg)
		case "assistant":
			c.processAssistantLine(sess, &msg)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if parseErrors > 0 && len(sessions) == 0 {
		fmt.Fprintf(c.warn, "warning: %s: all %d parsed lines failed, file may be corrupted or incompatible\n", sessionFile, parseErrors)
	}

	// Convert sessions to entries, filtering by time range (start time in [from, to])
	var entries []sources.Entry
	for _, sess := range sessions {
		if sess.startTime.Before(from) || sess.startTime.After(to) {
			continue
		}
		if sess.userTurns == 0 {
			continue
		}
		summary := sess.toSummary()
		entries = append(entries, summaryToEntry(summary, sessionFile, c.claudeHome))
	}

	return entries, nil
}

func (c *ClaudeSource) processUserLine(sess *sessionData, jl *jsonlLine, msg *message) {
	if jl.IsMeta {
		return
	}

	text := extractUserText(msg.Content)
	if text == "" {
		return
	}

	if isSystemMessage(text) {
		return
	}

	sess.userTurns++

	// Capture first prompt, skipping slash commands
	if sess.firstPrompt == "" && !strings.HasPrefix(text, "/") {
		sess.firstPrompt = text
	}
}

func (c *ClaudeSource) processAssistantLine(sess *sessionData, msg *message) {
	// Capture model from the first assistant message
	if sess.model == "" && msg.Model != "" {
		sess.model = msg.Model
	}

	// Extract tool_use blocks
	invocations := extractToolUses(msg.Content)
	for _, inv := range invocations {
		sess.toolSet[inv.Name] = true
		if inv.Path != "" {
			sess.fileSet[inv.Path] = true
		}
	}
}

func summaryToEntry(s aisession.SessionSummary, sessionFile, claudeHome string) sources.Entry {
	// Build content: [project] prompt -- N turns, M min
	prompt := s.FirstPrompt
	truncated := false
	if len(prompt) > maxContentPromptLength {
		prompt = prompt[:maxContentPromptLength]
		truncated = true
	}

	var content string
	if truncated {
		content = fmt.Sprintf("[%s] %s... -- %d turns, %d min", s.Project, prompt, s.TurnCount, s.DurationMinutes())
	} else {
		content = fmt.Sprintf("[%s] %s -- %d turns, %d min", s.Project, prompt, s.TurnCount, s.DurationMinutes())
	}

	meta := map[string]string{
		"project":      s.ProjectDir,
		"session_file": sessionFile,
	}

	setIfNotEmpty(meta, "session_id", s.SessionID)
	setIfNotEmpty(meta, "slug", s.Slug)
	setIfNotEmpty(meta, "project_name", s.Project)
	setIfNotEmpty(meta, "cwd", s.CWD)
	setIfNotEmpty(meta, "git_branch", s.GitBranch)
	setIfNotEmpty(meta, "model", s.Model)
	setIfNotEmpty(meta, "first_prompt", s.FirstPrompt)

	if s.TurnCount > 0 {
		meta["turn_count"] = strconv.Itoa(s.TurnCount)
	}
	meta["duration_minutes"] = strconv.Itoa(s.DurationMinutes())

	if len(s.ToolsUsed) > 0 {
		names := make([]string, len(s.ToolsUsed))
		for i, t := range s.ToolsUsed {
			names[i] = t.Name
		}
		meta["tools_used"] = strings.Join(names, ",")
	}

	return sources.Entry{
		Timestamp: s.StartTime,
		Source:    "claude",
		Location:  claudeHome,
		Content:   content,
		Metadata:  meta,
	}
}

func setIfNotEmpty(m map[string]string, key, value string) {
	if value != "" {
		m[key] = value
	}
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

// extractToolUses parses tool_use blocks from an assistant message's content.
func extractToolUses(raw json.RawMessage) []aisession.ToolInvocation {
	var blocks []toolUseBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil
	}

	var invocations []aisession.ToolInvocation
	for _, block := range blocks {
		if block.Type != "tool_use" || block.Name == "" {
			continue
		}
		inv := aisession.ToolInvocation{Name: block.Name}
		var ti toolInput
		if err := json.Unmarshal(block.Input, &ti); err == nil && ti.FilePath != "" {
			inv.Path = ti.FilePath
		}
		invocations = append(invocations, inv)
	}
	return invocations
}

// projectNameFromCWD extracts a human-readable project name from the working
// directory path found in JSONL session data.
func projectNameFromCWD(cwd string) string {
	if cwd == "" {
		return "unknown"
	}
	return filepath.Base(cwd)
}

// isSystemMessage checks if the text is an auto-generated system message
// rather than actual user input.
func isSystemMessage(text string) bool {
	return strings.HasPrefix(text, "[Request interrupted")
}
