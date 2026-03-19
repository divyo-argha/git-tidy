package git

import (
	"fmt"
	"strings"

	"github.com/yourusername/git-tidy/internal/domain"
	giterrors "github.com/yourusername/git-tidy/pkg/errors"
)

// CherryPickResult holds the outcome of a cherry-pick batch.
type CherryPickResult struct {
	Applied []domain.Commit
	Failed  *domain.Commit // non-nil if a conflict stopped the batch
}

// CherryPick applies the given commits in order onto the current branch.
// On conflict it aborts the cherry-pick, records the failing commit, and returns.
// Commits should be in the order they are to be applied (oldest first for a clean history).
func CherryPick(commits []domain.Commit) (CherryPickResult, error) {
	result := CherryPickResult{}

	for i := range commits {
		c := commits[i]
		_, err := run("cherry-pick", c.Hash)
		if err != nil {
			// Abort the in-progress cherry-pick so the repo is left clean.
			abortCherryPick()
			result.Failed = &c
			return result, &giterrors.CherryPickConflictError{
				Hash:    c.Hash,
				Subject: c.Subject,
			}
		}
		result.Applied = append(result.Applied, c)
		fmt.Printf("    ✓  %s  %s\n", c.ShortHash, truncate(c.Subject, 55))
	}

	return result, nil
}

// abortCherryPick runs `git cherry-pick --abort`, ignoring errors
// (the repo may already be clean if the conflict happened before any write).
func abortCherryPick() {
	run("cherry-pick", "--abort") //nolint:errcheck
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
