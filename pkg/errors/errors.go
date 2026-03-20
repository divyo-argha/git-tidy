package errors

import (
	"fmt"
	"strings"
)

// ── Repository state errors ────────────────────────────────────────────────

type NotARepoError struct{}

func (e *NotARepoError) Error() string {
	return "not inside a git repository (run `git init` first)"
}

// DirtyWorkDirError carries the list of modified files for rich display.
type DirtyWorkDirError struct {
	Files []DirtyFile
}

type DirtyFile struct {
	Status string // e.g. "M ", " M", "??"
	Path   string
}

func (e *DirtyWorkDirError) Error() string {
	return fmt.Sprintf("working directory has %d uncommitted change(s)", len(e.Files))
}

// DetachedHEADError is returned when the repo is in detached HEAD state.
type DetachedHEADError struct {
	Hash string
}

func (e *DetachedHEADError) Error() string {
	return fmt.Sprintf("repository is in detached HEAD state at %s — check out a branch first", e.Hash[:7])
}

// ── Branch errors ──────────────────────────────────────────────────────────

type BranchExistsError struct {
	Name string
}

func (e *BranchExistsError) Error() string {
	return fmt.Sprintf("branch %q already exists", e.Name)
}

// InvalidBranchNameError carries the reason git check-ref-format rejected it.
type InvalidBranchNameError struct {
	Name   string
	Reason string
}

func (e *InvalidBranchNameError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("invalid branch name %q: %s", e.Name, e.Reason)
	}
	return fmt.Sprintf("invalid branch name %q", e.Name)
}

// ── Commit selection errors ────────────────────────────────────────────────

type InvalidSelectionError struct {
	Input string
}

func (e *InvalidSelectionError) Error() string {
	return fmt.Sprintf("invalid selection %q — enter numbers from the list above", e.Input)
}

// DuplicateCommitError is returned when a selected commit already exists on the target branch.
type DuplicateCommitError struct {
	Commits []DuplicateEntry
}

type DuplicateEntry struct {
	Hash    string
	Subject string
}

func (e *DuplicateCommitError) Error() string {
	subjects := make([]string, len(e.Commits))
	for i, c := range e.Commits {
		subjects[i] = fmt.Sprintf("%s (%s)", c.Hash[:7], c.Subject)
	}
	return fmt.Sprintf("%d commit(s) already exist on the target branch: %s",
		len(e.Commits), strings.Join(subjects, ", "))
}

// ── Cherry-pick errors ─────────────────────────────────────────────────────

// CherryPickConflictError carries full conflict detail for guided resolution.
type CherryPickConflictError struct {
	Hash          string
	Subject       string
	ConflictFiles []string // paths of files with conflicts
	Applied       int      // how many commits succeeded before this one
	Total         int      // total commits in the batch
}

func (e *CherryPickConflictError) Error() string {
	return fmt.Sprintf("cherry-pick conflict on commit %s (%q)", e.Hash[:7], e.Subject)
}

// ── Rollback errors ────────────────────────────────────────────────────────

// RollbackError is returned when a rollback itself fails, so the caller can
// surface both the original error and the rollback failure.
type RollbackError struct {
	Original error
	Cause    error
}

func (e *RollbackError) Error() string {
	return fmt.Sprintf("rollback failed (%v) while handling: %v", e.Cause, e.Original)
}

func (e *RollbackError) Unwrap() error { return e.Original }

// ── Infrastructure errors ──────────────────────────────────────────────────

type GitNotFoundError struct{}

func (e *GitNotFoundError) Error() string {
	return "git not found on PATH — install Git and try again"
}
