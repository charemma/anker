package ai

import (
	"strings"
	"testing"
)

func TestRunCLI_EmptyCommand(t *testing.T) {
	err := RunCLI("", "prompt", "content", &strings.Builder{})
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	if !strings.Contains(err.Error(), "ai_cli_command is empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunCLI_CommandSplitting(t *testing.T) {
	var buf strings.Builder
	// Use echo which just prints its arguments, ignoring stdin
	err := RunCLI("echo hello", "world", "stdin content", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "hello world" {
		t.Errorf("expected 'hello world', got %q", got)
	}
}

func TestRunCLI_StdinPassthrough(t *testing.T) {
	var buf strings.Builder
	// sh -c runs the prompt as a shell command; "cat" reads stdin
	err := RunCLI("sh -c", "cat", "stdin content", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if got != "stdin content" {
		t.Errorf("expected 'stdin content', got %q", got)
	}
}

func TestRunCLI_NonexistentCommand(t *testing.T) {
	err := RunCLI("nonexistent-command-xyz", "prompt", "content", &strings.Builder{})
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}
