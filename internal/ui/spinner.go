package ui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// spinner frames from charmbracelet/bubbles dot style.
var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

// spinnerStyle colors the spinner character.
var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#82AAFF", Light: "#1A56DB"})

// StartSpinner starts an animated spinner on stderr with the given message.
// It returns a stop function that clears the spinner line.
// On non-TTY stderr the spinner is not shown and stop is a no-op.
func StartSpinner(message string) func() {
	if !isatty.IsTerminal(os.Stderr.Fd()) {
		return func() {}
	}

	var once sync.Once
	done := make(chan struct{})

	go func() {
		i := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			frame := spinnerStyle.Render(spinnerFrames[i%len(spinnerFrames)])
			line := fmt.Sprintf("\r%s %s", frame, message)
			_, _ = fmt.Fprint(os.Stderr, line)

			select {
			case <-done:
				// Clear the spinner line
				_, _ = fmt.Fprintf(os.Stderr, "\r%*s\r", len(message)+4, "")
				return
			case <-ticker.C:
				i++
			}
		}
	}()

	return func() {
		once.Do(func() {
			close(done)
			// Give the goroutine a moment to clear the line.
			time.Sleep(10 * time.Millisecond)
		})
	}
}
