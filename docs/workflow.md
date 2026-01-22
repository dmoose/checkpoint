# Workflow

The daily workflow for both humans and LLMs using checkpoint and guardrail.

## The Developer Story

### 1. Start a session

```bash
checkpoint start
```

This shows project status, recent changes, pending next steps from previous sessions, and AI authority boundaries. Share this output with your LLM at the start of every session.

### 2. Plan complex work (optional)

```bash
checkpoint plan
```

Creates `.checkpoint-session.yaml` with goals, approach, tasks, and open questions. Useful for multi-session features or when you want structured planning. The session file is transient -- it helps organize work but does not become part of permanent history.

```bash
checkpoint session           # View current session
checkpoint session handoff   # Prepare context for next LLM session
```

### 3. Code

Work on your changes -- with or without an LLM. Use `guardrail explain` if you need to remind yourself (or your LLM) of project patterns and guidelines.

### 4. Generate input template

```bash
checkpoint check
```

This creates `checkpoint-input` (YAML template pre-filled with git status and diff metadata) and `.checkpoint-diff` (the full diff for reference).

### 5. Review the input

This is the gating point. The LLM fills in `checkpoint-input` with change descriptions, context, and decisions. You review and edit before committing. This is your opportunity to:

- Correct inaccurate descriptions
- Add context the LLM missed
- Remove noise or over-documentation
- Adjust priorities on next steps

### 6. Validate and commit

```bash
checkpoint lint       # Optional -- catches placeholder text and schema issues
checkpoint commit     # Validates, appends to changelog, creates git commit
```

The commit appends to `.checkpoint-changelog.yaml`, extracts context to `.checkpoint-context.yaml`, and creates a single git commit. The tool backfills the commit hash into the changelog entry.

---

## The LLM Story

When an LLM is assisting development, this is the expected flow:

### 1. Read project context

```bash
guardrail explain
```

This outputs the full project knowledge base: architecture, tools, guidelines, skills, and AI authority boundaries. The LLM should read this before starting work.

### 2. Read session status

```bash
checkpoint start
```

This shows recent history, pending next steps, and current git status. The LLM should know what was done recently and what remains.

### 3. Follow guidelines during work

The LLM should reference `.checkpoint/guidelines.yaml` for coding standards, `.checkpoint/tools.yaml` for build/test commands, and the `ai_authority` section in `.checkpoint/project.yaml` for what requires human approval.

### 4. Fill `checkpoint-input`

After the human runs `checkpoint check`, the LLM reads `checkpoint-input` and `.checkpoint-diff` and fills in:

- **changes[]** -- Each logical change with summary, details, type, and scope
- **context.problem_statement** -- What problem this checkpoint solves
- **context.key_insights[]** -- What was learned during this work
- **context.decisions_made[]** -- Significant choices and their rationale
- **context.failed_approaches[]** -- What was tried and did not work
- **context.key_exchanges[]** -- Notable human-LLM interactions that shaped the work
- **next_steps[]** -- What remains to do, with priorities

### 5. Wait for human approval

The human reviews, edits, and runs `checkpoint commit`. The LLM should never run checkpoint commands autonomously -- the human decides when to checkpoint.

---

## Knowledge Maintenance

Between checkpoints, use guardrail to capture patterns and rules as you discover them:

```bash
# Capture a coding guideline
guardrail learn "Always validate input at API boundaries" --guideline

# Capture an anti-pattern
guardrail learn "Don't use global state for configuration" --avoid

# Capture a design principle
guardrail learn "Prefer composition over inheritance" --principle

# Capture a tool command
guardrail learn "make test-race" --tool race
```

Periodically review the knowledge base:

```bash
# Check for issues (stale config, missing files, etc.)
guardrail doctor

# Review project config for accumulated recommendations
$EDITOR .checkpoint/project.yaml

# Review and prune guidelines
$EDITOR .checkpoint/guidelines.yaml
```

Project-scoped insights from checkpoints are automatically appended as recommendations to `.checkpoint/project.yaml`. Review these periodically and incorporate the valuable ones into the main document.

---

## Session Handoff

When ending a session that will be continued by another LLM session (or by you later):

```bash
checkpoint session handoff
```

This prepares a context summary including:
- What was accomplished
- What remains from the session plan
- Current state of the codebase
- Open questions or blockers

The next session starts with `checkpoint start` to pick up where the previous session left off.
