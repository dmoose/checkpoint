# checkpoint

[![GO](https://github.com/dmoose/checkpoint/actions/workflows/ci.yml/badge.svg)](https://github.com/dmoose/checkpoint/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmoose/checkpoint.svg)](https://pkg.go.dev/github.com/dmoose/checkpoint)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmoose/checkpoint)](https://goreportcard.com/report/github.com/dmoose/checkpoint)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Portable development context that travels with your project.

## Why checkpoint exists

Development context gets lost. Decisions made during implementation, alternatives considered, failed approaches—all of it disappears into chat logs, IDE-specific storage, or forgotten conversations. When you switch tools, change LLM providers, or onboard new developers, you start from zero.

Checkpoint solves this by storing structured development history in git-tracked YAML files. The context lives in your repository, not in any external tool.

**What this means:**

- Switch from Claude to GPT to Gemini—your project context remains
- Change IDEs or coding assistants—history stays intact
- Onboard a new developer or LLM—they can read what was tried and why
- Revisit code months later—decisions and alternatives are documented

## What checkpoint captures

Each checkpoint links a git commit to structured metadata:

```yaml
changes:
  - summary: "Add rate limiting to API endpoints"
    change_type: "feature"
    scope: "api/middleware"

context:
  problem_statement: "API vulnerable to abuse without request limits"
  decisions_made:
    - decision: "Token bucket algorithm over sliding window"
      rationale: "Better burst handling, simpler implementation"
      alternatives_considered:
        - "Sliding window (rejected - memory overhead per client)"
        - "Fixed window (rejected - boundary spike issues)"
  failed_approaches:
    - approach: "Redis-based distributed rate limiting"
      why_failed: "Added infrastructure dependency for single-node deployment"

next_steps:
  - summary: "Add per-endpoint configurable limits"
    priority: "med"
```

This lives in `.checkpoint-changelog.yaml`—append-only, git-tracked, searchable.

## Installation

### From source

```bash
git clone https://github.com/dmoose/checkpoint.git
cd checkpoint
make install-user    # Installs to ~/.local/bin
```

Or with Go:

```bash
go install github.com/dmoose/checkpoint@latest
```

### Verify installation

```bash
checkpoint version
checkpoint doctor    # Check setup
```

## Quick start

```bash
# Initialize in your project
cd your-project
checkpoint init

# Start a session
checkpoint start     # Shows status and next steps

# Make changes to your code...

# Create a checkpoint
checkpoint check     # Generates input file
# Fill in .checkpoint-input (or have your LLM do it)
checkpoint commit    # Commits with structured metadata
```

## Commands

| Command | Purpose |
|---------|---------|
| `init` | Initialize checkpoint in a project |
| `start` | Begin session, show status and next steps |
| `plan` | Create planning session (.checkpoint-session.yaml) |
| `session` | View/manage current planning session |
| `check` | Generate input file for describing changes |
| `commit` | Validate input, append to changelog, git commit |
| `lint` | Validate input file before commit |
| `search <query>` | Search changelog and context history |
| `explain` | Show project context (patterns, tools, guidelines) |
| `doctor` | Verify checkpoint setup |

Run `checkpoint help` for the full command list.

## Files

**Git-tracked (permanent):**
- `.checkpoint-changelog.yaml` - Append-only changelog with all checkpoints
- `.checkpoint-context.yaml` - Accumulated decisions, patterns, failed approaches
- `.checkpoint-project.yaml` - Project-wide patterns and conventions
- `.checkpoint/` - Configuration, prompts, guides

**Not tracked (work-in-progress):**
- `.checkpoint-input` - Current checkpoint being edited
- `.checkpoint-diff` - Diff context for current checkpoint
- `.checkpoint-status.yaml` - Last commit metadata

## LLM integration

Checkpoint works with any LLM-assisted development workflow:

1. Run `checkpoint start` and share output with your LLM
2. Work on your task
3. Run `checkpoint check` when done
4. LLM reads `.checkpoint-input` and `.checkpoint-diff`, fills in the descriptions
5. Review, then run `checkpoint commit`

The LLM can reference project patterns via `checkpoint explain` and search history via `checkpoint search`.

### Session planning

For complex work:

```bash
checkpoint plan              # Create planning session
checkpoint session           # View current session
checkpoint session handoff   # Prepare context for next session
```

## Shell completion

```bash
# Bash
checkpoint completion bash >> ~/.bashrc

# Zsh
checkpoint completion zsh > "${fpath[1]}/_checkpoint"

# Fish
checkpoint completion fish > ~/.config/fish/completions/checkpoint.fish
```

## Documentation

**In-repo guides:**
- [Quickstart](docs/QUICKSTART.md) - Get productive in 5 minutes
- [User Guide](docs/USER-GUIDE.md) - Workflows, scenarios, best practices
- [LLM Integration](docs/LLM-INTEGRATION.md) - Configuring Claude, Cursor, Aider, etc.

**Built-in commands:**
```bash
checkpoint guide first-time-user  # Getting started
checkpoint guide llm-workflow     # LLM integration patterns
checkpoint examples               # Example checkpoints
```

## Development

```bash
make build          # Build to bin/
make test           # Run tests
make check          # Format, vet, lint, test
make install-user   # Install to ~/.local/bin
```

## License

Apache 2.0 - See [LICENSE](LICENSE)
