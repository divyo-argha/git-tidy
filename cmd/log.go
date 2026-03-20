package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/git-tidy/internal/git"
	"github.com/yourusername/git-tidy/internal/tui"
)

func newLogCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Show recent commits on the current branch",
		Long: `Displays the most recent commits with hash, subject, author, and age.

Examples:
  git tidy log
  git tidy log -n 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			commits, err := git.Log(limit)
			if err != nil {
				tui.PrintError("Could not read commit history", err.Error())
				os.Exit(1)
			}
			tui.PrintBanner()
			branch, _ := git.CurrentBranch()
			if branch != "" {
				fmt.Printf("  %s  Branch: %s\n", tui.Info(tui.IconArrow), tui.Branch(branch))
			}
			tui.ShowCommits(commits, map[int]bool{})
			fmt.Printf("  %s  Showing %s  %s\n\n",
				tui.Muted("→"),
				tui.Accent(fmt.Sprintf("%d commits", len(commits))),
				tui.Muted("(use -n to show more)"),
			)
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of commits to show")
	return cmd
}
