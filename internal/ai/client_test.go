package ai

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestBuildRequest_BasicFields(t *testing.T) {
	c := &Client{
		BaseURL: "https://api.example.com/v1/",
		APIKey:  "sk-test",
		Model:   "gpt-4",
	}

	req, err := c.BuildRequest("system prompt", "user content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.URL.String() != "https://api.example.com/v1/chat/completions" {
		t.Errorf("unexpected URL: %s", req.URL.String())
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("missing Content-Type header")
	}

	if req.Header.Get("Authorization") != "Bearer sk-test" {
		t.Errorf("unexpected Authorization header: %s", req.Header.Get("Authorization"))
	}

	body, _ := io.ReadAll(req.Body)
	var parsed chatRequest
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}

	if parsed.Model != "gpt-4" {
		t.Errorf("unexpected model: %s", parsed.Model)
	}
	if !parsed.Stream {
		t.Error("expected stream to be true")
	}
	if len(parsed.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(parsed.Messages))
	}
	if parsed.Messages[0].Role != "system" || parsed.Messages[0].Content != "system prompt" {
		t.Errorf("unexpected system message: %+v", parsed.Messages[0])
	}
	if parsed.Messages[1].Role != "user" || parsed.Messages[1].Content != "user content" {
		t.Errorf("unexpected user message: %+v", parsed.Messages[1])
	}
}

func TestBuildRequest_NoAPIKey(t *testing.T) {
	c := &Client{
		BaseURL: "http://localhost:11434/v1/",
		Model:   "llama3",
	}

	req, err := c.BuildRequest("prompt", "content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Header.Get("Authorization") != "" {
		t.Errorf("expected no Authorization header, got: %s", req.Header.Get("Authorization"))
	}
}

func TestBuildRequest_TrailingSlashNormalization(t *testing.T) {
	for _, baseURL := range []string{
		"https://api.example.com/v1",
		"https://api.example.com/v1/",
	} {
		c := &Client{BaseURL: baseURL, Model: "test"}
		req, err := c.BuildRequest("p", "c")
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", baseURL, err)
		}
		if req.URL.String() != "https://api.example.com/v1/chat/completions" {
			t.Errorf("for baseURL %q got URL %s", baseURL, req.URL.String())
		}
	}
}

func TestBuildRequest_MissingBaseURL(t *testing.T) {
	c := &Client{Model: "test"}
	_, err := c.BuildRequest("p", "c")
	if err == nil {
		t.Fatal("expected error for missing base URL")
	}
}

func TestStreamCompletion_EmptyContent(t *testing.T) {
	c := &Client{BaseURL: "http://localhost/v1", Model: "test"}
	err := c.StreamCompletion("prompt", "", io.Discard)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestStreamCompletion_MissingBaseURL(t *testing.T) {
	c := &Client{Model: "test"}
	err := c.StreamCompletion("prompt", "content", io.Discard)
	if err == nil {
		t.Fatal("expected error for missing base URL")
	}
}

func TestParseSSEStream(t *testing.T) {
	input := `data: {"choices":[{"delta":{"content":"Hello"}}]}
data: {"choices":[{"delta":{"content":" world"}}]}
data: [DONE]
`
	var buf strings.Builder
	err := parseSSEStream(strings.NewReader(input), &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", got)
	}
}

func TestParseSSEStream_SkipsNonDataLines(t *testing.T) {
	input := `: comment
event: message
data: {"choices":[{"delta":{"content":"ok"}}]}

data: [DONE]
`
	var buf strings.Builder
	err := parseSSEStream(strings.NewReader(input), &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "ok" {
		t.Errorf("expected 'ok', got %q", got)
	}
}
