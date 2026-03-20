package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	giterrors "github.com/divyo-argha/git-tidy/pkg/errors"
)

// run executes a git command in the current working directory.
// It returns trimmed stdout on success, or an error on failure.
func run(args ...string) (string, error) {
	path, err := exec.LookPath("git")
	if err != nil {
		return "", &giterrors.GitNotFoundError{}
	}

	cmd := exec.Command(path, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", args[0], msg)
	}

	return strings.TrimSpace(stdout.String()), nil
}
