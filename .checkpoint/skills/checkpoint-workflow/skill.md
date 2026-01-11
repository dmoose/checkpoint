# Checkpoint Workflow

Structured changelog capture for LLM-assisted development.

## Purpose

Capture what changed and why during development, maintaining context continuity across LLM sessions.

## When to Use

- After completing a logical unit of work
- Before ending a development session
- When you want to document decisions and reasoning

## Workflow

### 1. Start Session
```bash
checkpoint start
```
Shows project status, outstanding next_steps, and validates readiness.

### 2. Do Work
Make your code changes as usual.

### 3. Prepare Checkpoint
```bash
checkpoint check
```
Creates:
- `.checkpoint-input` - Template to fill with change descriptions
- `.checkpoint-diff` - Git diff for context

### 4. Fill Input
Edit `.checkpoint-input` with:

**changes[]** - What changed
- summary: <80 chars, present tense, specific
- change_type: feature|fix|refactor|docs|perf|other
- scope: component affected

**context** - Why and how (most valuable part)
- problem_statement: What problem solved
- key_insights: What learned (mark `scope: project` for project-wide)
- decisions_made: Why this approach, alternatives considered
- established_patterns: New conventions

**next_steps[]** - What's planned
- summary, priority (low|med|high), scope

### 5. Validate
```bash
checkpoint lint
```
Catches common mistakes before commit.

### 6. Commit
```bash
checkpoint commit
```
- Stages all changes
- Appends to changelog
- Creates git commit
- Backfills commit hash

## Tips

- Be specific in summaries: "Add JWT auth middleware" not "Update auth"
- Capture reasoning in context - the "why" is more valuable than "what"
- Mark project-wide patterns with `scope: project`
- Run `checkpoint explain` to understand project context before starting

## Related Commands

- `checkpoint explain` - Get project context
- `checkpoint search` - Query changelog history
- `checkpoint examples` - See good checkpoint examples
