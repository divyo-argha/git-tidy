package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/divyo-argha/git-tidy/internal/tui"
)

// buildInfo holds version metadata injected at link time.
type buildInfo struct {
	Version string
	Commit  string
	Date    string
}

var build = buildInfo{
	Version: "dev",
	Commit:  "none",
	Date:    "unknown",
}

// SetBuildInfo is called from main() with values injected by -ldflags.
func SetBuildInfo(version, commit, date string) {
	build.Version = version
	build.Commit = commit
	build.Date = date
}

func newVersionCmd() *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print the version, git commit, build date, and runtime information.",
		Run: func(cmd *cobra.Command, args []string) {
			if short {
				fmt.Println(build.Version)
				return
			}
			printVersion()
		},
	}

	cmd.Flags().BoolVarP(&short, "short", "s", false, "Print only the version number")
	return cmd
}

func printVersion() {
	fmt.Println()
	fmt.Printf("  %s  %s\n\n", tui.Header("git-tidy"), tui.Muted("— clean branch tool"))
	tui.ThinDivider()
	fmt.Printf("  %-12s  %s\n", tui.Bold("Version"), tui.Accent(build.Version))
	fmt.Printf("  %-12s  %s\n", tui.Bold("Commit"), tui.Muted(build.Commit))
	fmt.Printf("  %-12s  %s\n", tui.Bold("Built"), tui.Muted(build.Date))
	fmt.Printf("  %-12s  %s\n", tui.Bold("Go"), tui.Muted(runtime.Version()))
	fmt.Printf("  %-12s  %s\n", tui.Bold("Platform"), tui.Muted(runtime.GOOS+"/"+runtime.GOARCH))
	tui.ThinDivider()
	fmt.Println()
}
