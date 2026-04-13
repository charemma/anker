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
	Use:   "ikno",
	Short: "ikno - a fixpoint for your work",
	Long: `ikno is a local, text-first CLI tool that helps you remember what you actually did
without time tracking, productivity metrics, or background agents.`,
	Version: Version,
}

func Execute() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(fmt.Sprintf("ikno version %s (commit: %s, built: %s)\n", Version, Commit, Date))
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
