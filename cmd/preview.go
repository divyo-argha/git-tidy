package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/git-tidy/internal/git"
	"github.com/yourusername/git-tidy/internal/tui"
)

func newPreviewCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "preview [branch-name]",
		Short: "Preview what 'create' would do without making any changes",
		Long: `Shows the same interactive selection and execution plan as 'create',
but stops before creating any branch or applying any commits.

Safe to run at any time — your repository is never modified.

Examples:
  git tidy preview
  git tidy preview feature/my-branch
  git tidy preview --limit 30`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPreview(args, limit)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of commits to display")
	return cmd
}

func runPreview(args []string, limit int) error {
	tui.PrintBanner()
	tui.PrintDryRun()

	// ── Step 1: Preflight (read-only) ──────────────────────────────────────
	tui.PrintStep(1, "Checking repository state")

	if !git.IsRepo() {
		tui.PrintError("Not inside a git repository",
			"Run `git init` or navigate to an existing repository.")
		return fmt.Errorf("not a git repository")
	}

	originBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("could not determine current branch: %w", err)
	}

	// Warn about dirty state but don't block — preview is read-only.
	if cleanErr := git.IsClean(); cleanErr != nil {
		tui.PrintWarning("Working directory has uncommitted changes.")
		fmt.Printf("      %s\n", tui.Muted("(Fine for preview — 'create' will require a clean tree.)"))
	}

	fmt.Printf("  %s  On branch %s\n", tui.IconOK(), tui.Branch(originBranch))

	// ── Step 2: Resolve branch name ────────────────────────────────────────
	tui.PrintStep(2, "Target branch")

	branchName := fmt.Sprintf("tidy/%s", time.Now().Format("20060102-150405"))
	if len(args) == 1 && args[0] != "" {
		branchName = args[0]
	}

	exists := git.BranchExists(branchName)
	if exists {
		fmt.Printf("  %s  Branch: %s  %s\n",
			tui.IconWarn(), tui.Branch(branchName),
			tui.Warning("(already exists — 'create' would fail here)"))
	} else {
		fmt.Printf("  %s  Branch: %s\n", tui.IconOK(), tui.Branch(branchName))
		if len(args) == 0 {
			fmt.Printf("      %s\n", tui.Muted("(auto-generated — pass a name to override)"))
		}
	}

	// ── Step 3: Select commits ─────────────────────────────────────────────
	tui.PrintStep(3, "Select commits")

	commits, err := git.Log(limit)
	if err != nil {
		tui.PrintError("Could not read commit history", err.Error())
		return err
	}

	selected, err := tui.SelectCommits(commits)
	if err != nil {
		tui.PrintError("Commit selection failed", err.Error())
		return err
	}

	// ── Step 4: Show plan without prompting ────────────────────────────────
	tui.PrintStep(4, "Execution plan")

	_, _ = tui.ConfirmSelection(selected, branchName, true /*dryRun — skips prompt*/)

	// ── Summary ────────────────────────────────────────────────────────────
	fmt.Printf("  %s  Preview complete — no changes were made.\n", tui.IconOK())

	if exists {
		fmt.Printf("  %s  Branch %s already exists — pick a new name before running 'create'.\n\n",
			tui.IconWarn(), tui.Branch(branchName))
	} else {
		nextCmd := "git tidy create"
		if len(args) == 1 {
			nextCmd += " " + args[0]
		}
		fmt.Printf("\n  %s  Ready to execute? Run: %s\n\n",
			tui.Info(tui.IconArrow), tui.Accent(nextCmd))
	}

	return nil
}
