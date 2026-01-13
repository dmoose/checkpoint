# Integrating Checkpoint with LLM Tools

This guide covers how to configure various LLM coding tools to use checkpoint effectively.

## The Core Pattern

Regardless of tool, the integration pattern is:

1. **Session Start:** Provide project context from checkpoint
2. **During Work:** Follow patterns in guidelines, use established tools
3. **Before Finishing:** User runs `checkpoint check` when they decide work is complete
4. **Review:** User reviews and edits .checkpoint-input (what the LLM filled in)
5. **Commit:** User runs `checkpoint commit` after they're satisfied

**Important:** The user decides when to checkpoint. This is intentional - it ensures clean handoff and gives the user opportunity to review/edit what the LLM has added.

## Context Sources

Checkpoint provides several ways to get context:

| Source | Description |
|--------|-------------|
| `checkpoint explain` | Complete context dump for prompts |
| `checkpoint start` | Quick status + next steps |
| `CHECKPOINT.md` | Static reference file with workflow overview |
| `.checkpoint/project.yaml` | Architecture and structure |
| `.checkpoint/guidelines.yaml` | Coding standards |
| `.checkpoint/tools.yaml` | Build/test/lint commands |
| `.checkpoint-session.yaml` | Current planning session (if active) |

---

## Tool-Specific Setup

### Claude Code

Claude Code reads `CLAUDE.md` in your project root. This file already exists in checkpoint projects created with `checkpoint init`.

The `CLAUDE.md` file should contain:
- Build and test commands
- Architecture overview
- Key patterns and conventions
- Critical rules

For additional context at session start:
```bash
# Run this and share output with Claude
checkpoint start
checkpoint explain
```

### Cursor

Create a `.cursorrules` file in your project root:

```
# Project uses checkpoint for development context

## Before Starting
Run `checkpoint start` to see project status and pending work.
Read .checkpoint/guidelines.yaml for coding standards.

## Commands
Build: [from .checkpoint/tools.yaml]
Test: [from .checkpoint/tools.yaml]
Lint: [from .checkpoint/tools.yaml]

## Key Patterns
[Copy relevant sections from .checkpoint/guidelines.yaml]

## After Changes
The user will run `checkpoint check` and `checkpoint commit` when ready.
Do not run these commands automatically.
```

### Aider

Aider can include files in context. Use command line flags:

```bash
aider --read CHECKPOINT.md --read .checkpoint/project.yaml
```

Or create `.aider.conf.yml`:
```yaml
read:
  - CHECKPOINT.md
  - .checkpoint/project.yaml
  - .checkpoint/guidelines.yaml
```

### GitHub Copilot

Copilot doesn't support custom system prompts. Options:

1. Keep relevant files open (Copilot considers open tabs)
2. Reference checkpoint files in comments:
```go
// See .checkpoint/guidelines.yaml for error handling patterns
```

3. Use Copilot Chat with explicit file references:
```
@workspace Based on .checkpoint/guidelines.yaml, how should I handle errors here?
```

### Custom Scripts / API Usage

For direct API usage, include checkpoint context in your system prompt:

```python
import subprocess

def get_checkpoint_context():
    result = subprocess.run(
        ['checkpoint', 'explain'],
        capture_output=True,
        text=True
    )
    return result.stdout

system_prompt = f"""You are a coding assistant.

Project Context:
{get_checkpoint_context()}

Follow the patterns and guidelines above.
When you complete work, the user will run checkpoint commands to record changes.
"""
```

---

## Prompt Templates

### Starting a Session

```
I'm starting work on this project. Here's the current context:

$(checkpoint start)

Today I want to [GOAL]. Please review the context and let me know:
1. Any relevant patterns or decisions I should be aware of
2. Potential challenges based on previous work
3. Suggested approach
```

### Feature Implementation

```
I need to implement [FEATURE].

Project context:
$(checkpoint explain)

Requirements:
- [Requirement 1]
- [Requirement 2]

Please:
1. Propose an approach consistent with project patterns
2. Identify files that will need changes
3. Note any decisions we should document in the checkpoint
```

### Bug Investigation

```
I'm investigating a bug: [DESCRIPTION]

Project context:
$(checkpoint explain project)
$(checkpoint explain tools)

Please help me:
1. Identify likely causes based on codebase patterns
2. Suggest debugging approach
3. Note if this bug pattern might exist elsewhere
```

---

## Session Planning

For complex work, use checkpoint's planning features:

```bash
# Create a planning session
checkpoint plan

# This creates .checkpoint-session.yaml with:
# - goals: What to accomplish
# - approach: How to tackle it
# - next_actions: Tasks with priorities
# - open_questions: What needs clarification

# View current session
checkpoint session

# Prepare for handoff to another session
checkpoint session handoff
```

The session file is transient - it helps organize work but doesn't become part of permanent history. Delete items that are no longer relevant; ignore ones that don't apply.

---

## What the LLM Should Know

When an LLM doesn't know about checkpoint, explain:

```
This project uses 'checkpoint' for development context.

Key commands (run by the user, not you):
- `checkpoint start` - Shows project state and pending work
- `checkpoint plan` - Creates planning session
- `checkpoint check` - Creates input file for recording changes
- `checkpoint commit` - Finalizes the checkpoint

The .checkpoint/ directory contains:
- project.yaml: Architecture and key components
- guidelines.yaml: Coding standards and patterns
- tools.yaml: Build, test, lint commands

When I run `checkpoint check`, it creates .checkpoint-input for you to fill in:
- changes: What changed (summary, type, scope)
- context: Why it changed (problem, decisions, insights)
- next_steps: What remains to do

I will review and edit your input before committing.
```

---

## Common Mistakes

**LLM runs checkpoint commands without being asked:**
Checkpoint commands should be run by the user, not automated. The user decides when to checkpoint.

**Stale context:**
If the LLM seems to be using outdated patterns, re-run `checkpoint explain` and share fresh output.

**Ignoring guidelines:**
Verify the LLM has actually read `.checkpoint/guidelines.yaml`. Include specific rules in your prompt if needed.

**Over-filling checkpoint input:**
Context should be meaningful, not exhaustive. Focus on decisions and insights that help future sessions.
