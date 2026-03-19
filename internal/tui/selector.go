package tui

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/yourusername/git-tidy/internal/domain"
	giterrors "github.com/yourusername/git-tidy/pkg/errors"
)

// ShowCommits prints the numbered commit list to stdout.
func ShowCommits(commits []domain.Commit) {
	fmt.Println()
	fmt.Println(header("  Recent commits"))
	Divider()
	fmt.Printf(muted("  %3s  %-7s  %-60s  %-16s  %s\n"), "#", "hash", "subject", "author", "when")
	Divider()
	for i, c := range commits {
		fmt.Println(c.Display(i + 1))
	}
	Divider()
	fmt.Println()
}

// SelectCommits prompts the user to pick commits by number and returns them
// in chronological order (oldest first) so cherry-picks apply cleanly.
func SelectCommits(commits []domain.Commit) ([]domain.Commit, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("  %s Enter commit numbers to cherry-pick %s\n",
		highlight("→"), muted("(e.g. 1,3,5 or 1-4 or 2,4-6)"))
	fmt.Printf("  %s ", cyan+"›"+reset)

	raw, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("no commits selected")
	}

	indices, err := parseSelection(raw, len(commits))
	if err != nil {
		return nil, err
	}

	// Deduplicate and sort ascending so oldest commit is applied first.
	seen := make(map[int]bool)
	unique := []int{}
	for _, idx := range indices {
		if !seen[idx] {
			seen[idx] = true
			unique = append(unique, idx)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(unique))) // reverse because log is newest-first

	selected := make([]domain.Commit, 0, len(unique))
	for _, idx := range unique {
		selected = append(selected, commits[idx-1])
	}

	return selected, nil
}

// ConfirmSelection shows the selected commits and asks the user to confirm.
func ConfirmSelection(commits []domain.Commit, branchName string) (bool, error) {
	fmt.Println()
	fmt.Printf("  %s Will cherry-pick %s into branch %s:\n",
		warn("!"),
		highlight(fmt.Sprintf("%d commit(s)", len(commits))),
		highlight(branchName),
	)
	fmt.Println()
	for _, c := range commits {
		fmt.Printf("      %s  %s\n", muted(c.ShortHash), c.Subject)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("  %s Proceed? %s ", highlight("→"), muted("[y/N]"))
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

// parseSelection parses user input like "1,2,4-6" into a slice of 1-based indices.
func parseSelection(input string, max int) ([]int, error) {
	var indices []int

	// Split on commas first.
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for a range like "2-5".
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				return nil, &giterrors.InvalidSelectionError{Input: part}
			}
			lo, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			hi, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err1 != nil || err2 != nil || lo < 1 || hi > max || lo > hi {
				return nil, &giterrors.InvalidSelectionError{Input: part}
			}
			for i := lo; i <= hi; i++ {
				indices = append(indices, i)
			}
			continue
		}

		// Single number.
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > max {
			return nil, &giterrors.InvalidSelectionError{Input: part}
		}
		indices = append(indices, n)
	}

	if len(indices) == 0 {
		return nil, fmt.Errorf("no valid commit numbers found in %q", input)
	}

	return indices, nil
}
