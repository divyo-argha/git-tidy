package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/divyo-argha/git-tidy/internal/domain"
	"github.com/divyo-argha/git-tidy/internal/git"
	"github.com/divyo-argha/git-tidy/internal/tui"
	giterrors "github.com/divyo-argha/git-tidy/pkg/errors"
)

func newCreateCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "create [branch-name]",
		Short: "Select commits interactively and cherry-pick them into a new branch",
		Long: `Shows recent commits, lets you choose which to keep, then creates
a clean branch with only those commits applied.

Safety checks run automatically:
  • Detects detached HEAD
  • Warns on uncommitted changes (with stash option)
  • Marks duplicate commits in the selection table
  • Validates branch name before any changes are made
  • Rolls back completely on any failure

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

	// ── Step 1: Preflight checks ───────────────────────────────────────────
	tui.PrintStep(1, "Checking repository state")

	if !git.IsRepo() {
		tui.PrintError("Not inside a git repository",
			"Run `git init` or navigate to an existing repository.")
		return fmt.Errorf("not a git repository")
	}

	// Detached HEAD — cherry-pick would apply to no branch.
	if err := git.CheckDetachedHEAD(); err != nil {
		tui.PrintError("Repository is in detached HEAD state",
			"Check out a branch before running git-tidy:",
			"  git checkout main",
			"  git checkout -b my-branch")
		return err
	}

	// Open the rollback journal now — it records the branch we're on.
	journal, err := git.NewJournal()
	if err != nil {
		return fmt.Errorf("could not initialise rollback journal: %w", err)
	}

	fmt.Printf("  %s  On branch %s\n", tui.IconOK(), tui.Branch(journal.OriginBranch))

	// Dirty working tree — interactive: stash or abort.
	dirtyFiles, err := git.DirtyFiles()
	if err != nil {
		return fmt.Errorf("could not check working tree: %w", err)
	}

	if len(dirtyFiles) > 0 && !dryRun {
		choice := tui.AskDirtyTree(dirtyFiles)
		switch choice {
		case tui.DirtyTreeAbort:
			fmt.Printf("\n  %s  Aborted — no changes made.\n\n", tui.IconWarn())
			return nil
		case tui.DirtyTreeStash:
			ref, stashErr := git.StashChanges("")
			if stashErr != nil {
				tui.PrintError("Auto-stash failed", stashErr.Error(),
					"Commit or manually stash your changes, then re-run.")
				return stashErr
			}
			journal.RecordStash(ref)
			fmt.Printf("  %s  Changes stashed (%s) — will be restored on finish\n",
				tui.IconOK(), tui.Muted(ref))
		}
	} else if len(dirtyFiles) > 0 && dryRun {
		tui.PrintWarning(fmt.Sprintf("%d uncommitted change(s) present — fine for preview, 'create' will prompt.", len(dirtyFiles)))
	} else {
		fmt.Printf("  %s  Working tree clean\n", tui.IconOK())
	}

	// ── Step 2: Validate branch name ──────────────────────────────────────
	tui.PrintStep(2, "Validating target branch")

	branchName := fmt.Sprintf("tidy/%s", time.Now().Format("20060102-150405"))
	if len(args) == 1 && args[0] != "" {
		branchName = args[0]
	}

	if err := git.ValidateBranchName(branchName); err != nil {
		var invErr *giterrors.InvalidBranchNameError
		if asInvalidBranch(err, &invErr) {
			tui.PrintError(
				fmt.Sprintf("Invalid branch name: %q", invErr.Name),
				invErr.Reason,
				"Examples of valid names: feature/login  fix/null-check  tidy/cleanup",
			)
		}
		return err
	}

	if git.BranchExists(branchName) {
		tui.PrintError(fmt.Sprintf("Branch %q already exists", branchName),
			"Choose a different name or delete the existing branch:",
			"  git branch -D "+branchName)
		return &giterrors.BranchExistsError{Name: branchName}
	}

	fmt.Printf("  %s  Target branch: %s\n", tui.IconOK(), tui.Branch(branchName))
	if len(args) == 0 {
		fmt.Printf("      %s\n", tui.Muted("(auto-generated — pass a name to override)"))
	}

	// ── Step 3: Fetch commits and mark duplicates ──────────────────────────
	tui.PrintStep(3, "Select commits")

	commits, err := git.Log(limit)
	if err != nil {
		tui.PrintError("Could not read commit history", err.Error())
		return err
	}

	// Build a set of commit indices that already exist on the origin branch.
	// We compare against OriginBranch (before we create the new branch).
	dupIndexes := map[int]bool{}
	for i, c := range commits {
		exists, checkErr := git.CommitExistsOnBranch(c.Hash, journal.OriginBranch)
		if checkErr == nil && exists {
			dupIndexes[i+1] = true
		}
	}

	selected, err := tui.SelectCommits(commits, dupIndexes)
	if err != nil {
		tui.PrintError("Commit selection failed", err.Error())
		return err
	}

	// ── Step 4: Duplicate-commit guard ────────────────────────────────────
	// Check selected commits against the origin branch and prompt to skip.
	hashes := make([]string, len(selected))
	subjects := make([]string, len(selected))
	for i, c := range selected {
		hashes[i] = c.Hash
		subjects[i] = c.Subject
	}

	dupErr := git.FindDuplicates(hashes, subjects, journal.OriginBranch)
	if dupErr != nil {
		var dupTyped *giterrors.DuplicateCommitError
		if asDuplicate(dupErr, &dupTyped) && !dryRun {
			skip := tui.AskDuplicateSkip(dupTyped.Commits)
			if !skip {
				fmt.Printf("\n  %s  Aborted — review your selection and re-run.\n\n", tui.IconWarn())
				return nil
			}
			// Remove duplicates from the selected slice.
			dupHashes := map[string]bool{}
			for _, d := range dupTyped.Commits {
				dupHashes[d.Hash] = true
			}
			filtered := selected[:0]
			for _, c := range selected {
				if !dupHashes[c.Hash] {
					filtered = append(filtered, c)
				}
			}
			selected = filtered
			fmt.Printf("  %s  Skipped %d duplicate(s) — %d commit(s) remaining\n",
				tui.IconOK(), len(dupTyped.Commits), len(selected))

			if len(selected) == 0 {
				fmt.Printf("\n  %s  Nothing left to apply after removing duplicates.\n\n", tui.IconWarn())
				return nil
			}
		}
	}

	// ── Step 5: Review & confirm ───────────────────────────────────────────
	tui.PrintStep(4, "Review & confirm")

	confirmed, err := tui.ConfirmSelection(selected, branchName, dryRun)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("  %s  Dry run complete — nothing was changed.\n", tui.IconWarn())
		fmt.Printf("      %s  Run %s to execute.\n\n",
			tui.Muted("→"), tui.Accent("git tidy create "+branchName))
		// Restore any auto-stash that happened during preview preflight.
		if journal.StashRef != "" {
			git.PopStash(journal.StashRef)
		}
		return nil
	}

	if !confirmed {
		tui.Spacer()
		fmt.Printf("  %s  Aborted — no changes made.\n", tui.IconWarn())
		// Restore stash if we auto-stashed.
		if journal.StashRef != "" {
			if popErr := git.PopStash(journal.StashRef); popErr != nil {
				fmt.Fprintf(os.Stderr, "  Warning: could not restore stash: %v\n", popErr)
			} else {
				fmt.Printf("  %s  Stash restored\n", tui.IconOK())
			}
		}
		fmt.Println()
		return nil
	}

	// ── Step 6: Create branch & cherry-pick ───────────────────────────────
	tui.PrintStep(5, "Applying commits")
	fmt.Printf("  %s  Creating branch %s …\n\n", tui.IconOK(), tui.Branch(branchName))

	if err := git.CreateAndCheckout(branchName); err != nil {
		tui.PrintError("Failed to create branch", err.Error())
		_ = journal.Rollback(err) // restore stash if needed
		return err
	}
	journal.RecordBranch(branchName)

	result, cherryErr := git.CherryPickWithProgress(selected,
		func(i, total int, c domain.Commit) {
			tui.PrintProgress(i, total, c)
		},
	)
	journal.RecordApplied(len(result.Applied))

	// ── Error path: guided conflict resolution + full rollback ─────────────
	if cherryErr != nil {
		var conflictErr *giterrors.CherryPickConflictError
		if asConflict(cherryErr, &conflictErr) {
			tui.AskConflictResolution(conflictErr, branchName)
		}

		rollbackErr := journal.Rollback(cherryErr)
		tui.PrintRollbackResult(journal.Summary(), rollbackErr)

		if rollbackErr != nil {
			return rollbackErr
		}
		return cherryErr
	}

	// ── Step 7: Restore stash (if we auto-stashed) ─────────────────────────
	if journal.StashRef != "" {
		if popErr := git.PopStash(journal.StashRef); popErr != nil {
			fmt.Fprintf(os.Stderr, "  %s  Could not restore stash %q: %v\n",
				tui.IconWarn(), journal.StashRef, popErr)
			fmt.Fprintf(os.Stderr, "     Run: %s\n", tui.Info("git stash pop"))
		} else {
			fmt.Printf("  %s  Stash restored\n", tui.IconOK())
		}
	}

	// ── Step 8: Success ────────────────────────────────────────────────────
	tui.PrintSuccess(
		fmt.Sprintf("Branch ready:  %s", tui.Branch(branchName)),
		fmt.Sprintf("Applied:       %s", tui.Accent(fmt.Sprintf("%d commit(s)", len(result.Applied)))),
		fmt.Sprintf("Based on:      %s", tui.Muted(journal.OriginBranch)),
	)

	fmt.Printf("  %s Next steps:\n", tui.Info(tui.IconArrow))
	fmt.Printf("    %s  git push -u origin %s\n", tui.Muted("→"), branchName)
	fmt.Printf("    %s  open a pull request from %s\n\n", tui.Muted("→"), tui.Branch(branchName))

	return nil
}

// ── Type-assertion helpers ─────────────────────────────────────────────────

func asConflict(err error, target **giterrors.CherryPickConflictError) bool {
	if e, ok := err.(*giterrors.CherryPickConflictError); ok {
		*target = e
		return true
	}
	return false
}

func asDuplicate(err error, target **giterrors.DuplicateCommitError) bool {
	if e, ok := err.(*giterrors.DuplicateCommitError); ok {
		*target = e
		return true
	}
	return false
}

func asInvalidBranch(err error, target **giterrors.InvalidBranchNameError) bool {
	if e, ok := err.(*giterrors.InvalidBranchNameError); ok {
		*target = e
		return true
	}
	return false
}
