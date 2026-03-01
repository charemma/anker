package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	BaseURL string
	APIKey  string
	Model   string
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// StreamCompletion sends a chat completion request and streams the response to w.
func (c *Client) StreamCompletion(systemPrompt, userContent string, w io.Writer) error {
	if c.BaseURL == "" {
		return fmt.Errorf("ai_base_url is not configured (set it in ~/.anker/config.yaml or use a provider like Anthropic, OpenAI, or ollama)")
	}
	if userContent == "" {
		return fmt.Errorf("no recap content to summarize")
	}

	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"

	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent},
	}

	reqBody := chatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return parseSSEStream(resp.Body, w)
}

// BuildRequest constructs the HTTP request without sending it. Exported for testing.
func (c *Client) BuildRequest(systemPrompt, userContent string) (*http.Request, error) {
	if c.BaseURL == "" {
		return nil, fmt.Errorf("ai_base_url is not configured")
	}

	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"

	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent},
	}

	reqBody := chatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	return req, nil
}

func parseSSEStream(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			fmt.Fprint(w, chunk.Choices[0].Delta.Content)
		}
	}

	fmt.Fprintln(w)
	return scanner.Err()
}
