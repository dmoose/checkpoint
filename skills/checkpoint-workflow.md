# Skill: Checkpoint Workflow

## When to Checkpoint

- After each **logical unit of work** — a feature, bug fix, refactor, or meaningful step
- NOT after every file edit; group related changes
- NOT after accumulating hours of unrelated work
- Rule of thumb: if you'd write a different commit message, it's a different checkpoint

## The Core Cycle

```
checkpoint check    # Stage changes, generate diff, create input template
# Fill .checkpoint-input with change descriptions
# Review the filled input
checkpoint commit   # Validate, append to changelog, git commit
```

### Step by Step

1. **`checkpoint check`** — Reads staged/unstaged changes, writes `.checkpoint-diff` and `.checkpoint-input`
2. **Fill `.checkpoint-input`** — Describe what changed and why (see `filling-checkpoints.md`)
3. **Review** — Human reviews the filled input before committing
4. **`checkpoint commit`** — Validates input, appends to `.checkpoint-changelog.yaml`, creates git commit

## Filling checkpoint-input

The input file has these key sections:

- **changes[]** — One entry per logical change. Each needs `summary` (<80 chars), `change_type`, `scope`
- **context.problem_statement** — What problem this checkpoint solves (required)
- **key_insights** — What was learned during this work
- **decisions_made** — Choices and alternatives considered (most valuable section)
- **failed_approaches** — What was tried and didn't work
- **key_exchanges** — Captures human-LLM dialogue that led to decisions
- **next_steps** — What to do next, with priority

### The key_exchanges Section

Capture the dialogue that matters:
```yaml
key_exchanges:
  - topic: "Database choice"
    summary: "Discussed SQLite vs Postgres, chose SQLite for single-binary deployment"
    participants: [human, claude]
```

### scope:project — Promoting Learnings

Add `scope: project` to any insight or decision that applies beyond this commit:
```yaml
key_insights:
  - insight: "YAML multi-line strings need explicit style indicators"
    scope: project
```

Project-scoped items get promoted to `.checkpoint-context.yml` for future sessions.

## All Commands

| Command | Purpose |
|---------|---------|
| `checkpoint start [dir]` | Initialize checkpoint in a project |
| `checkpoint check` | Stage changes and create input template |
| `checkpoint commit` | Validate input and create checkpoint commit |
| `checkpoint lint` | Validate checkpoint-input without committing |
| `checkpoint clean` | Remove temporary files (.checkpoint-input, .checkpoint-diff) |
| `checkpoint search [query]` | Search changelog history |
| `checkpoint summary` | Summarize recent checkpoint history |
| `checkpoint plan` | Generate or display development plan |
| `checkpoint session` | Manage session state and handoffs |
| `checkpoint import` | Import existing commits into changelog format |

## Common Mistakes

- **Too vague**: "Fixed stuff" — Be specific about what changed and why
- **Too broad**: 15 unrelated changes in one checkpoint — Split into logical units
- **Skipping context**: Filling changes but leaving decisions_made empty — The reasoning is the most valuable part
- **Skipping review**: Running `checkpoint commit` without human review
- **Empty key_exchanges**: Not capturing the dialogue that led to decisions
- **Missing failed_approaches**: Not documenting what didn't work (saves future sessions from repeating mistakes)
