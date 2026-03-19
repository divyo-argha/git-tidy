package errors

import "fmt"

// NotARepoError is returned when the working directory is not inside a Git repository.
type NotARepoError struct{}

func (e *NotARepoError) Error() string {
	return "not inside a git repository (run `git init` first)"
}

// DirtyWorkDirError is returned when there are uncommitted changes.
type DirtyWorkDirError struct{}

func (e *DirtyWorkDirError) Error() string {
	return "working directory has uncommitted changes (stash or commit them first)"
}

// BranchExistsError is returned when the target branch already exists.
type BranchExistsError struct {
	Name string
}

func (e *BranchExistsError) Error() string {
	return fmt.Sprintf("branch %q already exists", e.Name)
}

// CherryPickConflictError is returned when a cherry-pick fails due to a conflict.
type CherryPickConflictError struct {
	Hash    string
	Subject string
}

func (e *CherryPickConflictError) Error() string {
	return fmt.Sprintf("cherry-pick conflict on commit %s (%q)", e.Hash[:7], e.Subject)
}

// InvalidSelectionError is returned when the user provides an out-of-range commit index.
type InvalidSelectionError struct {
	Input string
}

func (e *InvalidSelectionError) Error() string {
	return fmt.Sprintf("invalid selection %q — enter numbers from the list above", e.Input)
}

// GitNotFoundError is returned when git is not on PATH.
type GitNotFoundError struct{}

func (e *GitNotFoundError) Error() string {
	return "git not found on PATH — install Git and try again"
}
