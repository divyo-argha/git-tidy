package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/git-tidy/internal/git"
	"github.com/yourusername/git-tidy/internal/tui"
	giterrors "github.com/yourusername/git-tidy/pkg/errors"
)

func newCreateCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "create [branch-name]",
		Short: "Interactively select commits and cherry-pick them into a new branch",
		Long: `Shows recent commits, lets you pick which ones to keep,
then creates a clean branch with only those commits applied.

If no branch name is given, one is generated automatically: tidy/<timestamp>`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(args, limit)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of commits to show")
	return cmd
}

func runCreate(args []string, limit int) error {
	// ── 1. Preflight checks ────────────────────────────────────────────────
	if !git.IsRepo() {
		return &giterrors.NotARepoError{}
	}

	if err := git.IsClean(); err != nil {
		return err
	}

	originBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("could not determine current branch: %w", err)
	}

	// ── 2. Resolve branch name ─────────────────────────────────────────────
	branchName := fmt.Sprintf("tidy/%s", time.Now().Format("20060102-150405"))
	if len(args) == 1 && args[0] != "" {
		branchName = args[0]
	}

	if err := git.ValidateBranchName(branchName); err != nil {
		return fmt.Errorf("invalid branch name %q: %w", branchName, err)
	}

	if git.BranchExists(branchName) {
		return &giterrors.BranchExistsError{Name: branchName}
	}

	// ── 3. Fetch and display commits ───────────────────────────────────────
	commits, err := git.Log(limit)
	if err != nil {
		return err
	}

	tui.ShowCommits(commits)

	// ── 4. Interactive selection ───────────────────────────────────────────
	selected, err := tui.SelectCommits(commits)
	if err != nil {
		return err
	}

	// ── 5. Confirm before making any changes ──────────────────────────────
	confirmed, err := tui.ConfirmSelection(selected, branchName)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("\n  Aborted. No changes made.\n")
		return nil
	}

	// ── 6. Create branch and cherry-pick ──────────────────────────────────
	fmt.Printf("\n  Creating branch %s ...\n\n", highlight(branchName))

	if err := git.CreateAndCheckout(branchName); err != nil {
		return err
	}

	result, cherryErr := git.CherryPick(selected)

	// ── 7. Handle conflict: clean up and restore ───────────────────────────
	if cherryErr != nil {
		fmt.Fprintf(os.Stderr, "\n  %s Cherry-pick conflict — rolling back\n", errIcon())

		// Switch back to original branch.
		if checkoutErr := git.Checkout(originBranch); checkoutErr != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not restore branch %q: %v\n", originBranch, checkoutErr)
		}

		// Delete the partially-built branch.
		if deleteErr := git.DeleteBranch(branchName); deleteErr != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not delete branch %q: %v\n", branchName, deleteErr)
		}

		var conflictErr *giterrors.CherryPickConflictError
		if asConflict(cherryErr, &conflictErr) {
			fmt.Fprintf(os.Stderr,
				"\n  Conflict in commit %s: %s\n"+
					"  Tip: resolve the conflict in the original branch first, then re-run.\n\n",
				conflictErr.Hash[:7], conflictErr.Subject)
		}

		return cherryErr
	}

	// ── 8. Success summary ─────────────────────────────────────────────────
	fmt.Println()
	tui.Divider()
	fmt.Printf("  %s  Branch ready: %s\n", successIcon(), highlight(branchName))
	fmt.Printf("  %s  %d commit(s) applied\n", successIcon(), len(result.Applied))
	tui.Divider()
	fmt.Println()

	return nil
}

// ── Helpers ────────────────────────────────────────────────────────────────

func highlight(s string) string {
	return "\033[1m\033[97m" + s + "\033[0m"
}

func successIcon() string { return "\033[32m✓\033[0m" }
func errIcon() string     { return "\033[31m✗\033[0m" }

// asConflict is a simple type-assertion helper to avoid importing errors pkg twice.
func asConflict(err error, target **giterrors.CherryPickConflictError) bool {
	if e, ok := err.(*giterrors.CherryPickConflictError); ok {
		*target = e
		return true
	}
	return false
}
