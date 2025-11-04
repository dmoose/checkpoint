# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Commands

- Build
  - make build           # build bin/checkpoint with ldflags version
  - make build-dev       # build with -race for development
  - make build-all       # cross-compile to build/
  - make install         # go install with ldflags
  - make release         # tar.gz archives after build-all
- Run
  - make run                         # build then run
  - make run-with ARGS="start"        # run a subcommand
  - bin/checkpoint summary --json    # run built binary directly
- Lint and vet
  - make lint           # uses golangci-lint (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
  - make fmt            # go fmt ./...
  - make vet            # go vet ./...
  - make check          # fmt, vet, lint, test
- Tests
  - make test                          # go test -v ./...
  - make test-coverage                 # coverage.html
  - make test-race                     # race detector
  - Run a single package: go test -v ./cmd
  - Run a single test: go test -v ./cmd -run TestCommit
  - Run subtests with regex: go test -v ./internal/schema -run '^Test.*Validate'
  - Benchmarks: make bench or go test -bench=. -benchmem ./...
- Dependencies
  - make deps | make tidy | make update
- Clean
  - make clean | make clean-build

## Repository guardrails

- Treat checkpoint data files as read-only; never modify them without specific instructions:
  - .checkpoint-changelog.yaml
  - .checkpoint-context.yml
  - .checkpoint-project.yml
  - .checkpoint-status.yaml

## High-level architecture

- Entry point
  - main.go parses args/flags and dispatches to cmd package functions; version is injected via -ldflags (Makefile) and used by commit/init.
- CLI commands (cmd/)
  - start: readiness checks (git repo, checkpoint init, in-progress state, git status) and prints next steps from last status file.
  - check: generates .checkpoint-input (templated with git status and diff path) and .checkpoint-diff; guards with lock files.
  - lint: parses input and surfaces obvious issues before commit.
  - commit: validates input, appends a YAML document to changelog, stages/commits changes, then back-fills commit hash.
  - summary: prints recent activity; supports --json.
  - init: writes CHECKPOINT.md and seeds .checkpoint/ (examples, guides, prompts).
  - clean: removes input/diff to abort/restart.
  - examples | guide | prompt: read from .checkpoint/ to list/show curated examples, guides, and prompts with variable substitution.
- Core packages (internal/ and pkg/)
  - internal/schema: YAML data model for checkpoints; generation of input template; validation and linting; rendering persisted changelog docs; parsing git numstat.
  - internal/project: manages .checkpoint-project.yml (curated project context + appended recommendations); first-doc read/LLM render helpers.
  - internal/file, internal/git: filesystem and git helpers used by commands.
  - internal/prompts: loads .checkpoint/prompts/prompts.yaml, lists prompts, loads templates, substitutes variables.
  - internal/language and internal/context: language detection and additional context structures used in templates.
  - pkg/config: canonical filenames for all checkpoint artifacts.
- Data artifacts written in user repos
  - Tracked: .checkpoint-changelog.yaml (multi-document YAML; first doc is meta), .checkpoint-context.yml, .checkpoint-project.yml, .checkpoint/ (library content)
  - Untracked/temporary: .checkpoint-input, .checkpoint-diff, .checkpoint-status.yaml, .checkpoint-lock
- Typical workflow
  - checkpoint start → checkpoint check → edit .checkpoint-input → checkpoint lint (optional) → checkpoint commit → checkpoint summary
- Notable dependencies and versions
  - go.mod: module go-llm; Go 1.25.1; deps: gopkg.in/yaml.v3, github.com/oklog/ulid/v2