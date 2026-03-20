# Changelog

All notable changes to git-tidy are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions follow [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

## [0.3.0] — 2024-03-15

### Added
- Safety mechanisms: detached HEAD detection, uncommitted changes prompt with auto-stash option
- Rollback journal: complete state restoration on any failure (branch deleted, stash restored)
- Duplicate commit detection using `git cherry` (patch-ID comparison, not just hash)
- Guided cherry-pick conflict resolution with per-file conflict list and exact recovery commands
- Branch name validator with human-readable rule violations before any git call
- `git tidy version --short` flag for scripting

### Changed
- `DirtyWorkDirError` now carries the full list of dirty files for rich display
- `CherryPickConflictError` now includes `ConflictFiles`, `Applied`, and `Total` fields
- `ShowCommits` accepts a `duplicates` map to dim and mark already-applied commits in the table
- `SelectCommits` re-prompts on invalid input instead of returning an error

## [0.2.0] — 2024-03-10

### Added
- `git tidy preview` dry-run command — shows full execution plan without touching the repo
- `git tidy version` subcommand with Go runtime and platform info
- `--version` / `-v` flag on the root command
- Color system rebuilt: `NO_COLOR` env var support, TTY detection, semantic style functions
- Selected commits re-rendered highlighted in the table before confirmation
- `[s]` stash / `[a]` abort interactive prompt when dirty tree detected
- Progress bar during cherry-pick batch
- `all` keyword in commit selection

### Changed
- Confirmation step now shows a tree-structured apply order (`├─` / `└─`)
- `ShowCommits` column layout improved with per-column semantic colors
- Error messages upgraded to heading + indented detail lines

## [0.1.0] — 2024-03-01

### Added
- `git tidy create [branch-name]` — interactive commit selection and cherry-pick
- `git tidy log [-n N]` — display recent commits
- Modular architecture: `internal/git`, `internal/tui`, `internal/domain`, `pkg/errors`
- Typed error system: `NotARepoError`, `DirtyWorkDirError`, `BranchExistsError`, etc.
- Range syntax in commit selection (`1-4`, `1,3,5`, or mixed)
- Auto-generated branch name `tidy/<timestamp>` when no name is provided
- Cobra-based CLI with subcommand help

[Unreleased]: https://github.com/divyo-argha/git-tidy/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/divyo-argha/git-tidy/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/divyo-argha/git-tidy/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/divyo-argha/git-tidy/releases/tag/v0.1.0
