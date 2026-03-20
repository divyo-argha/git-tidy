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

// ShowCommits prints the numbered commit list, highlighting selected rows.
func ShowCommits(commits []domain.Commit, selected map[int]bool) {
	Spacer()
	fmt.Printf("  %s  %s\n", Header("Recent commits"), Muted(fmt.Sprintf("(%d shown)", len(commits))))
	Divider()
	fmt.Printf(Muted("  %3s  %-2s  %-7s  %-54s  %-18s  %s\n"),
		"#", " ", "hash", "subject", "author", "when")
	ThinDivider()

	for i, c := range commits {
		idx := i + 1
		num := fmt.Sprintf("%2d", idx)
		hash := Hash(c.ShortHash)
		subj := Truncate(c.Subject, 54)
		auth := Author(Truncate(c.Author, 18))
		when := When(c.RelDate)

		var icon string
		if selected[idx] {
			icon = Success(IconPick)
		} else {
			icon = Muted(IconEmpty)
		}

		row := fmt.Sprintf("  %s  %s  %s  %-54s  %-18s  %s",
			Muted(num), icon, hash, subj, auth, when)

		if selected[idx] {
			fmt.Println(SelectedRow(row))
		} else {
			fmt.Println(row)
		}
	}

	Divider()
}

// SelectCommits prompts the user to pick commits by number.
// Supports: individual numbers (1,3,5), ranges (2-6), "all", and re-prompts on bad input.
// Returns commits sorted oldest-first for clean cherry-pick ordering.
func SelectCommits(commits []domain.Commit) ([]domain.Commit, error) {
	reader := bufio.NewReader(os.Stdin)

	// First render with empty selection.
	ShowCommits(commits, map[int]bool{})

	for {
		Spacer()
		fmt.Printf("  %s Select commits to cherry-pick\n", Bold("→"))
		fmt.Printf("  %s\n", Muted("    Enter numbers, ranges, or \"all\"  e.g.  1,3,5   2-6   all"))
		fmt.Printf("  %s ", Info(IconArrow))

		raw, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}
		raw = strings.TrimSpace(raw)

		if raw == "" {
			PrintWarning("No input — enter at least one commit number.")
			continue
		}

		// "all" shorthand.
		if strings.EqualFold(raw, "all") {
			result := make([]domain.Commit, len(commits))
			copy(result, commits)
			reverseCommits(result) // log is newest-first; apply oldest-first
			allSelected := map[int]bool{}
			for i := range commits {
				allSelected[i+1] = true
			}
			Spacer()
			fmt.Printf("  %s  all %s commits selected\n",
				Success(fmt.Sprintf("%d", len(commits))),
				Muted("highlighted below"),
			)
			ShowCommits(commits, allSelected)
			return result, nil
		}

		indices, err := parseSelection(raw, len(commits))
		if err != nil {
			PrintWarning(err.Error())
			continue
		}

		// Build highlight map and re-render so user sees what they picked.
		selectedMap := map[int]bool{}
		for _, idx := range indices {
			selectedMap[idx] = true
		}

		Spacer()
		fmt.Printf("  %s  %s selected — highlighted below\n",
			Success(fmt.Sprintf("%d commit(s)", len(selectedMap))),
			Muted("confirm on the next step"),
		)
		ShowCommits(commits, selectedMap)

		// Deduplicate + sort descending (newest-first in log), then build slice.
		unique := uniqueSorted(indices)
		result := make([]domain.Commit, 0, len(unique))
		for _, idx := range unique {
			result = append(result, commits[idx-1])
		}

		return result, nil
	}
}

// ConfirmSelection shows the execution plan and asks the user to confirm.
// When dryRun is true it prints the plan without prompting and returns false.
func ConfirmSelection(commits []domain.Commit, branchName string, dryRun bool) (bool, error) {
	Spacer()
	Divider()
	fmt.Printf("  %-10s  %s\n", Bold("Branch"), Branch(branchName))
	fmt.Printf("  %-10s  %s\n", Bold("Commits"), Accent(fmt.Sprintf("%d selected", len(commits))))
	ThinDivider()
	Spacer()

	fmt.Printf("  %s  Apply order %s:\n", Info(IconArrow), Muted("(oldest → newest)"))
	Spacer()
	for i, c := range commits {
		connector := "├"
		if i == len(commits)-1 {
			connector = "└"
		}
		fmt.Printf("  %s  %s  %s  %s\n",
			Muted(connector+"─"),
			Hash(c.ShortHash),
			c.Subject,
			When("("+c.RelDate+")"),
		)
	}
	Spacer()

	if dryRun {
		PrintDryRun()
		return false, nil
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("  %s Proceed? %s ", Bold("→"), Muted("[y/N]"))
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

// PrintProgress renders a single cherry-pick progress line.
func PrintProgress(current, total int, c domain.Commit) {
	bar := progressBar(current, total, 14)
	fmt.Printf("  %s  %s  %s  %s\n",
		bar,
		Hash(c.ShortHash),
		Truncate(c.Subject, 48),
		Muted(fmt.Sprintf("%d/%d", current, total)),
	)
}

// ── Internal helpers ───────────────────────────────────────────────────────

func parseSelection(input string, max int) ([]int, error) {
	var indices []int

	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

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

// uniqueSorted deduplicates and sorts descending (newest log index first),
// so when we pull commits[idx-1] in order we get newest→oldest,
// which after reverseCommits becomes oldest→newest for cherry-pick.
func uniqueSorted(indices []int) []int {
	seen := make(map[int]bool, len(indices))
	var uniq []int
	for _, i := range indices {
		if !seen[i] {
			seen[i] = true
			uniq = append(uniq, i)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(uniq)))
	return uniq
}

func reverseCommits(s []domain.Commit) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func progressBar(current, total, width int) string {
	if total == 0 {
		return Muted(strings.Repeat("─", width))
	}
	filled := current * width / total
	bar := Success(strings.Repeat("█", filled)) + Muted(strings.Repeat("░", width-filled))
	return "[" + bar + "]"
}
