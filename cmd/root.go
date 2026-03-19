package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-tidy",
	Short: "Clean up messy Git history and prepare branches for pull requests",
	Long: `git-tidy helps you create clean branches by cherry-picking selected commits.

Designed to run as a git subcommand: git tidy <command>`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute is the entrypoint called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n  %s %s\n\n", "✗", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newLogCmd())
}
