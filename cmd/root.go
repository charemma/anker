package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "anker",
	Short: "anker - a fixpoint for your work",
	Long: `anker is a local, text-first CLI tool that helps you remember what you actually did
without time tracking, productivity metrics, or background agents.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
