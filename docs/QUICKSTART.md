# Checkpoint Quickstart

Get productive with checkpoint in 5 minutes.

## Install

```bash
# From source
go install github.com/dmoose/checkpoint@latest

# Or clone and build
git clone https://github.com/dmoose/checkpoint
cd checkpoint
make install-user  # installs to ~/bin
```

## Setup (Once Per Project)

```bash
cd your-project
checkpoint init
checkpoint doctor -v  # verify setup
```

Edit `.checkpoint/project.yaml` with your project's purpose and structure.

## Daily Workflow

```bash
# 1. Start session - see project state and pending work
checkpoint start

# 2. Plan complex work (optional but recommended)
checkpoint plan         # creates .checkpoint-session.yaml

# 3. Do work...

# 4. When ready to record changes (you decide when)
checkpoint check        # generates .checkpoint-input and .checkpoint-diff

# 5. Review and edit .checkpoint-input
#    - Fill in change summaries and context
#    - This is YOUR opportunity to review what the LLM added
#    - Edit/change/extend as needed

# 6. Validate and commit
checkpoint lint         # optional - catches obvious errors
checkpoint commit       # finalizes and creates git commit
```

## Essential Commands

| Command | Purpose |
|---------|---------|
| `checkpoint start` | See project state, next steps |
| `checkpoint plan` | Create planning session (.checkpoint-session.yaml) |
| `checkpoint check` | Create checkpoint input file |
| `checkpoint commit` | Finalize checkpoint |
| `checkpoint explain` | Get context for LLM prompts |
| `checkpoint session` | View/manage current session |
| `checkpoint learn "..." --guideline` | Capture a coding rule |
| `checkpoint search "term"` | Search changelog history |

## For LLM Sessions

Include at session start:
```bash
checkpoint explain
# or
cat CHECKPOINT.md
```

## Quick Captures

```bash
# Coding rule to follow
checkpoint learn "Always validate input at API boundaries" --guideline

# Anti-pattern to avoid
checkpoint learn "Don't use global state for config" --avoid

# Tool command
checkpoint learn "make test-race" --tool race
```

## File Locations

```
.checkpoint/
├── project.yaml      # Architecture, purpose
├── tools.yaml        # Build/test/lint commands
├── guidelines.yaml   # Coding standards
└── skills/           # Custom skill definitions

CHECKPOINT.md         # Summary for LLMs (auto-generated)
.checkpoint-changelog.yaml  # Full history
```

## Getting Help

```bash
checkpoint --help
checkpoint <command> --help
checkpoint doctor  # diagnose issues
```
