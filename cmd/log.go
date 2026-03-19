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
		RunE: func(cmd *cobra.Command, args []string) error {
			commits, err := git.Log(limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n  ✗  %s\n\n", err)
				os.Exit(1)
			}
			tui.ShowCommits(commits)
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of commits to show")
	return cmd
}
