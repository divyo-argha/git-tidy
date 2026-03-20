package main

import "github.com/yourusername/git-tidy/cmd"

// Build-time variables injected via -ldflags.
// Defaults to "dev" / "none" / "unknown" for local builds.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetBuildInfo(version, commit, date)
	cmd.Execute()
}
