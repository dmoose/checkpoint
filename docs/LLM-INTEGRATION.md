# Integrating with LLM Tools

This guide covers how to configure various LLM coding tools to use checkpoint and guardrail effectively.

## The Core Pattern

Regardless of tool, the integration pattern is:

1. **Session Start:** Provide project context from guardrail
2. **During Work:** Follow patterns in guidelines, use established tools
3. **Before Finishing:** User runs `checkpoint check` when they decide work is complete
4. **Review:** User reviews and edits checkpoint-input (what the LLM filled in)
5. **Commit:** User runs `checkpoint commit` after they're satisfied

**Important:** The user decides when to checkpoint. This is intentional -- it ensures clean handoff and gives the user opportunity to review/edit what the LLM has added.

## Context Sources

| Source | Description |
|--------|-------------|
| `guardrail explain` | Complete context dump for prompts |
| `checkpoint start` | Quick status + next steps |
| `guardrail guide` | Built-in workflow guides |
| `.checkpoint/project.yaml` | Architecture and structure |
| `.checkpoint/guidelines.yaml` | Coding standards |
| `.checkpoint/tools.yaml` | Build/test/lint commands |
| `.checkpoint/skills.yaml` | Available LLM skills |
| `.checkpoint-session.yaml` | Current planning session (if active) |

---

## MCP Server

`checkpoint-mcp` exposes project context over the Model Context Protocol, letting MCP-compatible editors query checkpoint data directly instead of piping CLI output.

```bash
checkpoint-mcp -project /path/to/your/project
```

The server provides 5 tools:

| Tool | Description |
|------|-------------|
| `explain` | Project context -- architecture, patterns, guidelines, tools, skills |
| `search` | Search changelog history for decisions, patterns, failed approaches |
| `guide` | Built-in guides for workflow, best practices, LLM integration |
| `status` | Recent activity, next steps, pending work |
| `context_template` | YAML template for filling checkpoint input |

Configure it in your editor's MCP settings (e.g. `.mcp.json`):

```json
{
  "mcpServers": {
    "checkpoint": {
      "command": "checkpoint-mcp",
      "args": ["-project", "."]
    }
  }
}
```

---

## Tool-Specific Setup

### Claude Code

Claude Code reads `CLAUDE.md` in your project root. This file should contain build/test commands, architecture overview, key patterns, and critical rules.

For additional context at session start:
```bash
# Run these and share output with Claude
checkpoint start
guardrail explain
```

### Cursor

Create a `.cursorrules` file in your project root:

```
# Project uses checkpoint + guardrail for development context

## Before Starting
Run `checkpoint start` to see project status and pending work.
Run `guardrail explain` for full project context.
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
aider --read .checkpoint/project.yaml --read .checkpoint/guidelines.yaml
```

Or create `.aider.conf.yml`:
```yaml
read:
  - .checkpoint/project.yaml
  - .checkpoint/guidelines.yaml
  - .checkpoint/tools.yaml
```

For full context, pipe guardrail output:
```bash
guardrail explain > /tmp/context.md
aider --read /tmp/context.md
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

For direct API usage, include guardrail context in your system prompt:

```python
import subprocess

def get_project_context():
    result = subprocess.run(
        ['guardrail', 'explain'],
        capture_output=True,
        text=True
    )
    return result.stdout

system_prompt = f"""You are a coding assistant.

Project Context:
{get_project_context()}

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
$(guardrail explain)

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
$(guardrail explain project)
$(guardrail explain tools)

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

The session file is transient -- it helps organize work but doesn't become part of permanent history.

---

## What the LLM Should Know

When an LLM doesn't know about checkpoint and guardrail, explain:

```
This project uses two tools for development context:

'checkpoint' records development history:
- `checkpoint start` - Shows project state and pending work
- `checkpoint plan` - Creates planning session
- `checkpoint check` - Creates input file for recording changes
- `checkpoint commit` - Finalizes the checkpoint

'guardrail' manages project knowledge:
- `guardrail explain` - Shows full project context (architecture, tools, guidelines)
- `guardrail learn` - Captures new guidelines or patterns
- `guardrail doctor` - Checks knowledge base health

The .checkpoint/ directory contains:
- project.yaml: Architecture and key components
- guidelines.yaml: Coding standards and patterns
- tools.yaml: Build, test, lint commands
- skills.yaml: Available LLM skills

When I run `checkpoint check`, it creates checkpoint-input for you to fill in:
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
If the LLM seems to be using outdated patterns, re-run `guardrail explain` and share fresh output.

**Ignoring guidelines:**
Verify the LLM has actually read `.checkpoint/guidelines.yaml`. Include specific rules in your prompt if needed.

**Over-filling checkpoint input:**
Context should be meaningful, not exhaustive. Focus on decisions and insights that help future sessions.
