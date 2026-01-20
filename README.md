# checkpoint + guardrail

[![GO](https://github.com/dmoose/checkpoint/actions/workflows/ci.yml/badge.svg)](https://github.com/dmoose/checkpoint/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmoose/checkpoint.svg)](https://pkg.go.dev/github.com/dmoose/checkpoint)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmoose/checkpoint)](https://goreportcard.com/report/github.com/dmoose/checkpoint)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Portable development context that travels with your project.

## Why this exists

Development context gets lost. Decisions, failed approaches, and reasoning disappear into chat logs or IDE-specific storage. When you switch LLM providers, change tools, or return to code months later, you start from zero.

Checkpoint and guardrail solve this by storing structured development history and project knowledge in git-tracked YAML files. The context lives in your repository, not in any external tool.

## Two tools, one knowledge base

This project provides two independent CLI binaries that share the same `.checkpoint/` directory:

**checkpoint** records what happened. It captures structured change history linked to git commits -- what changed, why, what was tried, and what comes next.

**guardrail** manages what the LLM knows. It provides project context, coding guidelines, tool commands, and skills to any LLM-assisted workflow.

Together they create a feedback loop: guardrail gives the LLM context before work begins, checkpoint captures context after work completes.

## What gets captured

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
  failed_approaches:
    - approach: "Redis-based distributed rate limiting"
      why_failed: "Added infrastructure dependency for single-node deployment"

next_steps:
  - summary: "Add per-endpoint configurable limits"
    priority: "med"
```

## Installation

### From source

```bash
git clone https://github.com/dmoose/checkpoint.git
cd checkpoint
make install-user    # Installs both binaries to ~/.local/bin
```

### With Go

```bash
go install github.com/dmoose/checkpoint/cmd/checkpoint@latest
go install github.com/dmoose/checkpoint/cmd/guardrail@latest
```

### Verify

```bash
checkpoint version
guardrail doctor
```

## Quick start

```bash
cd your-project

# Initialize both tools
checkpoint init      # Creates changelog, context files
guardrail init       # Creates .checkpoint/ config directory

# Start a session
checkpoint start     # Shows status, next steps, AI boundaries

# Make changes to your code...

# Record the checkpoint
checkpoint check     # Generates input file from git diff
# Fill in checkpoint-input (or have your LLM do it)
checkpoint commit    # Validates, appends to changelog, git commits
```

## Commands

| checkpoint | guardrail |
|-----------|-----------|
| `init` -- Initialize changelog files | `init` -- Create .checkpoint/ config |
| `start` -- Begin session, show status | `explain` -- Show project context for LLMs |
| `plan` -- Create planning session | `learn` -- Capture guidelines, patterns, anti-patterns |
| `session` -- View/manage planning session | `skill` -- Manage LLM skill definitions |
| `check` -- Generate input file for changes | `doctor` -- Verify knowledge base health |
| `commit` -- Validate and git commit | `prompt` -- Manage prompt templates |
| `lint` -- Validate input before commit | `guide` -- Show built-in guides |
| `search` -- Search changelog history | `examples` -- Show example checkpoints |
| `summary` -- Show changelog summary | `config` -- Get/set configuration values |
| `clean` -- Remove temporary files | `completion` -- Shell completions |
| `completion` -- Shell completions | |

## Files

**Git-tracked (permanent):**
- `.checkpoint-changelog.yaml` -- Append-only changelog with all checkpoints
- `.checkpoint-context.yaml` -- Accumulated decisions, patterns, failed approaches
- `.checkpoint/` -- Configuration: project, tools, guidelines, skills

**Not tracked (work-in-progress):**
- `checkpoint-input` -- Current checkpoint being edited
- `.checkpoint-diff` -- Diff context for current checkpoint
- `.checkpoint-status.yaml` -- Last commit metadata
- `.checkpoint-session.yaml` -- Current planning session

## Documentation

- [Getting Started](docs/getting-started.md) -- Installation and first checkpoint
- [Workflow](docs/workflow.md) -- Daily workflow for humans and LLMs
- [LLM Integration](docs/llm-integration.md) -- Configuring Claude, Cursor, Aider, etc.
- [Schema Reference](docs/schema-reference.md) -- Complete YAML schema documentation

## License

Apache 2.0 -- See [LICENSE](LICENSE)
