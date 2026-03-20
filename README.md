<div align="center">

# git-tidy

**Clean up messy Git history. Prepare perfect pull requests.**

[![CI](https://github.com/yourusername/git-tidy/actions/workflows/ci.yml/badge.svg)](https://github.com/yourusername/git-tidy/actions/workflows/ci.yml)
[![Release](https://github.com/yourusername/git-tidy/actions/workflows/release.yml/badge.svg)](https://github.com/yourusername/git-tidy/actions/workflows/release.yml)
[![Go version](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

---

`git-tidy` is a CLI tool that lets you interactively select commits from your history and cherry-pick them into a clean branch — ready to open as a pull request. Works as a native Git subcommand: `git tidy`.

```
  git-tidy  — clean branch tool

  step 1  Checking repository state
  ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
  ✓  On branch main — working tree clean

  step 2  Validating target branch
  ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
  ✓  New branch: feature/auth-fixes

  step 3  Select commits
  ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌

  Recent commits  (5 shown)
  ────────────────────────────────────────────────────────────────────────
   #     hash     subject                                  author    when
  ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
   1  ◇  a3f9c12  fix: correct JWT expiry check            Alice     2h ago
   2  ◇  b91d034  wip: half-done logging changes           Alice     3h ago
   3  ◇  c44e801  fix: handle nil user in auth middleware  Alice     5h ago
   4  ~  d02b998  chore: update .gitignore                 Bob       1d ago
   5  ◇  e77a123  feat: add password reset endpoint        Alice     2d ago
  ────────────────────────────────────────────────────────────────────────

  → Select commits to cherry-pick
      Enter numbers, ranges, or "all"  e.g.  1,3,5   2-6   all
  › 1,3,5
```

---

## Table of contents

- [Why git-tidy](#why-git-tidy)
- [Installation](#installation)
  - [One-line install (Linux / macOS)](#one-line-install)
  - [Build from source](#build-from-source)
  - [Manual install](#manual-install)
- [Commands](#commands)
  - [git tidy create](#git-tidy-create)
  - [git tidy preview](#git-tidy-preview)
  - [git tidy log](#git-tidy-log)
  - [git tidy version](#git-tidy-version)
- [Safety features](#safety-features)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [License](#license)

---

## Why git-tidy

You're working on a feature and your branch looks like this:

```
a3f9c12  fix: correct JWT expiry check
b91d034  wip: debugging, remove before merge  ← noise
c44e801  fix: handle nil user in auth middleware
d02b998  fixup! forgot to add import          ← noise
e77a123  feat: add password reset endpoint
f112233  WIP WIP WIP                           ← noise
```

You want a PR that contains only the three real commits — but `git rebase -i` is error-prone when you're mid-feature, and `git cherry-pick` requires you to copy hashes manually.

`git tidy create feature/auth-fixes` lets you point at the commits you want, then handles the branch creation and cherry-picks — with full rollback if anything conflicts.

---

## Installation

### One-line install

```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/git-tidy/main/install.sh | bash
```

**Options:**

```bash
# Install a specific version
curl -fsSL .../install.sh | bash -s -- --version v0.3.0

# Install to a custom directory (e.g. no sudo needed)
curl -fsSL .../install.sh | bash -s -- --dir ~/.local/bin

# Preview what would happen without installing
curl -fsSL .../install.sh | bash -s -- --dry-run
```

The installer:
- Detects your OS and architecture (Linux/macOS, amd64/arm64)
- Downloads the correct binary from GitHub Releases
- Verifies the SHA-256 checksum
- Installs to `/usr/local/bin` (or `--dir`), using `sudo` only if needed
- Checks that the install directory is on your `$PATH`

### Build from source

**Requirements:** Go 1.21+, Git, Make

```bash
git clone https://github.com/yourusername/git-tidy
cd git-tidy
go mod tidy
make install          # builds and copies to /usr/local/bin
```

Install to a custom directory:

```bash
make install INSTALL_DIR=~/.local/bin
```

### Manual install

```bash
# Download the binary for your platform from:
# https://github.com/yourusername/git-tidy/releases

# Linux amd64
curl -fsSL https://github.com/yourusername/git-tidy/releases/latest/download/git-tidy-linux-amd64 \
  -o /usr/local/bin/git-tidy && chmod +x /usr/local/bin/git-tidy

# macOS Apple Silicon
curl -fsSL https://github.com/yourusername/git-tidy/releases/latest/download/git-tidy-darwin-arm64 \
  -o /usr/local/bin/git-tidy && chmod +x /usr/local/bin/git-tidy
```

### Verify installation

```bash
git tidy version
# git-tidy
# Version  v0.3.0
# Commit   a3f9c12
# Built    2024-03-15T10:00:00Z
# Go       go1.21.6
# Platform linux/amd64
```

### How the git subcommand works

Git automatically resolves `git tidy` to `git-tidy` when a binary named `git-tidy` is on your `$PATH`. No aliases or git config needed.

---

## Commands

### `git tidy create`

The primary command. Shows recent commits, lets you select which to keep, then creates a clean branch with only those commits applied.

```bash
git tidy create [branch-name] [flags]

Flags:
  -n, --limit int   Number of commits to display (default 20)
```

**Examples:**

```bash
# Auto-generate a branch name: tidy/20240315-143022
git tidy create

# Use a specific branch name
git tidy create feature/auth-cleanup

# Show more commits to choose from
git tidy create fix/payments --limit 50
```

**Selection syntax:**

| Input    | Selects                        |
|----------|--------------------------------|
| `1,3,5`  | Commits 1, 3, and 5            |
| `2-6`    | Commits 2 through 6 inclusive  |
| `1,3-5`  | Commit 1, and commits 3, 4, 5  |
| `all`    | Every commit shown             |

Commits are always applied **oldest-first**, regardless of how you enter them, so the branch history reads naturally.

---

### `git tidy preview`

A full dry-run of `create`. Runs every step — preflight checks, branch name validation, interactive commit selection, execution plan — but **stops before touching the repo**.

```bash
git tidy preview [branch-name] [flags]

Flags:
  -n, --limit int   Number of commits to display (default 20)
```

Safe to run any time. Use it to verify your selection before committing to a `create` run.

```bash
git tidy preview feature/my-branch
# … review the plan …
git tidy create feature/my-branch   # execute when ready
```

Preview differences from `create`:
- Dirty working tree is a **warning**, not a blocker
- Branch existence is a **warning** with a note, not an error
- Ends with the exact command to run: `Ready to execute? Run: git tidy create feature/my-branch`

---

### `git tidy log`

Display recent commits without taking any action.

```bash
git tidy log [flags]

Flags:
  -n, --limit int   Number of commits to show (default 20)
```

```bash
git tidy log
git tidy log -n 50
```

---

### `git tidy version`

Print version and build information.

```bash
git tidy version [flags]
git tidy --version          # same, shorter

Flags:
  -s, --short   Print only the version number (useful in scripts)
```

```bash
git tidy version
git tidy version --short    # → v0.3.0
git tidy --version          # same as version --short
```

---

## Safety features

git-tidy runs several safety checks automatically before making any changes.

### Detached HEAD detection

If the repo is in detached HEAD state, `create` refuses to run with a clear message and the correct `git checkout` command to fix it. Cherry-picking onto a detached HEAD would silently produce unreachable commits.

### Uncommitted changes prompt

Instead of blocking, git-tidy offers a choice:

```
  !  Uncommitted changes detected

   M  src/auth/jwt.go
  ??  notes.txt

  → How would you like to proceed?

    [s]  Auto-stash changes, create branch, then restore them
    [a]  Abort — I'll handle this manually
```

Choosing `[s]` runs `git stash push`, records the stash ref in the rollback journal, and pops it automatically when the command finishes (or on any error).

### Duplicate commit detection

Before showing the selection table, git-tidy checks each commit against the current branch using `git cherry` (patch-ID comparison, not just hash). Commits already applied are marked with `~` and dimmed:

```
   4  ~  d02b998  chore: update .gitignore   ← already on this branch
```

If you select a marked commit anyway, a second prompt gives you the option to skip duplicates and continue with the rest.

### Branch name validation

Branch names are validated with explicit human-readable rules before any git call:

```
  ✗  Invalid branch name: "my branch"
     spaces are not allowed — use hyphens or slashes
     Examples of valid names: feature/login  fix/null-check  tidy/cleanup
```

### Rollback journal

A `RollbackJournal` records repository state at the start of every `create` run:

- The branch you were on (`OriginBranch`)
- The branch created (`CreatedBranch`)  
- Any auto-stash ref (`StashRef`)

If anything fails, rollback runs four steps in order:

1. `git cherry-pick --abort` (idempotent)
2. `git checkout <OriginBranch>`
3. `git branch -D <CreatedBranch>` (remove partial branch)
4. `git stash pop <StashRef>` (restore your changes)

The repo is always left in the same state it was in before you ran the command.

### Guided conflict resolution

When a cherry-pick conflicts, git-tidy doesn't just print an error — it shows:

```
  ✗  Cherry-pick conflict — rolling back

  Commit      a3f9c12  fix: correct JWT expiry check
  Progress    2 of 5 applied successfully

  → Files that would have conflicted:
      ✖  src/auth/jwt.go
      ✖  src/middleware/auth.go

  What happened:  The selected commit conflicts with changes already in the base branch.

  Next steps:
    1.  Resolve the conflict on your current branch:
           git merge origin/main
    2.  Then re-run:
           git tidy create feature/auth-fixes
    3.  Or cherry-pick manually:
           git cherry-pick a3f9c12
```

---

## Configuration

No configuration file is needed. All behaviour is controlled via flags.

**Environment variables:**

| Variable             | Effect                             |
|----------------------|------------------------------------|
| `NO_COLOR`           | Disable all ANSI colours           |
| `GIT_TIDY_NO_COLOR`  | Same, scoped to git-tidy only      |

git-tidy also respects your existing git configuration — credential helpers, hooks, `.gitconfig` settings — because all git operations shell out to the system `git` binary.

---

## Uninstall

```bash
make uninstall                        # if installed via make
make uninstall INSTALL_DIR=~/.local/bin   # custom dir

# Or directly:
rm /usr/local/bin/git-tidy
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development setup, architecture guide, commit conventions, and PR checklist.

```bash
git clone https://github.com/yourusername/git-tidy
cd git-tidy && go mod tidy
make check && make test
```

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for the full version history.

---

## License

[MIT](LICENSE) — © 2024 yourusername
