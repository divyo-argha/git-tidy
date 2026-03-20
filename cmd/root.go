package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/git-tidy/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "git-tidy",
	Short: "Clean up messy Git history and prepare branches for pull requests",
	Long: tui.Header("git-tidy") + `

  Interactively select commits from your history and cherry-pick
  them into a clean branch — ready to open as a pull request.

  ` + tui.Bold("Commands:") + `
    ` + tui.Accent("create") + `    Select commits and create a clean branch
    ` + tui.Accent("preview") + `   Dry-run of create — no changes made
    ` + tui.Accent("log") + `       Show recent commits

  ` + tui.Bold("Quick start:") + `
    git tidy preview              # see what will happen first
    git tidy create my-branch     # then execute`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute is the entrypoint called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n  %s  %s\n\n", tui.IconErr(), err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newPreviewCmd())
	rootCmd.AddCommand(newLogCmd())
}
