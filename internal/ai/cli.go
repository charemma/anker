package ai

import (
	"errors"
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

	bin := parts[0]

	// Check PATH before exec so we can give a clear error instead of "exit status 127".
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("AI command %q not found in PATH.\n\nCheck that the command is installed and accessible:\n  which %s\n\nOr update the command:\n  ikno config set ai_cli_command \"/full/path/to/%s -p\"", bin, bin, bin)
	}

	args := append(parts[1:], prompt)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = strings.NewReader(recapContent)
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 127 {
			return fmt.Errorf("AI command %q not found in PATH.\n\nCheck that the command is installed and accessible:\n  which %s\n\nOr update the command:\n  ikno config set ai_cli_command \"/full/path/to/%s -p\"", bin, bin, bin)
		}
		return fmt.Errorf("%s failed: %w", bin, err)
	}
	return nil
}
