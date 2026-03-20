package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/git-tidy/internal/domain"
	"github.com/yourusername/git-tidy/internal/git"
	"github.com/yourusername/git-tidy/internal/tui"
	giterrors "github.com/yourusername/git-tidy/pkg/errors"
)

func newCreateCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "create [branch-name]",
		Short: "Select commits interactively and cherry-pick them into a new branch",
		Long: `Shows recent commits, lets you choose which to keep, then creates
a clean branch with only those commits applied.

If no branch name is provided, one is auto-generated: tidy/<timestamp>

Examples:
  git tidy create
  git tidy create feature/auth-cleanup
  git tidy create fix/payments --limit 30`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(args, limit, false)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of commits to display")
	return cmd
}

func runCreate(args []string, limit int, dryRun bool) error {
	tui.PrintBanner()

	// ── Step 1: Preflight ──────────────────────────────────────────────────
	tui.PrintStep(1, "Checking repository state")

	if !git.IsRepo() {
		tui.PrintError("Not inside a git repository",
			"Run `git init` or navigate to an existing repository.")
		return fmt.Errorf("not a git repository")
	}

	if err := git.IsClean(); err != nil {
		tui.PrintError("Working directory has uncommitted changes",
			"Stash or commit your changes before running git-tidy.",
			"  git stash        — temporarily save changes",
			"  git commit -am   — commit everything")
		return err
	}

	originBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("could not determine current branch: %w", err)
	}
	fmt.Printf("  %s  On branch %s — working tree clean\n",
		tui.IconOK(), tui.Branch(originBranch))

	// ── Step 2: Resolve branch name ────────────────────────────────────────
	tui.PrintStep(2, "Target branch")

	branchName := fmt.Sprintf("tidy/%s", time.Now().Format("20060102-150405"))
	if len(args) == 1 && args[0] != "" {
		branchName = args[0]
	}

	if err := git.ValidateBranchName(branchName); err != nil {
		tui.PrintError(fmt.Sprintf("Invalid branch name: %q", branchName),
			"Branch names cannot contain spaces or special characters.")
		return err
	}

	if git.BranchExists(branchName) {
		tui.PrintError(fmt.Sprintf("Branch %q already exists", branchName),
			"Choose a different name or delete the existing branch first.")
		return &giterrors.BranchExistsError{Name: branchName}
	}

	fmt.Printf("  %s  New branch: %s\n", tui.IconOK(), tui.Branch(branchName))
	if len(args) == 0 {
		fmt.Printf("      %s\n", tui.Muted("(auto-generated — pass a name to override)"))
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

	// ── Step 4: Review & confirm ───────────────────────────────────────────
	tui.PrintStep(4, "Review & confirm")

	confirmed, err := tui.ConfirmSelection(selected, branchName, dryRun)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("  %s  Dry run complete — nothing was changed.\n", tui.IconWarn())
		fmt.Printf("      %s  Run %s to execute.\n\n",
			tui.Muted("→"), tui.Accent("git tidy create "+branchName))
		return nil
	}

	if !confirmed {
		tui.Spacer()
		fmt.Printf("  %s  Aborted — no changes made.\n\n", tui.IconWarn())
		return nil
	}

	// ── Step 5: Apply ──────────────────────────────────────────────────────
	tui.PrintStep(5, "Applying commits")
	fmt.Printf("  %s  Creating branch %s …\n\n", tui.IconOK(), tui.Branch(branchName))

	if err := git.CreateAndCheckout(branchName); err != nil {
		tui.PrintError("Failed to create branch", err.Error())
		return err
	}

	result, cherryErr := git.CherryPickWithProgress(selected,
		func(i, total int, c domain.Commit) {
			tui.PrintProgress(i, total, c)
		},
	)

	// ── Error path: roll back ──────────────────────────────────────────────
	if cherryErr != nil {
		tui.Spacer()
		tui.PrintError("Cherry-pick conflict — rolling back all changes")

		if checkoutErr := git.Checkout(originBranch); checkoutErr != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not restore branch %q: %v\n",
				originBranch, checkoutErr)
		}
		if deleteErr := git.DeleteBranch(branchName); deleteErr != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not delete partial branch %q: %v\n",
				branchName, deleteErr)
		}

		var conflictErr *giterrors.CherryPickConflictError
		if asConflict(cherryErr, &conflictErr) {
			tui.PrintError(
				fmt.Sprintf("Conflict in %s: %s", conflictErr.Hash[:7], conflictErr.Subject),
				"Resolve the conflict in your working branch first, then re-run.",
				"  git log --oneline   — review history",
				"  git tidy create     — start again after resolving",
			)
		}
		return cherryErr
	}

	// ── Step 6: Done ───────────────────────────────────────────────────────
	tui.PrintSuccess(
		fmt.Sprintf("Branch ready:  %s", tui.Branch(branchName)),
		fmt.Sprintf("Applied:       %s", tui.Accent(fmt.Sprintf("%d commit(s)", len(result.Applied)))),
		fmt.Sprintf("Based on:      %s", tui.Muted(originBranch)),
	)

	fmt.Printf("  %s Next steps:\n", tui.Info(tui.IconArrow))
	fmt.Printf("    %s  git push -u origin %s\n", tui.Muted("→"), branchName)
	fmt.Printf("    %s  open a pull request from %s\n\n", tui.Muted("→"), tui.Branch(branchName))

	return nil
}

func asConflict(err error, target **giterrors.CherryPickConflictError) bool {
	if e, ok := err.(*giterrors.CherryPickConflictError); ok {
		*target = e
		return true
	}
	return false
}
