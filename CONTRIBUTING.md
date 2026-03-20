# Contributing to git-tidy

Thank you for taking the time to contribute. This guide covers everything you need to get up and running.

---

## Development setup

**Requirements:** Go 1.21+, Git 2.30+, Make

```bash
git clone https://github.com/divyo-argha/git-tidy
cd git-tidy
go mod tidy
make build
./bin/git-tidy --help
```

Run the full check suite before opening a PR:

```bash
make check   # vet + fmt
make test    # unit tests with race detector
```

---

## Project layout

```
git-tidy/
├── main.go                      # Entrypoint — injects build info, calls cmd.Execute()
├── cmd/                         # Cobra commands (one file per subcommand)
│   ├── root.go                  # Root command, --version flag
│   ├── create.go                # git tidy create
│   ├── preview.go               # git tidy preview
│   ├── log.go                   # git tidy log
│   └── version.go               # git tidy version
├── internal/
│   ├── domain/commit.go         # Commit value object — no dependencies
│   ├── git/
│   │   ├── runner.go            # exec.Command choke point — ALL git calls go here
│   │   ├── safety.go            # IsRepo, IsClean, DirtyFiles, duplicate detection, stash
│   │   ├── log.go               # git log → []Commit
│   │   ├── branch.go            # Branch create/checkout/delete, ValidateBranchName
│   │   ├── cherry.go            # Cherry-pick loop with progress callback
│   │   └── rollback.go          # RollbackJournal — state capture and restoration
│   └── tui/
│       ├── styles.go            # ANSI color system, layout helpers, NO_COLOR support
│       ├── selector.go          # Commit table, SelectCommits, ConfirmSelection
│       └── prompt.go            # Interactive safety prompts (dirty tree, duplicates, conflict)
└── pkg/
    └── errors/errors.go         # Typed error types — all public, used across packages
```

**Architectural rules:**
- `internal/git/` must never import `internal/tui/` — no circular concern
- `internal/domain/` has zero dependencies
- `pkg/errors/` may only import stdlib
- `cmd/` is the only place `internal/git/` and `internal/tui/` are wired together

---

## Adding a new command

1. Create `cmd/<name>.go` with a `newXxxCmd() *cobra.Command` function
2. Register it in `cmd/root.go` `init()`
3. Add git layer functions to the appropriate `internal/git/*.go` file
4. Add any new error types to `pkg/errors/errors.go`
5. Update `README.md` and `CHANGELOG.md`

---

## Commit message convention

```
<type>: <short description>

type = feat | fix | refactor | docs | test | chore
```

Examples:
```
feat: add git tidy squash command
fix: handle empty repo in log command
docs: add fish shell completion example
```

---

## Pull request checklist

- [ ] `make check` passes (vet + fmt)
- [ ] `make test` passes
- [ ] New behaviour documented in `README.md`
- [ ] `CHANGELOG.md` updated under `[Unreleased]`
- [ ] Commit messages follow the convention above

---

## Reporting bugs

Open an issue with:
1. Your OS and `git tidy version` output
2. The command you ran
3. The full terminal output (paste as text, not a screenshot)
4. What you expected to happen

---

## License

By contributing you agree that your contributions will be licensed under the MIT License.
