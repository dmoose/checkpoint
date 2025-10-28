# LLM Workflow Guide

This guide explains how to effectively use checkpoint in LLM-assisted development workflows.

## Overview

Checkpoint is designed for human-LLM collaboration. The tool helps maintain development continuity across sessions by capturing not just what changed, but why decisions were made.

**Key principle:** LLM fills checkpoints, human reviews and commits.

## Quick LLM Workflow

### 1. Human Starts Session

```bash
checkpoint start
```

**Human shares output with LLM:**
- "Here's our project status. Let's work on [next step from list]"

### 2. Human Describes Task

Be specific about what you want:
- "Add password reset functionality with email tokens"
- "Fix the memory leak in database connection pool"
- "Refactor validation logic into separate package"

### 3. LLM Makes Changes

LLM modifies code files as needed.

### 4. LLM Creates Checkpoint

When work is complete:

```bash
checkpoint check
```

### 5. LLM Fills Input File

LLM reads:
- `.checkpoint-input` - Template to fill
- `.checkpoint-diff` - Full diff of changes
- `.checkpoint-project.yml` - Project patterns to follow

LLM fills:
- `changes[]` - What changed (specific summaries)
- `context` - Why changes were made (critical!)
- `next_steps[]` - What should be done next

### 6. LLM Validates

```bash
checkpoint lint
```

Catches common mistakes before human review.

### 7. Human Reviews and Commits

Human reviews `.checkpoint-input`, then:

```bash
checkpoint commit
```

## LLM Responsibilities

### Analyzing Changes

**Do:**
- Read the full diff carefully
- Group related changes logically
- Identify distinct units of work

**Example:**
```yaml
changes:
  - summary: "Add password reset endpoint"
    change_type: "feature"
    scope: "auth"
  
  - summary: "Add password reset tests"
    change_type: "feature"
    scope: "tests"
  
  - summary: "Update API docs for password reset"
    change_type: "docs"
    scope: "api"
```

### Writing Summaries

**Rules:**
- Under 80 characters
- Present tense, imperative mood
- Specific, not vague
- Start with verb ("Add", "Fix", "Refactor", "Update")

**Good examples:**
- "Add JWT token validation to auth middleware"
- "Fix null pointer in user profile rendering"
- "Refactor validation into internal/validator package"
- "Update README with installation instructions"

**Bad examples:**
- "Updated some files" (vague)
- "Added feature" (not specific)
- "Fixed bug" (which bug?)
- "Made improvements" (meaningless)

### Capturing Context

**This is the most valuable part!**

The context section should explain:
- **Problem:** What problem did we solve?
- **Insights:** What did we learn?
- **Decisions:** Why did we choose this approach?
- **Alternatives:** What else did we consider?
- **Patterns:** What should future work follow?

**Example - Good Context:**
```yaml
context:
  problem_statement: "Password reset was broken - tokens weren't expiring, allowing indefinite reuse after reset"
  
  key_insights:
    - insight: "JWT self-contained tokens eliminate database lookup overhead"
      impact: "Reset endpoint handles 1000 QPS without database bottleneck"
      scope: "checkpoint"
    
    - insight: "Time-limited tokens are critical for security operations"
      impact: "All security-sensitive operations should use expiring tokens"
      scope: "project"
  
  decisions_made:
    - decision: "Use JWT with 1-hour expiration for reset tokens"
      rationale: "Balance security (limited exposure) with UX (reasonable time to reset)"
      alternatives_considered:
        - "Random tokens in database (rejected - adds database dependency)"
        - "15-minute expiration (rejected - too short for email delivery delays)"
      constraints_that_influenced: "Must handle email delivery delays up to 30 minutes"
      scope: "project"
```

**Example - Poor Context:**
```yaml
context:
  problem_statement: "Fixed password reset"
  
  key_insights:
    - insight: "JWT is good"
      impact: "Helps performance"
  
  decisions_made:
    - decision: "Used JWT"
      rationale: "Better than alternatives"
```

### Scope: Project vs Checkpoint

**Use `scope: project` when:**
- Establishes pattern for entire codebase
- Should be followed in future work
- Represents architectural decision
- Would be valuable in project patterns document

**Examples:**
```yaml
key_insights:
  - insight: "All security operations use time-limited tokens"
    scope: "project"
  
  - insight: "Validation errors return 400 with field-level details"
    scope: "project"

established_patterns:
  - pattern: "All API endpoints validate input before processing"
    rationale: "Prevents injection and ensures data consistency"
    scope: "project"
```

**Use `scope: checkpoint` (or omit) when:**
- Specific to this implementation
- Technical detail that doesn't generalize
- One-time decision

**Examples:**
```yaml
key_insights:
  - insight: "Reset endpoint reduced from 200ms to 15ms with JWT"
    scope: "checkpoint"
  
decisions_made:
  - decision: "Store reset attempts in Redis with 1-hour TTL"
    scope: "checkpoint"
```

## Reading Project Context

Before filling checkpoint input, LLM should read:

### `.checkpoint-project.yml`

Contains project-wide patterns:
- Dependencies and their rationale
- Established coding patterns
- Testing methodologies
- Error handling approaches
- Performance considerations

**Use this to:**
- Ensure new code follows existing patterns
- Mark similar patterns with `scope: project`
- Understand project constraints

### `.checkpoint-context.yml`

Recent checkpoint contexts (last few entries):
- Recent decisions
- Current work direction
- Failed approaches to avoid

**Use this to:**
- Maintain consistency with recent work
- Avoid repeating failed approaches
- Build on recent insights

### `.checkpoint/examples/`

Reference examples:
```bash
checkpoint examples feature      # See feature example
checkpoint examples context      # See context examples
checkpoint examples anti-patterns # Avoid mistakes
```

## Common LLM Mistakes

### 1. Vague Summaries

**Bad:**
```yaml
- summary: "Update files"
  change_type: "other"
```

**Good:**
```yaml
- summary: "Add input validation to user registration endpoint"
  change_type: "feature"
  scope: "api/handlers"
```

### 2. Missing Context

**Bad:**
```yaml
context:
  problem_statement: "Need to fix bugs"
```

**Good:**
```yaml
context:
  problem_statement: "User registration failing for emails with + character due to improper URL encoding in validation"
```

### 3. No Alternatives Documented

**Bad:**
```yaml
decisions_made:
  - decision: "Used Redis"
    rationale: "It's fast"
```

**Good:**
```yaml
decisions_made:
  - decision: "Use Redis for session storage with 7-day TTL"
    rationale: "Sub-millisecond reads, handles 5000 QPS, built-in expiration"
    alternatives_considered:
      - "PostgreSQL sessions (rejected - too slow at our QPS)"
      - "In-memory (rejected - doesn't survive restarts)"
      - "JWT only (rejected - can't invalidate sessions)"
    constraints_that_influenced: "Must support session invalidation, handle 5000 QPS, survive app restarts"
```

### 4. Forgetting scope: project

**Bad:**
```yaml
established_patterns:
  - pattern: "All database queries use connection pooling"
    scope: "checkpoint"  # Should be project!
```

**Good:**
```yaml
established_patterns:
  - pattern: "All database queries use connection pooling"
    rationale: "Prevents connection exhaustion, improves performance"
    examples: "API handlers, background jobs, cron tasks"
    scope: "project"
```

### 5. Past Tense

**Bad:**
```yaml
changes:
  - summary: "Added authentication middleware"
```

**Good:**
```yaml
changes:
  - summary: "Add authentication middleware"
```

## Advanced Patterns

### Capturing Failed Approaches

**Always document what didn't work:**

```yaml
failed_approaches:
  - approach: "Tried optimizing with database indexes only"
    why_failed: "Query time reduced to 5s but still exceeded 2s SLA; root cause was N+1 pattern (500 queries), not query speed"
    lessons_learned: "Profile before optimizing - measure actual bottleneck not assumed one. Query count matters more than query speed."
    scope: "project"
```

This prevents repeating mistakes in future features.

### Conversation Context

**Capture key discussions:**

```yaml
conversation_context:
  - exchange: "Discussed whether to cache at API vs database layer"
    outcome: "API layer chosen - gives cache invalidation control and keeps database layer simple"
  
  - exchange: "Debated 5-minute vs 1-minute TTL for cache"
    outcome: "5-minute chosen after reviewing analytics - data updates every 2-3 minutes average, so 1min provides minimal benefit while doubling miss rate"
```

### Next Steps

**Be specific and prioritized:**

```yaml
next_steps:
  - summary: "Add rate limiting to password reset endpoint"
    details: "Prevent brute force attacks, limit to 3 attempts per hour per email"
    priority: "high"
    scope: "auth"
  
  - summary: "Add monitoring for reset token expiry rate"
    details: "Track how many tokens expire unused - indicates TTL might be too short"
    priority: "med"
    scope: "observability"
```

## Prompt Templates

### Session Start Prompt

```
I'm working on a checkpoint-managed project. Here's the current status:

[paste output of: checkpoint start]

Let's work on: [describe task]

When we're done, I'll run `checkpoint check` and you'll fill the checkpoint input.
```

### Checkpoint Filling Prompt

```
I've run `checkpoint check`. Please:

1. Read .checkpoint-diff to understand all changes
2. Read .checkpoint-project.yml to understand project patterns
3. Fill .checkpoint-input with:
   - Specific summaries for each change (<80 chars)
   - Context explaining WHY we made these changes
   - Decisions made and alternatives considered
   - Mark project-wide patterns with scope: project
   - List concrete next steps

Focus on capturing reasoning and decisions, not just describing changes.
Then run `checkpoint lint` to validate.
```

### Review Recommendations Prompt

```
Please review the recommendations in .checkpoint-project.yml:

[paste recommendations section]

For each recommendation:
- Should it be added to the main project document?
- Which section does it belong in?
- Does it conflict with existing patterns?

Suggest which to merge, which to refine, and which to delete.
```

## Integration with Development Tools

### With Code Review

**Before submitting PR:**
1. Use checkpoint for each logical commit
2. Share checkpoint context in PR description
3. Reviewers can reference `.checkpoint-changelog.yaml` for history

### With CI/CD

**Checkpoint in automation:**
```bash
# In CI, verify checkpoint consistency
checkpoint verify  # (future feature)

# Check for placeholder text in last commit
checkpoint lint --last-commit  # (future feature)
```

### With Project Documentation

**Generate docs from checkpoints:**
```bash
# Extract architectural decisions
grep -A 10 "decisions_made:" .checkpoint-context.yml

# Find patterns for specific component
grep -B 2 "scope: auth" .checkpoint-changelog.yaml
```

## Best Practices for LLMs

### ✓ Do This

1. **Read project context first**
   - Understand established patterns
   - Follow existing conventions
   - Build on recent decisions

2. **Be specific and concrete**
   - Use metrics where possible
   - Name specific files/functions
   - Explain technical details

3. **Explain alternatives**
   - What else did you consider?
   - Why was it rejected?
   - What constraints influenced the decision?

4. **Mark patterns appropriately**
   - scope: project for reusable patterns
   - scope: checkpoint for one-time details

5. **Run lint before finishing**
   - Catch mistakes early
   - Validate completeness

### ✗ Avoid This

1. **Don't leave placeholders**
   - Fill all required fields
   - Remove example text

2. **Don't be vague**
   - No "update stuff" or "fix things"
   - Be specific about what and why

3. **Don't forget context**
   - Context is more valuable than change list
   - Explain reasoning, not just results

4. **Don't mix unrelated changes**
   - Group logically related changes
   - Separate different concerns

5. **Don't use past tense**
   - "Add feature" not "Added feature"
   - Present tense, imperative mood

## Troubleshooting

**"How detailed should context be?"**
- More detail is better than less
- Explain WHY you made decisions
- Include metrics when available
- Document alternatives considered

**"Should I mention every file changed?"**
- No, group by logical change
- Example: "Add auth middleware" covers multiple files
- One change = one logical unit of work

**"What if I don't know why something was done?"**
- Ask the human
- Explain what you understand
- Note uncertainty in context

**"Should I fill next_steps?"**
- Yes, if there are logical follow-ups
- No, if work is complete
- Be specific about what's next

## Example Session

**Human:** "Add user authentication to the API"

**LLM:** [Implements JWT authentication, tests, docs]

**Human:** `checkpoint check`

**LLM:** 
```
I'll fill the checkpoint input now.

Reading .checkpoint-diff... 
Changes span auth middleware, tests, and API docs.

Reading .checkpoint-project.yml...
Existing pattern: Use HTTP-only cookies for sensitive data
Existing pattern: All endpoints validate input

Filling .checkpoint-input...
- 3 changes: middleware, tests, docs
- Context: Explains JWT choice over sessions
- Decisions: Token expiration times with rationale
- Marked cookie storage as scope: project (follows existing pattern)
- Next steps: Add rate limiting

Running checkpoint lint...
✅ No issues found
```

**Human:** [Reviews, approves] `checkpoint commit`

---

This workflow ensures rich context capture while maintaining development velocity.