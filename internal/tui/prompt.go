package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	giterrors "github.com/divyo-argha/git-tidy/pkg/errors"
)

// DirtyTreeChoice represents the user's response to an uncommitted-changes warning.
type DirtyTreeChoice int

const (
	DirtyTreeAbort  DirtyTreeChoice = iota // user wants to stop
	DirtyTreeStash                         // auto-stash, proceed, pop on finish
	DirtyTreeIgnore                        // proceed anyway (preview mode only)
)

// AskDirtyTree shows the list of uncommitted files and asks the user what to do.
// Returns DirtyTreeAbort, DirtyTreeStash, or DirtyTreeIgnore.
func AskDirtyTree(files []giterrors.DirtyFile) DirtyTreeChoice {
	Spacer()
	fmt.Printf("  %s  %s\n", IconWarn(), Warning("Uncommitted changes detected"))
	ThinDivider()
	Spacer()

	// Show up to 10 files; summarise the rest.
	show := files
	extra := 0
	if len(files) > 10 {
		show = files[:10]
		extra = len(files) - 10
	}
	for _, f := range show {
		statusLabel := formatStatus(f.Status)
		fmt.Printf("    %s  %s\n", statusLabel, f.Path)
	}
	if extra > 0 {
		fmt.Printf("    %s\n", Muted(fmt.Sprintf("… and %d more", extra)))
	}
	Spacer()

	fmt.Printf("  %s  How would you like to proceed?\n\n", Bold("→"))
	fmt.Printf("    %s  Auto-stash changes, create branch, then restore them\n", Accent("[s]"))
	fmt.Printf("    %s  Abort — I'll handle this manually\n", Muted("[a]"))
	Spacer()
	fmt.Printf("  %s ", Info(IconArrow))

	reader := bufio.NewReader(os.Stdin)
	for {
		raw, err := reader.ReadString('\n')
		if err != nil {
			return DirtyTreeAbort
		}
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "s", "stash":
			return DirtyTreeStash
		case "a", "abort", "q", "quit", "":
			return DirtyTreeAbort
		default:
			fmt.Printf("  %s  Enter %s or %s: ", IconWarn(), Accent("s"), Muted("a"))
		}
	}
}

// AskDuplicateSkip shows duplicate commits and asks whether to skip or abort.
// Returns true if the user wants to skip duplicates and continue.
func AskDuplicateSkip(dupes []giterrors.DuplicateEntry) bool {
	Spacer()
	fmt.Printf("  %s  %s\n", IconWarn(), Warning("Some commits already exist on the target branch"))
	ThinDivider()
	Spacer()

	for _, d := range dupes {
		fmt.Printf("    %s  %s  %s\n",
			Warning("~"),
			Hash(d.Hash[:7]),
			Truncate(d.Subject, 58),
		)
	}
	Spacer()

	fmt.Printf("  %s  These commits appear to have already been applied.\n", Info(IconArrow))
	fmt.Printf("    %s  Skip them and continue with the rest?\n\n", Muted("→"))
	fmt.Printf("    %s  Skip duplicates and continue\n", Accent("[s]"))
	fmt.Printf("    %s  Abort — let me review my selection\n", Muted("[a]"))
	Spacer()
	fmt.Printf("  %s ", Info(IconArrow))

	reader := bufio.NewReader(os.Stdin)
	for {
		raw, err := reader.ReadString('\n')
		if err != nil {
			return false
		}
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "s", "skip":
			return true
		case "a", "abort", "q", "quit", "":
			return false
		default:
			fmt.Printf("  %s  Enter %s or %s: ", IconWarn(), Accent("s"), Muted("a"))
		}
	}
}

// AskConflictResolution is shown when a cherry-pick conflicts mid-batch.
// It explains what happened and what to do next.
// Returns true if the user confirms they understand and want to exit cleanly.
func AskConflictResolution(err *giterrors.CherryPickConflictError, branchName string) {
	Spacer()
	Divider()
	fmt.Printf("  %s  %s\n\n", IconErr(), Danger("Cherry-pick conflict — rolling back"))

	fmt.Printf("  %-12s  %s  %s\n", Bold("Commit"), Hash(err.Hash[:7]), err.Subject)
	fmt.Printf("  %-12s  %s of %s applied successfully\n",
		Bold("Progress"),
		Success(fmt.Sprintf("%d", err.Applied)),
		fmt.Sprintf("%d", err.Total),
	)

	if len(err.ConflictFiles) > 0 {
		Spacer()
		fmt.Printf("  %s  Files that would have conflicted:\n", Bold("→"))
		for _, f := range err.ConflictFiles {
			fmt.Printf("      %s  %s\n", Danger("✖"), f)
		}
	}

	Spacer()
	fmt.Printf("  %s  %s\n", Bold("What happened:"),
		"The selected commit conflicts with changes already in the base branch.")
	Spacer()
	fmt.Printf("  %s  %s\n", Bold("Next steps:"), "")
	fmt.Printf("    %s  Resolve the conflict on your current branch:\n", Muted("1."))
	fmt.Printf("         %s\n", Info("git merge origin/main  # or rebase"))
	fmt.Printf("    %s  Then re-run:\n", Muted("2."))
	fmt.Printf("         %s\n", Info("git tidy create "+branchName))
	fmt.Printf("    %s  Or cherry-pick manually:\n", Muted("3."))
	fmt.Printf("         %s\n", Info("git cherry-pick "+err.Hash[:7]))

	Spacer()
	Divider()
}

// PrintRollbackResult prints what the rollback restored.
func PrintRollbackResult(summary string, err error) {
	if err == nil {
		fmt.Printf("  %s  Rollback complete: %s\n", IconOK(), Muted(summary))
	} else {
		fmt.Fprintf(os.Stderr, "  %s  Rollback incomplete: %s\n", IconErr(), Danger(err.Error()))
		fmt.Fprintf(os.Stderr, "     %s\n", Muted("You may need to clean up manually:"))
		fmt.Fprintf(os.Stderr, "     %s\n", Info("git branch -D <partial-branch>"))
		fmt.Fprintf(os.Stderr, "     %s\n", Info("git checkout <your-branch>"))
	}
}

// formatStatus maps git porcelain status codes to readable colored labels.
func formatStatus(code string) string {
	if len(code) < 2 {
		return Muted("??")
	}
	switch strings.TrimSpace(code) {
	case "M", "MM":
		return Warning(" M") // modified
	case "A":
		return Success(" A") // added
	case "D":
		return Danger(" D") // deleted
	case "R":
		return Info(" R") // renamed
	case "??":
		return Muted("??") // untracked
	case "UU":
		return Danger("UU") // conflict (shown in rollback context)
	default:
		return Muted(code)
	}
}
