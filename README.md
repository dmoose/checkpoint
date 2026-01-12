# checkpoint

Structured development history that you own.

## The Problem

LLM-assisted development loses context between sessions. Your decisions, failed approaches, and project patterns vanish when the conversation ends. That knowledge gets locked in chat histories you can't search, IDE plugins you might switch away from, or just forgotten.

## The Solution

Checkpoint captures **what changed and why** in git-tracked YAML files that travel with your project. Not locked in any tool. Not dependent on any LLM provider. Just structured data in your repo that any future LLM, IDE, or developer can read.

**What you get:**
- Append-only changelog linking every commit to its reasoning
- Decision history with alternatives considered and why they were rejected
- Failed approaches so you don't repeat mistakes
- Project patterns that emerge over time
- Next steps that persist across sessions

**Tool independence:** Switch LLMs, switch IDEs, switch developers—the context stays with the project.

## How It Works

Each checkpoint = one git commit + structured metadata:

```yaml
changes:
  - summary: "Add JWT authentication"
    change_type: "feature"
    scope: "auth"
context:
  problem_statement: "Need stateless auth for horizontal scaling"
  decisions_made:
    - decision: "Use JWT over sessions"
      rationale: "Stateless, no shared session store needed"
      alternatives_considered:
        - "Redis sessions (rejected - added infrastructure)"
  failed_approaches:
    - approach: "Tried passport.js"
      why_failed: "Too much magic, hard to customize"
next_steps:
  - summary: "Add refresh token rotation"
    priority: "high"
```

This lives in `.checkpoint-changelog.yaml`—append-only, git-tracked, searchable.

## Installation

```bash
go install github.com/dmoose/checkpoint@latest
```

Or build from source:
```bash
git clone https://github.com/dmoose/checkpoint.git
cd checkpoint
go build -o checkpoint .
```

## Quick Start

```bash
checkpoint init          # Initialize (auto-detects language/tools)
checkpoint doctor        # Verify setup

# Work session
checkpoint start         # See status and next steps
# ... make changes ...
checkpoint check         # Generate input file
# Edit .checkpoint-input with what changed and why
checkpoint commit        # Commit with metadata
```

## Commands

| Command | Purpose |
|---------|---------|
| `start` | Begin session, see next steps from last checkpoint |
| `check` | Generate input file for describing changes |
| `commit` | Commit with structured metadata |
| `explain` | Get project context (tools, guidelines, patterns) |
| `explain history` | View recent decisions and patterns |
| `search <query>` | Search changelog and context |
| `learn "insight"` | Capture knowledge mid-session |
| `session handoff` | Generate context doc for next LLM |

Run `checkpoint help` for full command list.

## Files

**Git-tracked (permanent):**
- `.checkpoint-changelog.yaml` - Append-only changelog
- `.checkpoint-context.yml` - Decisions, patterns, failed approaches
- `.checkpoint-project.yml` - Project-wide patterns (human-curated)
- `.checkpoint/` - Config, prompts, examples, guides

**Not tracked (temporary):**
- `.checkpoint-input` - Edit during checkpoint
- `.checkpoint-diff` - Diff context for LLM

## LLM Workflow

1. Start session: share `checkpoint start` output with LLM
2. Work: LLM makes changes
3. Checkpoint: run `checkpoint check`, LLM fills `.checkpoint-input`
4. Commit: run `checkpoint commit`
5. Handoff: run `checkpoint session handoff` for next session

The LLM reads project context via `checkpoint explain` and learns patterns from history via `checkpoint search`.

## Shell Completion

```bash
# Bash
checkpoint completion bash >> ~/.bashrc

# Zsh
checkpoint completion zsh > ~/.oh-my-zsh/completions/_checkpoint

# Fish
checkpoint completion fish > ~/.config/fish/completions/checkpoint.fish
```

## Documentation

```bash
checkpoint guide first-time-user  # Getting started
checkpoint guide llm-workflow     # LLM integration patterns
checkpoint examples               # Example checkpoints
```

## License

Apache 2.0
