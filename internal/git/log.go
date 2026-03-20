package git

import (
	"fmt"
	"strings"

	"github.com/divyo-argha/git-tidy/internal/domain"
	giterrors "github.com/divyo-argha/git-tidy/pkg/errors"
)

const logSeparator = "|||"

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
	if strings.TrimSpace(out) == "" {
		return nil, fmt.Errorf("no commits found in this repository")
	}

	var commits []domain.Commit
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, logSeparator, 4)
		if len(parts) != 4 {
			continue
		}
		hash := strings.TrimSpace(parts[0])
		if len(hash) < 7 {
			continue
		}
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
