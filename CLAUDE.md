# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

```bash
# Build
make build

# Run tests
make test                 # All tests
make test-unit           # Unit tests only (fast, no git required)
make test-integration    # Integration tests (requires git)

# Install to GOPATH/bin + shell integration
make install

# Lint and format
make lint
make fmt

# Coverage
make coverage            # Generates coverage.html
```

To run a specific test: `go test -v -run TestSpecificName ./...`

## Project Architecture

ccswitch is a CLI tool for managing git worktrees through a "session" abstraction.

### Layer Architecture

```
main.go (entry point)
  └── cmd/              # Cobra command layer
      ├── root.go       # Command registration
      └── *.go          # Individual commands (create, switch, cleanup, rebase, fanout, etc.)
          └── internal/ # Business logic layer
              ├── session/   # Session Manager (orchestrates operations)
              ├── git/       # Git operations (worktree, branch, commit, rebase)
              ├── config/    # YAML configuration (~/.ccswitch/config.yaml)
              ├── ui/        # Colored output helpers
              ├── errors/    # Error types with hints
              └── utils/     # Utilities (slugify, shell detection)
```

### Key Concepts

**Sessions vs Worktrees**
- A "worktree" is a git concept: a separate working directory for a branch
- A "session" is ccswitch's abstraction: worktrees stored under `~/.ccswitch/worktrees/{repo-name}/{session-name}/`
- The `GetSessionsFromWorktrees()` function in `internal/git/worktree.go` filters git's worktree list to identify ccswitch-managed sessions
- Some commands (like `rebase`, `fanout`) operate on **all** git worktrees, not just ccswitch sessions

**Session Manager** (`internal/session/manager.go`)
- Orchestrates operations across WorktreeManager and BranchManager
- `repoPath` is the current working directory (or main repo path for worktree listing)
- `mainRepoPath` (derived via `GetMainRepoPath()`) is used for worktree operations
- Always use `WorktreeManager.Create()` with paths under `~/.ccswitch/worktrees/` for new sessions

**Git Operations Pattern**
All git operations use `exec.Command` with `cmd.Dir` set to the target repository path:
```go
cmd := exec.Command("git", "subcommand", "args")
cmd.Dir = repoPath  // Critical: specifies which worktree/repo to operate on
output, err := cmd.CombinedOutput()
```

**Shell Integration**
- The `shell-init` command outputs shell code for eval
- The bash wrapper captures command output and extracts paths for `cd`-ing
- Commands output paths to stdout as the last line: `fmt.Printf("\ncd %s\n", path)`

### Configuration

Config stored in `~/.ccswitch/config.yaml` with defaults in `internal/config/config.go`:
- `branch.prefix`: Default branch name prefix (default: "feature/")
- `git.default_branch`: Main branch name (default: "main")
- `ui.show_emoji`: Toggle emoji output

### Branch Naming

Session descriptions are converted to branch names using `utils.Slugify()`:
- Converts to lowercase, replaces spaces with hyphens, removes special chars
- Branch name: `{config.Branch.Prefix}{slugify(description)}`
- Session name (directory): `{slugify(description)}`

### Error Handling

Errors from `internal/errors/` include hints via `ErrorHint(err)`. Wrap errors with context:
```go
return fmt.Errorf("failed to do something: %w", err)
```

### Recent Commands

- `rebase`: Commit changes in a worktree and rebase to current branch. Checks for uncommitted changes first - only prompts for commit message if changes exist.
- `fanout`: Propagate current branch commits to all other worktrees. Safety checks: no uncommitted changes, no worktree ahead of current.
