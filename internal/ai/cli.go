package ai

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// RunCLI pipes recapContent via stdin to a CLI tool and connects its output to w.
// The command string is split on whitespace into program + args. The prompt is
// appended as the last argument.
func RunCLI(command, prompt, recapContent string, w io.Writer) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("ai_cli_command is empty")
	}

	args := append(parts[1:], prompt)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = strings.NewReader(recapContent)
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", parts[0], err)
	}
	return nil
}
