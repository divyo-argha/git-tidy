package git

import (
	"strings"

	"github.com/divyo-argha/git-tidy/internal/domain"
	giterrors "github.com/divyo-argha/git-tidy/pkg/errors"
)

// CherryPickResult holds the full outcome of a cherry-pick batch.
type CherryPickResult struct {
	Applied []domain.Commit
	Failed  *domain.Commit // non-nil when a conflict stopped the batch
}

// ProgressFunc is called after each successful cherry-pick.
// current is 1-based.
type ProgressFunc func(current, total int, c domain.Commit)

// CherryPickWithProgress applies commits in order onto the current branch.
// On conflict it:
//  1. Collects the conflicting file list for rich error output
//  2. Aborts the cherry-pick (leaving the working tree clean)
//  3. Returns a CherryPickConflictError with full context
//
// The caller (cmd/create.go) is responsible for branch cleanup via the journal.
func CherryPickWithProgress(commits []domain.Commit, progress ProgressFunc) (CherryPickResult, error) {
	result := CherryPickResult{}
	total := len(commits)

	for i, c := range commits {
		_, err := run("cherry-pick", c.Hash)
		if err != nil {
			// Collect conflict file list before aborting — abort clears this info.
			conflictFiles, _ := ConflictingFiles()

			abortCherryPick()
			result.Failed = &commits[i]

			return result, &giterrors.CherryPickConflictError{
				Hash:          c.Hash,
				Subject:       c.Subject,
				ConflictFiles: conflictFiles,
				Applied:       len(result.Applied),
				Total:         total,
			}
		}

		result.Applied = append(result.Applied, c)
		if progress != nil {
			progress(i+1, total, c)
		}
	}

	return result, nil
}

// CherryPick is a convenience wrapper with no progress callback.
func CherryPick(commits []domain.Commit) (CherryPickResult, error) {
	return CherryPickWithProgress(commits, nil)
}

func abortCherryPick() {
	run("cherry-pick", "--abort") //nolint:errcheck
}

// commitMessages returns the subjects of all commits in a branch's history.
// Used by duplicate detection.
func commitMessages(branch string) ([]string, error) {
	out, err := run("log", branch, "--pretty=format:%s", "--max-count=500")
	if err != nil {
		return nil, err
	}
	var msgs []string
	for _, line := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			msgs = append(msgs, t)
		}
	}
	return msgs, nil
}
