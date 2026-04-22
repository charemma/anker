package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// httpClient is shared by all LLM API calls. Deadlines bound connection
// setup and the wait for response headers so a slow or hung endpoint
// cannot tie up the CLI indefinitely. Once headers arrive, the streaming
// body is unrestricted — legitimate long generations (e.g. a 7B model on
// CPU taking several minutes) complete normally.
//
// We clone http.DefaultTransport to preserve proxy support, system dialer
// settings, HTTP/2, and other defaults that a bare &http.Transport{} would lose.
var httpClient = func() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout: 10 * time.Second,
	}).DialContext
	transport.ResponseHeaderTimeout = 60 * time.Second
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.IdleConnTimeout = 30 * time.Second
	return &http.Client{Transport: transport}
}()

// newHTTPClientWithTimeout creates an HTTP client with a custom ResponseHeaderTimeout.
// It clones http.DefaultTransport to preserve proxy support and system settings.
func newHTTPClientWithTimeout(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout: 10 * time.Second,
	}).DialContext
	transport.ResponseHeaderTimeout = timeout
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.IdleConnTimeout = 30 * time.Second
	return &http.Client{Transport: transport}
}

type Client struct {
	BaseURL    string
	APIKey     string
	Model      string
	httpClient *http.Client // optional; if nil, uses the global default
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
func (c *Client) StreamCompletion(ctx context.Context, systemPrompt, userContent string, w io.Writer) error {
	if userContent == "" {
		return fmt.Errorf("no recap content to summarize")
	}

	req, err := c.BuildRequest(ctx, systemPrompt, userContent)
	if err != nil {
		return err
	}

	client := httpClient
	if c.httpClient != nil {
		client = c.httpClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return parseSSEStream(resp.Body, w)
}

// BuildRequest constructs the HTTP request without sending it.
func (c *Client) BuildRequest(ctx context.Context, systemPrompt, userContent string) (*http.Request, error) {
	if c.BaseURL == "" {
		return nil, fmt.Errorf("ai_base_url is not configured (set it in ~/.config/ikno/config.yaml or use a provider like Anthropic, OpenAI, or ollama)")
	}

	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"

	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		Stream: true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
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
			_, _ = fmt.Fprint(w, chunk.Choices[0].Delta.Content)
		}
	}

	_, _ = fmt.Fprintln(w)
	return scanner.Err()
}
