# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Checkpoint is a Go CLI tool that captures structured development history in git-tracked YAML files. It solves the problem of LLM-assisted development losing context between sessions by creating an append-only changelog linking every commit to its reasoning, decisions, and failed approaches.

## Build and Test Commands

```bash
# Build
make build              # Build to bin/checkpoint
make install            # Install to GOPATH/bin

# Test
make test               # Run all tests
go test -v ./...        # Verbose output
go test -v ./internal/schema/...  # Run single package tests
go test -run TestName ./...       # Run specific test

# Quality checks (run before checkpoint commit)
make check              # Run fmt, vet, lint, test

# Run
make run-with ARGS="start ."      # Run with arguments
./bin/checkpoint help             # After building
```

## Architecture

```
main.go              # CLI entry point, calls cmd.Execute()
cmd/                 # Cobra command handlers (one file per command)
  root.go            # Root command and Execute() function
internal/            # Core logic packages
  changelog/         # Append-only changelog operations
  schema/            # YAML schema definitions and validation
  explain/           # Project context rendering
  context/           # Checkpoint context handling
  git/               # Git operations wrapper
  language/          # Language detection
  templates/         # Embedded project templates
  prompts/           # Prompt loading from .checkpoint/prompts/
pkg/config/          # Configuration constants (file names, paths)
.checkpoint/         # Project's own checkpoint config
```

**Data flow:**
1. `checkpoint check` → creates `.checkpoint-input` + `.checkpoint-diff`
2. LLM/user fills `.checkpoint-input` with change descriptions
3. `checkpoint commit` → validates, appends to changelog, git commits
4. Tool backfills commit hash into last changelog document

## Key Patterns

**Adding a new command (Cobra):**
1. Create `cmd/{name}.go`
2. Define `var {name}Cmd = &cobra.Command{...}`
3. In `init()`, add command: `rootCmd.AddCommand({name}Cmd)`
4. Flags: use `{name}Cmd.Flags().BoolVar(...)` etc.

```go
var exampleOpts struct {
    json bool
}

func init() {
    rootCmd.AddCommand(exampleCmd)
    exampleCmd.Flags().BoolVar(&exampleOpts.json, "json", false, "Output as JSON")
}

var exampleCmd = &cobra.Command{
    Use:   "example [args]",
    Short: "Brief description",
    Run: func(cmd *cobra.Command, args []string) {
        // implementation
    },
}
```

**Error handling:**
```go
return fmt.Errorf("operation failed: %w", err)
// User-facing errors include hints:
// error: .checkpoint-input not found
// hint: Run 'checkpoint check' first
```

**Testing:**
- Unit tests: `{file}_test.go` in same package
- Integration tests needing git: `integration_test.go` (creates temp dirs)
- Table-driven tests preferred (see `internal/schema/schema_test.go`)

## Critical Rules

- Append-only semantics for changelog/context files (only commit_hash backfill exception)
- One checkpoint = one git commit
- Run `make check` before any checkpoint commit
- All user-facing errors need actionable hints
- Deps: stdlib + yaml + ulid + mcp-go + cobra

## File Types

**Git-tracked (permanent):**
- `.checkpoint-changelog.yaml` - Append-only changelog
- `.checkpoint-context.yml` - Decisions and reasoning
- `.checkpoint-project.yml` - Project patterns

**Not tracked (temporary):**
- `.checkpoint-input` - Edit during checkpoint
- `.checkpoint-diff` - Diff context for LLM
- `.checkpoint-status.yaml` - Last commit metadata
