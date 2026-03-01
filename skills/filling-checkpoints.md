# Skill: Filling Checkpoints

Step-by-step instructions for filling `.checkpoint-input` after running `checkpoint check`.

## Prerequisites

1. Run `checkpoint check` — this creates `.checkpoint-input` (template) and `.checkpoint-diff` (changes)
2. Read both files before filling

## Filling Each Section

### changes[] (required)

One entry per logical change. Read the diff to identify distinct changes.

```yaml
changes:
  - summary: "Add retry logic to HTTP client with exponential backoff"
    change_type: feature
    scope: internal/http
  - summary: "Fix nil pointer when config file is missing"
    change_type: fix
    scope: pkg/config
```

Rules:
- `summary` must be under 80 characters
- `change_type`: one of `feature`, `fix`, `refactor`, `test`, `docs`, `chore`
- `scope`: the package, directory, or component affected
- One entry per logical change, not per file

### context.problem_statement (required)

What problem does this checkpoint solve? Be specific.

```yaml
context:
  problem_statement: "HTTP requests to flaky upstream API fail permanently on first timeout, causing data sync to stall"
```

Bad: "Fix HTTP stuff"
Good: "HTTP requests to flaky upstream API fail permanently on first timeout"

### key_insights

What was learned during this work. Include `scope: project` if it applies broadly.

```yaml
key_insights:
  - insight: "Go's http.Client doesn't retry by default; need explicit retry wrapper"
    scope: commit
  - insight: "Exponential backoff with jitter prevents thundering herd on recovery"
    scope: project
```

### decisions_made (most valuable section)

Every decision with the alternatives you considered. This is what saves future sessions the most time.

```yaml
decisions_made:
  - decision: "Use exponential backoff with jitter instead of fixed intervals"
    reasoning: "Fixed intervals cause thundering herd when service recovers"
    alternatives_considered:
      - "Fixed 1s retry interval — simple but causes load spikes"
      - "Linear backoff — better but still clusters retries"
      - "Circuit breaker pattern — overkill for our traffic volume"
```

### failed_approaches

What was tried and didn't work. Prevents future sessions from repeating mistakes.

```yaml
failed_approaches:
  - approach: "Wrapped http.Transport with retry logic"
    why_failed: "Transport operates at connection level, can't inspect response status codes"
    lesson: "Retry logic belongs at the Client/request level, not Transport level"
```

### key_exchanges

Capture the human-LLM dialogue that influenced decisions.

```yaml
key_exchanges:
  - topic: "Retry strategy"
    summary: "Human asked about circuit breakers; decided they're overkill for current traffic. Will revisit if we add more upstream dependencies."
    participants: [human, claude]
```

### next_steps

What to do next, ordered by priority.

```yaml
next_steps:
  - step: "Add integration test with mock flaky server"
    priority: high
  - step: "Consider circuit breaker if we add more upstream APIs"
    priority: low
```

### Using scope:project

Add `scope: project` to any item that future sessions should know about:

```yaml
key_insights:
  - insight: "All HTTP clients in this project must use the retry wrapper"
    scope: project    # This gets promoted to project context
```

## Validation and Review

```bash
checkpoint lint      # Validate the filled input — fix any errors
```

After filling and validating:
- **Wait for human review** before running `checkpoint commit`
- Do not commit automatically unless the human has explicitly approved

## Checklist

- [ ] Read `.checkpoint-diff` to understand all changes
- [ ] One `changes[]` entry per logical change with specific summary
- [ ] `context.problem_statement` filled with the actual problem
- [ ] `decisions_made` includes alternatives_considered
- [ ] `failed_approaches` filled if anything was tried and abandoned
- [ ] `key_exchanges` captures important dialogue
- [ ] `scope: project` on broadly applicable items
- [ ] `next_steps` filled with priorities
- [ ] `checkpoint lint` passes
- [ ] Human has reviewed before `checkpoint commit`
