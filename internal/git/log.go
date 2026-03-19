package git

import (
	"fmt"
	"strings"

	"github.com/yourusername/git-tidy/internal/domain"
	giterrors "github.com/yourusername/git-tidy/pkg/errors"
)

const logSeparator = "|||"

// IsRepo returns true if the current directory is inside a Git repository.
func IsRepo() bool {
	_, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// IsClean returns true if the working tree has no uncommitted changes.
func IsClean() error {
	out, err := run("status", "--porcelain")
	if err != nil {
		return err
	}
	if out != "" {
		return &giterrors.DirtyWorkDirError{}
	}
	return nil
}

// Log returns the last `limit` commits on the current branch.
func Log(limit int) ([]domain.Commit, error) {
	if !IsRepo() {
		return nil, &giterrors.NotARepoError{}
	}

	format := fmt.Sprintf("%%H%s%%s%s%%an%s%%ar", logSeparator, logSeparator, logSeparator)
	out, err := run("log",
		fmt.Sprintf("--pretty=format:%s", format),
		fmt.Sprintf("-n%d", limit),
	)
	if err != nil {
		return nil, err
	}

	if out == "" {
		return nil, fmt.Errorf("no commits found in this repository")
	}

	lines := strings.Split(out, "\n")
	commits := make([]domain.Commit, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, logSeparator, 4)
		if len(parts) != 4 {
			continue
		}
		hash := strings.TrimSpace(parts[0])
		commits = append(commits, domain.Commit{
			Hash:      hash,
			ShortHash: hash[:7],
			Subject:   strings.TrimSpace(parts[1]),
			Author:    strings.TrimSpace(parts[2]),
			RelDate:   strings.TrimSpace(parts[3]),
		})
	}

	return commits, nil
}
