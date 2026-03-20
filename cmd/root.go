package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/divyo-argha/git-tidy/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "git-tidy",
	Short: "Clean up messy Git history and prepare branches for pull requests",
	Long: tui.Header("git-tidy") + `  ` + tui.Muted("clean branch tool") + `

  Interactively select commits from your history and cherry-pick
  them into a clean branch — ready to open as a pull request.

  ` + tui.Bold("Commands:") + `
    ` + tui.Accent("create") + `     Select commits and cherry-pick into a new branch
    ` + tui.Accent("preview") + `    Dry-run of create — nothing is changed
    ` + tui.Accent("log") + `        Show recent commits on the current branch
    ` + tui.Accent("version") + `    Print version and build information

  ` + tui.Bold("Quick start:") + `
    git tidy preview              # see what will happen first
    git tidy create my-branch     # then execute

  ` + tui.Bold("Flags:") + `
    --version, -v                 # same as: git tidy version --short

  ` + tui.Muted("Docs: https://github.com/divyo-argha/git-tidy"),
	// --version / -v flag handled below via PersistentPreRunE.
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute is the entrypoint called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n  %s  %s\n\n", tui.IconErr(), err)
		os.Exit(1)
	}
}

func init() {
	// --version / -v on the root command (mirrors `git --version` convention).
	var showVersion bool
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false,
		"Print version information and exit")

	// Intercept --version before any subcommand runs.
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if showVersion {
			printVersion()
			os.Exit(0)
		}
		return nil
	}

	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newPreviewCmd())
	rootCmd.AddCommand(newLogCmd())
	rootCmd.AddCommand(newVersionCmd())
}
