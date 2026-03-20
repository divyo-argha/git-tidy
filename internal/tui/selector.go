package tui

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/divyo-argha/git-tidy/internal/domain"
	giterrors "github.com/divyo-argha/git-tidy/pkg/errors"
)

// ShowCommits prints the numbered commit list.
// selected: 1-based index set of chosen commits (highlighted in green).
// duplicates: 1-based index set of commits already on the target branch (marked with ~).
func ShowCommits(commits []domain.Commit, selected map[int]bool, duplicates map[int]bool) {
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

		isDup := duplicates != nil && duplicates[idx]
		isSel := selected[idx]

		// Choose icon: selected > duplicate > empty
		var icon string
		switch {
		case isSel:
			icon = Success(IconPick)
		case isDup:
			icon = Warning("~") // already applied
		default:
			icon = Muted(IconEmpty)
		}

		// Dim duplicates that aren't explicitly selected.
		if isDup && !isSel {
			subj = Muted(subj)
		}

		row := fmt.Sprintf("  %s  %s  %s  %-54s  %-18s  %s",
			Muted(num), icon, hash, subj, auth, when)

		switch {
		case isSel:
			fmt.Println(SelectedRow(row))
		case isDup:
			fmt.Println(Muted(row))
		default:
			fmt.Println(row)
		}
	}

	Divider()

	// Legend line if any duplicates are present.
	if duplicates != nil && len(duplicates) > 0 {
		fmt.Printf("  %s  %s  already on target branch\n\n",
			Warning("~"), Muted("="))
	}
}

// SelectCommits prompts the user to pick commits by number.
// duplicates: optional set of 1-based indices to mark as already-applied.
// Returns commits oldest-first for clean cherry-pick ordering.
func SelectCommits(commits []domain.Commit, duplicates map[int]bool) ([]domain.Commit, error) {
	reader := bufio.NewReader(os.Stdin)

	ShowCommits(commits, map[int]bool{}, duplicates)

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

		if strings.EqualFold(raw, "all") {
			result := make([]domain.Commit, len(commits))
			copy(result, commits)
			reverseCommits(result)
			allSel := map[int]bool{}
			for i := range commits {
				allSel[i+1] = true
			}
			Spacer()
			fmt.Printf("  %s  all %s commits selected\n",
				Success(fmt.Sprintf("%d", len(commits))), Muted("highlighted below"))
			ShowCommits(commits, allSel, duplicates)
			return result, nil
		}

		indices, err := parseSelection(raw, len(commits))
		if err != nil {
			PrintWarning(err.Error())
			continue
		}

		selectedMap := map[int]bool{}
		for _, idx := range indices {
			selectedMap[idx] = true
		}

		Spacer()
		fmt.Printf("  %s  %s selected — highlighted below\n",
			Success(fmt.Sprintf("%d commit(s)", len(selectedMap))),
			Muted("confirm on the next step"),
		)
		ShowCommits(commits, selectedMap, duplicates)

		unique := uniqueSorted(indices)
		result := make([]domain.Commit, 0, len(unique))
		for _, idx := range unique {
			result = append(result, commits[idx-1])
		}
		return result, nil
	}
}

// ConfirmSelection renders the execution plan.
// dryRun=true skips the y/N prompt and always returns false.
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
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				return nil, &giterrors.InvalidSelectionError{Input: part}
			}
			lo, e1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			hi, e2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if e1 != nil || e2 != nil || lo < 1 || hi > max || lo > hi {
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
