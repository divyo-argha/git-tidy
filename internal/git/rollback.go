package git

import (
	"fmt"

	giterrors "github.com/yourusername/git-tidy/pkg/errors"
)

// RollbackJournal records the repository state before a mutating operation
// so it can be cleanly restored if anything goes wrong.
type RollbackJournal struct {
	OriginBranch  string // branch we were on before create
	CreatedBranch string // branch we created (to delete on rollback)
	StashRef      string // non-empty if we auto-stashed a dirty tree
	Applied       int    // how many cherry-picks succeeded before failure
}

// NewJournal captures the current branch name as the restore point.
func NewJournal() (*RollbackJournal, error) {
	branch, err := CurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("could not determine current branch for rollback journal: %w", err)
	}
	return &RollbackJournal{OriginBranch: branch}, nil
}

// RecordBranch notes that we created a branch (so rollback can delete it).
func (j *RollbackJournal) RecordBranch(name string) {
	j.CreatedBranch = name
}

// RecordStash notes that we auto-stashed changes (so rollback can restore them).
func (j *RollbackJournal) RecordStash(ref string) {
	j.StashRef = ref
}

// RecordApplied tracks progress so the error message can report partial success.
func (j *RollbackJournal) RecordApplied(n int) {
	j.Applied = n
}

// Rollback restores the repo to the state captured at journal creation:
//  1. Abort any in-progress cherry-pick
//  2. Switch back to OriginBranch
//  3. Delete CreatedBranch (force, since it may be partially populated)
//  4. Pop any auto-stash
//
// All steps are attempted even if earlier ones fail.
// Returns a RollbackError if any step fails, wrapping the original trigger error.
func (j *RollbackJournal) Rollback(triggerErr error) error {
	var rollbackErr error

	// 1. Abort any in-progress cherry-pick (idempotent — safe to call even if none).
	abortCherryPick()

	// 2. Return to the origin branch.
	if j.OriginBranch != "" {
		if err := Checkout(j.OriginBranch); err != nil {
			rollbackErr = fmt.Errorf("could not restore branch %q: %w", j.OriginBranch, err)
		}
	}

	// 3. Delete the partial branch we created.
	if j.CreatedBranch != "" && BranchExists(j.CreatedBranch) {
		if err := DeleteBranch(j.CreatedBranch); err != nil && rollbackErr == nil {
			rollbackErr = fmt.Errorf("could not delete partial branch %q: %w", j.CreatedBranch, err)
		}
	}

	// 4. Restore any auto-stashed changes.
	if j.StashRef != "" {
		if err := PopStash(j.StashRef); err != nil && rollbackErr == nil {
			rollbackErr = fmt.Errorf("could not restore stash %q: %w", j.StashRef, err)
		}
	}

	if rollbackErr != nil {
		return &giterrors.RollbackError{
			Original: triggerErr,
			Cause:    rollbackErr,
		}
	}
	return triggerErr // rollback succeeded — return the original error unchanged
}

// Summary returns a human-readable description of what the rollback restored.
func (j *RollbackJournal) Summary() string {
	parts := []string{}
	if j.OriginBranch != "" {
		parts = append(parts, fmt.Sprintf("restored branch → %s", j.OriginBranch))
	}
	if j.CreatedBranch != "" {
		parts = append(parts, fmt.Sprintf("deleted partial branch %s", j.CreatedBranch))
	}
	if j.StashRef != "" {
		parts = append(parts, fmt.Sprintf("restored stash %s", j.StashRef))
	}
	if len(parts) == 0 {
		return "nothing to restore"
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
