package tui

import (
	"fmt"
	"os"
	"strings"
)

// ── Color detection ────────────────────────────────────────────────────────
// Respects NO_COLOR (https://no-color.org) and non-terminal stdout.

var colorsEnabled = func() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("GIT_TIDY_NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}()

// ── Raw ANSI codes ─────────────────────────────────────────────────────────

const (
	ansiReset      = "\033[0m"
	ansiBold       = "\033[1m"
	ansiDim        = "\033[2m"
	ansiItalic     = "\033[3m"
	ansiGreen      = "\033[32m"
	ansiYellow     = "\033[33m"
	ansiBlue       = "\033[34m"
	ansiMagenta    = "\033[35m"
	ansiCyan       = "\033[36m"
	ansiWhite      = "\033[97m"
	ansiRed        = "\033[31m"
	ansiBgSelected = "\033[48;5;236m" // dark gray bg — selected row
	ansiFgSelected = "\033[38;5;117m" // sky blue fg — selected row
)

func esc(code, s string) string {
	if !colorsEnabled {
		return s
	}
	return code + s + ansiReset
}

// ── Semantic style functions ───────────────────────────────────────────────

func Bold(s string) string    { return esc(ansiBold, s) }
func Muted(s string) string   { return esc(ansiDim, s) }
func Italic(s string) string  { return esc(ansiItalic, s) }
func Header(s string) string  { return esc(ansiBold+ansiCyan, s) }
func Success(s string) string { return esc(ansiBold+ansiGreen, s) }
func Warning(s string) string { return esc(ansiYellow, s) }
func Danger(s string) string  { return esc(ansiBold+ansiRed, s) }
func Accent(s string) string  { return esc(ansiBold+ansiWhite, s) }
func Info(s string) string    { return esc(ansiCyan, s) }
func Branch(s string) string  { return esc(ansiBold+ansiMagenta, s) }
func Hash(s string) string    { return esc(ansiYellow, s) }
func Author(s string) string  { return esc(ansiBlue, s) }
func When(s string) string    { return esc(ansiDim+ansiCyan, s) }

// SelectedRow highlights an entire row as selected.
func SelectedRow(s string) string {
	if !colorsEnabled {
		return "▶ " + s
	}
	return ansiBgSelected + ansiBold + ansiFgSelected + s + ansiReset
}

// ── Icons ──────────────────────────────────────────────────────────────────

const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "!"
	IconArrow   = "›"
	IconPick    = "◆"
	IconEmpty   = "◇"
)

func IconOK() string   { return Success(IconSuccess) }
func IconErr() string  { return Danger(IconError) }
func IconWarn() string { return Warning(IconWarning) }
func IconInfo() string { return Info(IconArrow) }

// ── Layout helpers ─────────────────────────────────────────────────────────

const dividerWidth = 74

func Divider() {
	fmt.Println(Muted("  " + strings.Repeat("─", dividerWidth)))
}

func ThinDivider() {
	fmt.Println(Muted("  " + strings.Repeat("╌", dividerWidth)))
}

func Spacer() { fmt.Println() }

// PrintBanner prints the styled tool name header.
func PrintBanner() {
	fmt.Println()
	fmt.Println(Header("  git-tidy") + Muted("  — clean branch tool"))
	fmt.Println()
}

// PrintStep prints a numbered workflow step heading.
func PrintStep(n int, label string) {
	fmt.Printf("\n  %s  %s\n", Muted(fmt.Sprintf("step %d", n)), Bold(label))
	ThinDivider()
}

// PrintSuccess prints a final success block.
func PrintSuccess(lines ...string) {
	fmt.Println()
	Divider()
	for _, l := range lines {
		fmt.Printf("  %s  %s\n", IconOK(), l)
	}
	Divider()
	fmt.Println()
}

// PrintError prints a formatted error block to stderr.
func PrintError(heading string, details ...string) {
	fmt.Fprintf(os.Stderr, "\n  %s  %s\n", IconErr(), Danger(heading))
	for _, d := range details {
		fmt.Fprintf(os.Stderr, "     %s\n", Muted(d))
	}
	fmt.Fprintln(os.Stderr)
}

// PrintWarning prints a warning line.
func PrintWarning(s string) {
	fmt.Printf("  %s  %s\n", IconWarn(), Warning(s))
}

// PrintDryRun prints the dry-run mode notice.
func PrintDryRun() {
	fmt.Println()
	fmt.Printf("  %s %s %s\n",
		esc(ansiBold+ansiYellow, "["),
		Accent("DRY RUN — no changes will be made"),
		esc(ansiBold+ansiYellow, "]"),
	)
	fmt.Println()
}

// Truncate shortens s to max runes, appending "…" if cut.
func Truncate(s string, max int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max-1]) + "…"
}
