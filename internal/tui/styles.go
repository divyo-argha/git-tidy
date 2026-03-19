package tui

import "fmt"

// ANSI escape helpers — no external dependency needed for this minimal palette.
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	red    = "\033[31m"
	white  = "\033[97m"
)

func header(s string) string  { return bold + cyan + s + reset }
func success(s string) string { return bold + green + s + reset }
func warn(s string) string    { return yellow + s + reset }
func errStyle(s string) string { return bold + red + s + reset }
func muted(s string) string   { return dim + s + reset }
func highlight(s string) string { return bold + white + s + reset }

// Divider prints a visual separator line.
func Divider() {
	fmt.Println(muted("  " + repeat("─", 72)))
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
