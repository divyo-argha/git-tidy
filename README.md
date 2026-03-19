# git-tidy

> Clean up messy Git history and prepare branches for pull requests.

`git-tidy` is a production-grade CLI tool that lets you interactively select commits from your history and cherry-pick them into a clean branch — ready to open as a pull request.

---

## Install

### Prerequisites

- Go 1.21+
- Git installed and on your `$PATH`

### Build from source

```bash
git clone https://github.com/yourusername/git-tidy
cd git-tidy

# Download dependencies
go mod tidy

# Build the binary
make build
# → ./bin/git-tidy

# Install to /usr/local/bin so `git tidy` works as a subcommand
make install
```

### Manual install (no make)

```bash
go mod tidy
go build -o git-tidy .
mv git-tidy /usr/local/bin/git-tidy
```

> **How the git subcommand works**: Git automatically resolves `git tidy` to `git-tidy` if the binary is on your `$PATH`. No configuration needed.

---

## Usage

### `git tidy create [branch-name]`

The primary command. Shows recent commits, lets you pick which ones to keep, then creates a clean branch with only those commits applied.

```
git tidy create feature/my-clean-branch
```

If you omit the branch name, one is auto-generated: `tidy/20240315-143022`

**Flags:**
```
-n, --limit int   Number of commits to show (default 20)
```

**Example session:**

```
$ git tidy create feature/auth-fixes

  Recent commits
  ────────────────────────────────────────────────────────────────────────
  #    hash     subject                                              author           when
  ────────────────────────────────────────────────────────────────────────
   1)  a3f9c12  fix: correct JWT expiry check                       Alice            2 hours ago
   2)  b91d034  wip: half-done logging changes                      Alice            3 hours ago
   3)  c44e801  fix: handle nil user in auth middleware             Alice            5 hours ago
   4)  d02b998  chore: update .gitignore                            Bob              1 day ago
   5)  e77a123  feat: add password reset endpoint                   Alice            2 days ago
  ────────────────────────────────────────────────────────────────────────

  → Enter commit numbers to cherry-pick (e.g. 1,3,5 or 1-4 or 2,4-6)
  › 1,3,5

  ! Will cherry-pick 3 commit(s) into branch feature/auth-fixes:

      e77a123  feat: add password reset endpoint
      c44e801  fix: handle nil user in auth middleware
      a3f9c12  fix: correct JWT expiry check

  → Proceed? [y/N] y

  Creating branch feature/auth-fixes ...

    ✓  e77a123  feat: add password reset endpoint
    ✓  c44e801  fix: handle nil user in auth middleware
    ✓  a3f9c12  fix: correct JWT expiry check

  ────────────────────────────────────────────────────────────────────────
  ✓  Branch ready: feature/auth-fixes
  ✓  3 commit(s) applied
  ────────────────────────────────────────────────────────────────────────
```

---

### `git tidy log`

Show recent commits without doing anything. Useful for reviewing history.

```
git tidy log
git tidy log -n 50
```

**Flags:**
```
-n, --limit int   Number of commits to show (default 20)
```

---

## Selection syntax

| Input    | Meaning                        |
|----------|--------------------------------|
| `1,3,5`  | Commits 1, 3, and 5            |
| `1-4`    | Commits 1 through 4 inclusive  |
| `2,4-6`  | Commit 2, and commits 4, 5, 6  |

Commits are always applied oldest-first, regardless of selection order, so your branch history reads naturally.

---

## Error handling

| Situation                    | Behaviour                                                                 |
|-----------------------------|---------------------------------------------------------------------------|
| Not inside a git repo        | Exits immediately with a clear message                                    |
| Uncommitted changes          | Refuses to run — stash or commit first                                    |
| Branch already exists        | Exits before creating anything                                            |
| Invalid branch name          | Caught via `git check-ref-format` before any changes                     |
| Cherry-pick conflict         | Aborts the pick, deletes the partial branch, restores your original HEAD |
| Out-of-range commit number   | Re-prompts with a clear error                                             |
| `git` not on PATH            | Friendly error instead of a raw exec failure                              |

---

## Project structure

```
git-tidy/
├── main.go                      # Entrypoint — calls cmd.Execute()
├── cmd/
│   ├── root.go                  # Cobra root + subcommand registration
│   ├── create.go                # `git tidy create` — main workflow
│   └── log.go                   # `git tidy log` — display only
├── internal/
│   ├── domain/
│   │   └── commit.go            # Commit value object
│   ├── git/
│   │   ├── runner.go            # Single exec.Command choke point
│   │   ├── log.go               # Parse `git log` → []Commit
│   │   ├── branch.go            # Branch create / checkout / delete
│   │   └── cherry.go           # Cherry-pick loop + conflict handling
│   └── tui/
│       ├── selector.go          # Commit display + interactive prompts
│       └── styles.go            # ANSI colour helpers
└── pkg/
    └── errors/
        └── errors.go            # Typed error types
```

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI subcommand parsing |

No Git library — all Git operations shell out to the system `git` binary via `exec.Command`. This means git-tidy inherits your credential helpers, hooks, and config automatically.

---

## Future phases (not in MVP)

- `git tidy rebase` — interactive rebase with TUI
- `git tidy push` — push clean branch and open PR URL
- `git tidy stash` — stash-aware workflow
- Shell completions (`git tidy completion bash/zsh/fish`)
