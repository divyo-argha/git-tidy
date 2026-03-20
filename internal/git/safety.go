package git

import (
	"fmt"
	"strings"

	giterrors "github.com/yourusername/git-tidy/pkg/errors"
)

// ── Repository state checks ────────────────────────────────────────────────

// IsRepo returns true if the current directory is inside a Git repository.
func IsRepo() bool {
	_, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// CheckDetachedHEAD returns an error if the repo is in detached HEAD state.
func CheckDetachedHEAD() error {
	branch, err := run("symbolic-ref", "--short", "HEAD")
	if err != nil || strings.TrimSpace(branch) == "" {
		// Detached — get the current hash for the error message.
		hash, _ := run("rev-parse", "--short", "HEAD")
		if hash == "" {
			hash = "unknown"
		}
		return &giterrors.DetachedHEADError{Hash: hash}
	}
	return nil
}

// DirtyFiles returns a structured list of all uncommitted changes.
// Returns nil (not an error) when the tree is clean.
func DirtyFiles() ([]giterrors.DirtyFile, error) {
	out, err := run("status", "--porcelain")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}

	var files []giterrors.DirtyFile
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 4 {
			continue
		}
		status := line[:2]
		path := strings.TrimSpace(line[3:])
		files = append(files, giterrors.DirtyFile{Status: status, Path: path})
	}
	return files, nil
}

// IsClean returns a DirtyWorkDirError (with file list) when the tree is dirty.
func IsClean() error {
	files, err := DirtyFiles()
	if err != nil {
		return err
	}
	if len(files) > 0 {
		return &giterrors.DirtyWorkDirError{Files: files}
	}
	return nil
}

// ── Duplicate-commit detection ─────────────────────────────────────────────

// CommitExistsOnBranch returns true if a commit with the same patch-id
// already exists on the given branch. Uses `git cherry` which compares
// patch content, not just hash — so cherry-picked equivalents are caught.
func CommitExistsOnBranch(commitHash, targetBranch string) (bool, error) {
	// `git cherry <upstream> <head> <limit>` lists commits in HEAD not in upstream.
	// We check from the targetBranch perspective.
	out, err := run("cherry", "-v", targetBranch, commitHash+"^", commitHash)
	if err != nil {
		// If the commit has no parent (first commit), fall back to log-based check.
		return commitExistsByMessage(commitHash, targetBranch)
	}
	// A leading '+' means the commit is NOT in targetBranch.
	// A leading '-' means it IS equivalent to one already there.
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.HasPrefix(line, "-") {
			return true, nil
		}
	}
	return false, nil
}

// commitExistsByMessage is a fallback: checks if any commit on targetBranch
// shares the same commit message subject as the given hash.
func commitExistsByMessage(commitHash, targetBranch string) (bool, error) {
	subject, err := run("log", "-1", "--pretty=format:%s", commitHash)
	if err != nil {
		return false, err
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return false, nil
	}

	out, err := run("log", targetBranch, "--pretty=format:%s", "--max-count=200")
	if err != nil {
		return false, nil // branch may not exist yet — not an error
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == subject {
			return true, nil
		}
	}
	return false, nil
}

// FindDuplicates checks all selected commits against the target branch and
// returns a DuplicateCommitError if any are already present.
// Pass targetBranch="" to skip (used before the branch is created).
func FindDuplicates(hashes []string, subjects []string, targetBranch string) error {
	if targetBranch == "" {
		return nil
	}

	var dupes []giterrors.DuplicateEntry
	for i, hash := range hashes {
		exists, err := CommitExistsOnBranch(hash, targetBranch)
		if err != nil {
			continue // skip on error — don't block on an uncertain check
		}
		if exists {
			subject := ""
			if i < len(subjects) {
				subject = subjects[i]
			}
			dupes = append(dupes, giterrors.DuplicateEntry{Hash: hash, Subject: subject})
		}
	}

	if len(dupes) > 0 {
		return &giterrors.DuplicateCommitError{Commits: dupes}
	}
	return nil
}

// ConflictingFiles returns the list of files currently in a conflicted state.
// Should be called after a failed cherry-pick, before aborting.
func ConflictingFiles() ([]string, error) {
	out, err := run("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// StashChanges stashes the current dirty state and returns the stash ref.
// Returns an error if nothing was stashed (clean tree).
func StashChanges(message string) (string, error) {
	if message == "" {
		message = "git-tidy: auto-stash before create"
	}
	_, err := run("stash", "push", "-m", message)
	if err != nil {
		return "", fmt.Errorf("stash failed: %w", err)
	}
	// Get the stash ref we just created.
	ref, err := run("stash", "list", "--max-count=1", "--pretty=format:%gd")
	if err != nil || strings.TrimSpace(ref) == "" {
		return "stash@{0}", nil
	}
	return strings.TrimSpace(ref), nil
}

// PopStash re-applies the given stash ref.
func PopStash(ref string) error {
	if ref == "" {
		ref = "stash@{0}"
	}
	_, err := run("stash", "pop", ref)
	return err
}
