# Checkpoint User Guide

Checkpoint captures development context so LLM coding agents can understand your project's history, patterns, and decisions. This guide covers common workflows and how to get the most value from checkpoint.

## Quick Reference

| Command | When to Use |
|---------|-------------|
| `checkpoint init` | First time setup for a project |
| `checkpoint start` | Beginning of any work session |
| `checkpoint plan` | Before complex changes - creates .checkpoint-session.yaml |
| `checkpoint session` | View/manage current planning session |
| `checkpoint check` | When YOU decide to record changes |
| `checkpoint commit` | After reviewing and editing .checkpoint-input |
| `checkpoint explain` | Get context for LLM prompts |
| `checkpoint doctor` | Diagnose configuration issues |
| `checkpoint learn` | Capture knowledge mid-session |

## Understanding Checkpoint's Purpose

Checkpoint solves the "context loss" problem with LLM coding agents. Every time you start a new session, the LLM has no memory of:
- Why the code is structured this way
- What approaches were tried and failed
- What patterns the project follows
- What decisions were made and why

Checkpoint creates a persistent record that any LLM can read, regardless of which tool or model you use tomorrow.

---

## Workflows by Scenario

### 1. Project Setup (One Time)

**When:** Starting to use checkpoint on a new or existing project.

```bash
# Initialize checkpoint in your project
checkpoint init

# Review what was detected and what needs attention
checkpoint doctor -v

# Edit the generated files to add project-specific details
# .checkpoint/project.yaml  - project purpose, architecture
# .checkpoint/tools.yaml    - build, test, lint commands
# .checkpoint/guidelines.yaml - coding standards, patterns
```

**What to configure in project.yaml:**
- `purpose`: One sentence explaining what this project does
- `architecture.overview`: High-level structure (monolith, microservices, CLI, library)
- `key_components`: The main modules and what they do
- `dependencies.runtime`: Critical external dependencies

**What to configure in tools.yaml:**
- `build`: How to compile/build the project
- `test`: How to run tests (unit, integration, e2e)
- `lint`: Code quality checks
- `run`: How to run the project locally

**What to configure in guidelines.yaml:**
- `naming`: Conventions for files, functions, variables
- `structure`: Where different types of code belong
- `rules`: Must-follow guidelines
- `avoid`: Anti-patterns specific to this project

### 2. Starting a Work Session

**When:** Beginning any development work, especially with an LLM agent.

```bash
# See project state, recent changes, and pending work
checkpoint start
```

This shows:
- Project summary and purpose
- Recent checkpoint history
- Pending next steps from previous sessions
- Current git status

**For LLM agents:** Include the output of `checkpoint start` or `checkpoint explain` at the beginning of your session. This gives the LLM immediate context about the project.

### 3. Quick Bug Fix

**When:** Small, focused fix with clear scope.

```bash
# 1. Check current state
checkpoint start

# 2. Make the fix
# ... edit code ...

# 3. Create checkpoint
checkpoint check

# 4. Fill in the generated .checkpoint-input file:
#    - summary: "Fix null pointer in user validation"
#    - change_type: fix
#    - context.problem_statement: "Users with empty email caused crash"
#    - context.key_insights: What you learned about the bug

# 5. Validate and commit
checkpoint lint
checkpoint commit
```

**Tips for bug fix context:**
- Document the root cause, not just the symptom
- Note if this bug pattern could exist elsewhere
- Record any failed debugging approaches

### 4. Feature Development

**When:** Adding new functionality, especially multi-session work.

```bash
# 1. Start session and review pending work
checkpoint start

# 2. Create a session plan (recommended for features)
checkpoint plan
# This creates .checkpoint-session.yaml - edit it with:
# - goals: What you want to accomplish
# - approach: How you'll tackle it
# - next_actions: Tasks with priorities
# - open_questions: What needs clarification

# 3. Implement the feature
# ... development work ...

# 4. When YOU decide you've reached a logical stopping point:
checkpoint check

# 5. Review .checkpoint-input - this is your chance to:
#    - Verify the LLM's descriptions are accurate
#    - Edit/correct/extend the context
#    - Add decisions_made, alternatives_considered, etc.

# 6. Commit when you're satisfied with the input
checkpoint commit
```

**When to create checkpoints during feature work:**
- After completing a logical unit (not every file change)
- Before switching to a different part of the feature
- At end of day/session
- After making significant architectural decisions

**Tips for feature context:**
- Record WHY you chose this approach over alternatives
- Document any patterns you're establishing
- Note dependencies or prerequisites for future work
- Be specific about what's done vs what remains

### 5. Research and Exploration

**When:** Investigating approaches, evaluating options, or learning the codebase.

```bash
# 1. Start with context
checkpoint start

# 2. Create exploration plan
checkpoint plan
# Document:
# - What question you're trying to answer
# - What approaches you'll evaluate
# - Success criteria

# 3. During exploration, capture learnings
checkpoint learn "React Query handles cache invalidation better than SWR for our use case" --principle
checkpoint learn "Don't use ORM for batch operations - use raw SQL" --avoid

# 4. When exploration is complete, checkpoint the findings
checkpoint check
# Focus context on:
# - What was learned
# - What was ruled out and why
# - Recommended approach
```

**Tips for exploration context:**
- Failed approaches are valuable - document them
- Record the evaluation criteria you used
- Note any surprises or counter-intuitive findings
- Distinguish between "won't work" and "didn't try"

### 6. Refactoring

**When:** Restructuring code without changing behavior.

```bash
checkpoint start

# For large refactors, plan first
checkpoint plan

# Make changes in logical chunks, checkpoint each
# Chunk 1: Extract interface
checkpoint check && checkpoint commit

# Chunk 2: Move implementations
checkpoint check && checkpoint commit

# Chunk 3: Update consumers
checkpoint check && checkpoint commit
```

**Tips for refactoring context:**
- Document the goal of the refactor (why now?)
- Record the before/after structure
- Note any behavior that accidentally changed
- Capture patterns that should be followed going forward

### 7. Onboarding (New Developer or New LLM Session)

**When:** Getting up to speed on an unfamiliar project.

```bash
# Get comprehensive project context
checkpoint explain

# Or get specific sections
checkpoint explain project    # Just project overview
checkpoint explain tools      # Build/test commands
checkpoint explain guidelines # Coding standards

# Review recent history
checkpoint summary --recent 10

# Search for specific topics
checkpoint search "authentication"
checkpoint search "database migration"
```

**For LLM agents:** When starting work on an unfamiliar project, request:
1. Output of `checkpoint explain`
2. Recent changelog entries relevant to your task
3. Any session plans from previous work

---

## Writing Effective Context

The context you capture determines how useful checkpoint is. Here's how to write context that helps future LLM sessions.

### Problem Statement

Bad: "Fix bug"
Good: "Users with special characters in email addresses couldn't log in because the validation regex didn't account for plus signs"

The problem statement should be specific enough that someone unfamiliar with the code understands what was wrong and why it mattered.

### Key Insights

Capture learnings that would save time if known earlier:

```yaml
key_insights:
  - insight: "The auth middleware caches user permissions for 5 minutes"
    impact: "Permission changes aren't immediate - tests need to account for this"
    scope: project  # This affects all future auth work

  - insight: "Batch inserts over 1000 rows trigger a different code path"
    impact: "Need to test both small and large batches"
    scope: checkpoint  # Specific to this change
```

### Decisions Made

Document the reasoning, not just the choice:

```yaml
decisions_made:
  - decision: "Use UUIDs instead of auto-increment IDs for user records"
    rationale: "Prevents enumeration attacks and simplifies multi-region sync"
    alternatives_considered:
      - "Auto-increment with obfuscation - rejected due to complexity"
      - "Snowflake IDs - rejected due to additional infrastructure"
    scope: project
```

### Failed Approaches

These prevent repeating mistakes:

```yaml
failed_approaches:
  - approach: "Tried using database triggers for audit logging"
    why_failed: "Triggers don't have access to application context (user ID, request ID)"
    lessons_learned: "Audit logging must happen at application layer"
    scope: project
```

### Established Patterns

Document patterns that should be followed:

```yaml
established_patterns:
  - pattern: "All API endpoints return {data, error, metadata} structure"
    rationale: "Consistent response format simplifies client error handling"
    examples: "See /api/users/handlers.go for reference implementation"
    scope: project
```

---

## Integrating with LLM Tools

### Claude Code / Cursor / Aider

Add to your system prompt or project instructions:

```
Before starting work, run `checkpoint start` to see project context.
After completing changes, use `checkpoint check` to create a checkpoint.
Reference .checkpoint/project.yaml for architecture decisions.
Reference .checkpoint/guidelines.yaml for coding standards.
```

### Custom Prompts

Use `checkpoint explain` output in your prompts:

```bash
# Generate context for a prompt
checkpoint explain > /tmp/context.md

# Or pipe directly
echo "Given this project context:
$(checkpoint explain)

Please implement..."
```

### CHECKPOINT.md

The `CHECKPOINT.md` file in your project root is designed for LLMs to read directly. It contains:
- Project overview
- Recent changes
- Pending next steps
- Key patterns and decisions

Include it in your LLM's file context or reference it in prompts.

---

## Best Practices

### Checkpoint Frequency

- **Too often:** Every small change creates noise, makes history hard to scan
- **Too rare:** Large checkpoints lose detail, context becomes vague
- **Right frequency:** Logical units of work - a complete bug fix, a feature milestone, end of focused session

### Context Quality Over Quantity

A few well-written insights are more valuable than many vague ones. Focus on:
- Information that would change how you approach similar problems
- Decisions that aren't obvious from the code
- Gotchas that wasted time

### Scope: Checkpoint vs Project

Use `scope: checkpoint` for:
- Insights specific to this change
- Decisions that only affect this feature
- Local patterns within a module

Use `scope: project` for:
- Patterns that should be followed everywhere
- Architectural decisions
- Anti-patterns to avoid project-wide

Project-scoped items are extracted as recommendations for human review.

### Maintaining Project Files

The files in `.checkpoint/` should evolve:
- After several checkpoints, review accumulated recommendations
- Incorporate valuable patterns into main project.yaml/guidelines.yaml
- Remove outdated or superseded guidance
- Keep files focused - delete noise

---

## Troubleshooting

### "LLM keeps making the same mistakes"

1. Check if the mistake is documented in `guidelines.yaml` under `avoid`
2. If not, add it: `checkpoint learn "Don't do X because Y" --avoid`
3. Ensure LLM is reading project context at session start

### "Checkpoints feel like busywork"

1. Reduce frequency - checkpoint at logical boundaries, not every change
2. Focus on context that helps YOU when you return to this code
3. Use `checkpoint learn` for quick captures during work

### "Context is getting stale"

1. Run `checkpoint doctor` to identify issues
2. Review and prune `.checkpoint/` files
3. Update project.yaml when architecture changes
4. Remove completed items from next_steps

### "LLM doesn't seem to use the context"

1. Verify context is included in prompt/session start
2. Check that `checkpoint explain` output is readable
3. Try more explicit instructions: "Read CHECKPOINT.md before starting"
4. Some LLMs need context in specific formats - adjust as needed
