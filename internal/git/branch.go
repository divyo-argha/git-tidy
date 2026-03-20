package git

import (
	"strings"

	giterrors "github.com/divyo-argha/git-tidy/pkg/errors"
)

// BranchExists returns true if a local branch with the given name exists.
func BranchExists(name string) bool {
	out, err := run("branch", "--list", name)
	return err == nil && strings.TrimSpace(out) != ""
}

// ValidateBranchName checks the name against git's ref-format rules and
// returns a typed InvalidBranchNameError with a human-readable reason.
func ValidateBranchName(name string) error {
	if strings.TrimSpace(name) == "" {
		return &giterrors.InvalidBranchNameError{
			Name:   name,
			Reason: "name cannot be empty",
		}
	}

	// Forbidden patterns git itself enforces, with readable explanations.
	rules := []struct {
		check  func(string) bool
		reason string
	}{
		{func(s string) bool { return strings.Contains(s, " ") },
			"spaces are not allowed — use hyphens or slashes"},
		{func(s string) bool { return strings.Contains(s, "..") },
			"double dots (..) are not allowed"},
		{func(s string) bool { return strings.Contains(s, "@{") },
			"@{ sequence is not allowed"},
		{func(s string) bool { return strings.HasSuffix(s, ".lock") },
			"name cannot end with .lock"},
		{func(s string) bool { return strings.HasSuffix(s, ".") },
			"name cannot end with a dot"},
		{func(s string) bool { return strings.HasPrefix(s, "-") },
			"name cannot start with a hyphen"},
		{func(s string) bool { return strings.Contains(s, "\\") },
			"backslashes are not allowed"},
		{func(s string) bool { return strings.Contains(s, "~") || strings.Contains(s, "^") || strings.Contains(s, ":") },
			"characters ~, ^, and : are not allowed"},
	}

	for _, rule := range rules {
		if rule.check(name) {
			return &giterrors.InvalidBranchNameError{Name: name, Reason: rule.reason}
		}
	}

	// Final authority: let git validate the rest (unicode edge cases, etc.)
	if _, err := run("check-ref-format", "--branch", name); err != nil {
		return &giterrors.InvalidBranchNameError{Name: name, Reason: "rejected by git check-ref-format"}
	}
	return nil
}

// CreateAndCheckout creates a new branch from HEAD and switches to it.
// Returns BranchExistsError if the name is already taken.
func CreateAndCheckout(name string) error {
	if BranchExists(name) {
		return &giterrors.BranchExistsError{Name: name}
	}
	_, err := run("checkout", "-b", name)
	return err
}

// CurrentBranch returns the short name of the checked-out branch.
func CurrentBranch() (string, error) {
	return run("rev-parse", "--abbrev-ref", "HEAD")
}

// Checkout switches to an existing branch.
func Checkout(name string) error {
	_, err := run("checkout", name)
	return err
}

// DeleteBranch force-deletes a branch. Used for cleanup after a failed run.
func DeleteBranch(name string) error {
	_, err := run("branch", "-D", name)
	return err
}
