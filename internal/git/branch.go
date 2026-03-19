package git

import (
	"strings"

	giterrors "github.com/yourusername/git-tidy/pkg/errors"
)

// BranchExists returns true if a branch with the given name already exists.
func BranchExists(name string) bool {
	out, err := run("branch", "--list", name)
	return err == nil && strings.TrimSpace(out) != ""
}

// ValidateBranchName returns an error if the name is not a legal Git ref.
func ValidateBranchName(name string) error {
	_, err := run("check-ref-format", "--branch", name)
	return err
}

// CreateAndCheckout creates a new branch from HEAD and checks it out.
func CreateAndCheckout(name string) error {
	if BranchExists(name) {
		return &giterrors.BranchExistsError{Name: name}
	}
	_, err := run("checkout", "-b", name)
	return err
}

// CurrentBranch returns the name of the currently checked-out branch.
func CurrentBranch() (string, error) {
	return run("rev-parse", "--abbrev-ref", "HEAD")
}

// Checkout switches to an existing branch.
func Checkout(name string) error {
	_, err := run("checkout", name)
	return err
}

// DeleteBranch force-deletes a branch (used for cleanup on error).
func DeleteBranch(name string) error {
	_, err := run("branch", "-D", name)
	return err
}
