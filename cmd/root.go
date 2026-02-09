package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "anker",
	Short: "anker - a fixpoint for your work",
	Long: `anker is a local, text-first CLI tool that helps you remember what you actually did
without time tracking, productivity metrics, or background agents.`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("anker version %s (commit: %s, built: %s)\n", Version, Commit, Date))
}
