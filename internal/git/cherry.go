package git

import (
	"github.com/yourusername/git-tidy/internal/domain"
	giterrors "github.com/yourusername/git-tidy/pkg/errors"
)

// CherryPickResult holds the outcome of a cherry-pick batch.
type CherryPickResult struct {
	Applied []domain.Commit
	Failed  *domain.Commit // non-nil when a conflict stopped the batch
}

// ProgressFunc is called after each successful cherry-pick.
// current is 1-based, total is len(commits).
type ProgressFunc func(current, total int, c domain.Commit)

// CherryPickWithProgress applies commits in order onto the current branch,
// calling progress after each success.
// On conflict it aborts, leaves the caller responsible for branch cleanup.
func CherryPickWithProgress(commits []domain.Commit, progress ProgressFunc) (CherryPickResult, error) {
	result := CherryPickResult{}
	total := len(commits)

	for i, c := range commits {
		if _, err := run("cherry-pick", c.Hash); err != nil {
			abortCherryPick()
			result.Failed = &commits[i]
			return result, &giterrors.CherryPickConflictError{
				Hash:    c.Hash,
				Subject: c.Subject,
			}
		}
		result.Applied = append(result.Applied, c)
		if progress != nil {
			progress(i+1, total, c)
		}
	}

	return result, nil
}

// CherryPick is a convenience wrapper without a progress callback.
func CherryPick(commits []domain.Commit) (CherryPickResult, error) {
	return CherryPickWithProgress(commits, nil)
}

func abortCherryPick() {
	run("cherry-pick", "--abort") //nolint:errcheck
}
