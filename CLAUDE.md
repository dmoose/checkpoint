# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repo produces two binaries from one Go module:

- **checkpoint** — captures structured development history in git-tracked YAML files. Append-only changelog linking every commit to its reasoning, decisions, and failed approaches.
- **guardrail** — manages the living knowledge base that LLMs consume: project patterns, guidelines, skills, prompts, tools config. The guardrails that keep LLM agents on track.

Both share the same `.checkpoint/` data directory and internal packages but are fully independent binaries.

## Build and Test Commands

```bash
# Build
make build                # Build both binaries to bin/
make build-checkpoint     # Build only checkpoint
make build-guardrail      # Build only guardrail
make install              # Install both to GOPATH/bin

# Test
make test                 # Run all tests
go test -v ./...          # Verbose output
go test -v ./internal/schema/...  # Run single package tests
go test -run TestName ./...       # Run specific test

# Quality checks
make check                # Run fmt, vet, lint, test

# Run
make run-checkpoint ARGS="start"    # Run checkpoint with arguments
make run-guardrail ARGS="explain"   # Run guardrail with arguments
```

## Architecture

```
cmd/
  checkpoint/main.go      # checkpoint binary entry point
  guardrail/main.go       # guardrail binary entry point
internal/
  app/
    checkpoint/            # Cobra commands for checkpoint binary
      root.go              # rootCmd, Execute(), BuildInfo
      init.go, start.go, check.go, commit.go, clean.go
      search.go, summary.go, lint.go, plan.go, session.go
      completion.go
    guardrail/             # Cobra commands for guardrail binary
      root.go              # rootCmd, Execute(), BuildInfo
      init.go, explain.go, learn.go, skill.go, doctor.go
      prompt.go, guide.go, examples.go, config.go
      completion.go
  changelog/               # Append-only changelog operations
  schema/                  # YAML schema definitions and validation
  explain/                 # Project context rendering (shared types)
  context/                 # Checkpoint context handling
  git/                     # Git operations wrapper
  project/                 # Project file management
  language/                # Language detection
  detect/                  # Project auto-detection
  templates/               # Embedded project templates
  prompts/                 # Prompt loading from .checkpoint/prompts/
  guides/                  # Embedded guide content
  skills/                  # Embedded default skills
  file/                    # File I/O utilities
pkg/config/                # Configuration constants (file names, paths)
.checkpoint/               # Project's own checkpoint config
```

**Data flow (checkpoint side):**
1. `checkpoint check` → creates `checkpoint-input` + `.checkpoint-diff`
2. LLM/user fills `checkpoint-input` with change descriptions
3. `checkpoint commit` → validates, appends to changelog, git commits
4. Tool backfills commit hash into last changelog document

**Data flow (guardrail side):**
- `guardrail init` → creates knowledge base files in `.checkpoint/`
- `guardrail explain` → renders project context for LLMs
- `guardrail learn` → captures guidelines, patterns, anti-patterns
- `guardrail skill` → manages skill definitions

## Key Patterns

**Adding a new command:**
1. Decide which binary it belongs to (history → checkpoint, knowledge → guardrail)
2. Create file in `internal/app/checkpoint/` or `internal/app/guardrail/`
3. Define command with `rootCmd.AddCommand()` in `init()`
4. Package must be `checkpoint` or `guardrail` respectively

```go
// internal/app/checkpoint/example.go
package checkpoint

var exampleCmd = &cobra.Command{
    Use:   "example [args]",
    Short: "Brief description",
    Run: func(cmd *cobra.Command, args []string) {
        // implementation
    },
}

func init() {
    rootCmd.AddCommand(exampleCmd)
}
```

**Error handling:**
```go
return fmt.Errorf("operation failed: %w", err)
// User-facing errors include hints:
// error: checkpoint-input not found
// hint: Run 'checkpoint check' first
```

**Testing:**
- Unit tests: `{file}_test.go` in same package
- Integration tests: `integration_test.go` at repo root
- Table-driven tests preferred (see `internal/schema/schema_test.go`)

## Critical Rules

- Append-only semantics for changelog/context files (only commit_hash backfill exception)
- One checkpoint = one git commit
- Run `make check` before any checkpoint commit
- All user-facing errors need actionable hints
- Deps: stdlib + yaml + ulid + cobra
- checkpoint and guardrail binaries must be independent — no runtime dependency on each other

## File Types

**Git-tracked (permanent):**
- `.checkpoint-changelog.yaml` - Append-only changelog
- `.checkpoint-context.yaml` - Decisions and reasoning
- `.checkpoint/project.yaml` - Project patterns (with appended recommendations)

**Not tracked (temporary):**
- `checkpoint-input` - Edit during checkpoint
- `.checkpoint-diff` - Diff context for LLM
- `.checkpoint-status.yaml` - Last commit metadata
