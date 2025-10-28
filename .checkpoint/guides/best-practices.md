# Checkpoint Best Practices

This guide covers proven practices for getting the most value from checkpoint.

## Core Principles

### 1. Context is More Valuable Than Content

The "why" behind decisions is more valuable than the "what" of changes.

**Why:** Code shows what changed. Context explains why it changed, what alternatives were considered, and what constraints influenced the decision. Future developers (including you in 3 months) need this context.

**Practice:**
- Spend 80% of checkpoint time on context, 20% on change descriptions
- Always explain the problem you're solving
- Document alternatives you considered
- Capture constraints that influenced decisions

### 2. Be Specific, Not Generic

Vague descriptions lose value immediately. Specific details remain valuable forever.

**Bad:**
- "Fix bug"
- "Update code"
- "Improve performance"

**Good:**
- "Fix null pointer in user profile when avatar is missing"
- "Extract validation into internal/validator package"
- "Add index on users.email reducing login from 250ms to 15ms"

**Practice:**
- Include metrics when available
- Name specific files/functions/components
- Use concrete examples, not abstractions

### 3. One Checkpoint = One Logical Unit

Group related changes together, but separate unrelated concerns.

**Related (good for one checkpoint):**
- Feature + its tests + updated docs
- Bug fix + regression test
- Refactor + updated callers

**Unrelated (split into separate checkpoints):**
- New feature + unrelated bug fix
- Multiple independent features
- Refactor + new feature

**Practice:**
- Ask: "Does this all serve the same goal?"
- If you use the word "and" twice, probably split it
- Small checkpoints are fine - frequent checkpoints better than large

### 4. Document What Didn't Work

Failed approaches are valuable teaching moments.

**Why:** Prevents repeating mistakes. Shows future developers what was tried and why it failed.

**Example:**
```yaml
failed_approaches:
  - approach: "Tried caching with Redis TTL only"
    why_failed: "Couldn't invalidate cache when underlying data changed; led to stale data shown to users"
    lessons_learned: "Need explicit invalidation mechanism for data that can change; TTL alone insufficient for cache coherency"
    scope: "project"
```

**Practice:**
- Document every significant approach that failed
- Explain WHY it failed, not just THAT it failed
- Mark as scope: project if lesson applies broadly

### 5. Mark Project Patterns Explicitly

Use `scope: project` to identify reusable patterns and principles.

**When to use scope: project:**
- Establishes pattern for entire codebase
- Future features should follow this approach
- Represents architectural decision
- Would be valuable in project documentation

**Example:**
```yaml
established_patterns:
  - pattern: "All API endpoints validate input before processing"
    rationale: "Prevents injection attacks, ensures data consistency, provides clear error messages"
    examples: "User registration, login, profile updates, data imports"
    scope: "project"
```

**Practice:**
- Review each context item: "Does this apply project-wide?"
- Be generous with scope: project - false positives are fine
- Human curator will review recommendations later

## Writing Effective Changes

### Summary Guidelines

**Format:**
- Under 80 characters
- Present tense, imperative mood ("Add" not "Added")
- Start with verb
- Specific, not vague

**Good patterns:**
- "Add [thing]" - for new features
- "Fix [problem]" - for bugs
- "Refactor [component] to [improvement]" - for refactoring
- "Update [thing] with [detail]" - for modifications
- "Remove [thing]" - for deletions

**Examples:**
- "Add JWT authentication middleware to API"
- "Fix memory leak in database connection pool"
- "Refactor validation into separate package"
- "Update README with Docker setup instructions"
- "Remove deprecated user sync endpoint"

### Details Guidelines

**When to include details:**
- Summary is accurate but incomplete
- Technical explanation adds value
- Implementation approach worth noting

**When to omit details:**
- Summary is self-explanatory
- Details would just restate summary
- Change is trivial

**Good details:**
- "Implemented token validation, refresh logic, and error handling for protected routes"
- "Connections weren't being released in error paths, causing pool exhaustion after 2 hours"
- "Moved input validation from HTTP handlers into internal/validator package with reusable functions"

### Change Type Selection

**feature** - New functionality
- Adding endpoints
- Implementing new features
- Creating new components

**fix** - Bug corrections
- Fixing crashes
- Correcting logic errors
- Resolving data issues

**refactor** - Code restructuring
- Extracting packages
- Renaming for clarity
- Improving code organization

**docs** - Documentation only
- README updates
- API documentation
- Code comments
- Guides and examples

**perf** - Performance improvements
- Adding indexes
- Optimizing algorithms
- Reducing latency

**other** - Everything else
- Dependency updates
- Configuration changes
- Build system modifications

## Writing Effective Context

### Problem Statement

**Purpose:** Explain what problem you're solving and why it matters.

**Bad:**
- "Need to improve the app"
- "Users want this feature"
- "Fix issues"

**Good:**
- "Login API timing out under load (>30s response) due to N+1 query pattern fetching 500+ individual records"
- "Password reset tokens never expire, allowing indefinite reuse and potential account hijacking"
- "Validation logic duplicated across 15 handlers, causing inconsistent error messages and difficult maintenance"

**Template:**
```
[Problem description] causing [impact/symptoms]. [Context on how it was discovered or why it matters].
```

### Key Insights

**Purpose:** Capture what you learned during implementation.

**Structure:**
```yaml
key_insights:
  - insight: "[What you learned]"
    impact: "[Why it matters / how it affects future work]"
    scope: "[checkpoint|project]"
```

**Examples:**
```yaml
- insight: "N+1 queries are common when using ORM eager loading incorrectly"
  impact: "Established code review checklist item; affects all ORM usage project-wide"
  scope: "project"

- insight: "Redis caching reduced dashboard load from 8s to 400ms"
  impact: "Meets 2s SLA with 5x margin; reduces database load by 90%"
  scope: "checkpoint"
```

**Guidelines:**
- Be specific and technical
- Include measurements when possible
- Explain broader implications
- Use scope: project for lessons that apply widely

### Decisions Made

**Purpose:** Document significant choices and their rationale.

**Structure:**
```yaml
decisions_made:
  - decision: "[What you decided]"
    rationale: "[Why this approach]"
    alternatives_considered:
      - "[Alternative 1 (why rejected)]"
      - "[Alternative 2 (why rejected)]"
    constraints_that_influenced: "[What limited your options]"
    scope: "[checkpoint|project]"
```

**Example:**
```yaml
- decision: "Use Redis for session storage with 7-day TTL"
  rationale: "Sub-millisecond reads handle our 5000 QPS; built-in expiration simplifies lifecycle management"
  alternatives_considered:
    - "PostgreSQL sessions (rejected - too slow at our QPS, adds database load)"
    - "In-memory only (rejected - sessions lost on app restart)"
    - "JWT only (rejected - can't invalidate compromised sessions)"
  constraints_that_influenced: "Must support session invalidation, handle 5000 QPS, survive app restarts, minimal operational complexity"
  scope: "project"
```

**Guidelines:**
- Explain why you chose this approach
- List alternatives you seriously considered
- Explain why alternatives were rejected
- Note constraints that limited choices
- Use scope: project for architectural decisions

### Established Patterns

**Purpose:** Document conventions for future code to follow.

**Structure:**
```yaml
established_patterns:
  - pattern: "[What pattern/convention]"
    rationale: "[Why this works for your codebase]"
    examples: "[Where to apply it]"
    scope: "[checkpoint|project]"
```

**Example:**
```yaml
- pattern: "All database queries use connection pooling with max 50 connections"
  rationale: "Prevents connection exhaustion while maintaining performance; 50 conns sufficient for our load (2000 QPS)"
  examples: "API handlers, background jobs, cron tasks, migrations"
  scope: "project"
```

**Guidelines:**
- Make patterns actionable
- Explain why they work for your project
- Give concrete examples of where to apply
- Almost always scope: project

### Conversation Context

**Purpose:** Capture key discussions that influenced implementation.

**Structure:**
```yaml
conversation_context:
  - exchange: "[Discussion point or question raised]"
    outcome: "[How it influenced the implementation]"
```

**Examples:**
```yaml
- exchange: "Discussed whether to use library X vs implement ourselves"
  outcome: "Library chosen - saves 2 weeks development time, well-maintained, team has experience with it"

- exchange: "Debated 5-minute vs 1-minute cache TTL"
  outcome: "5-minute chosen after reviewing analytics - data updates every 2-3 minutes average, so 1-minute provides minimal freshness benefit while doubling cache miss rate"
```

**Guidelines:**
- Capture significant discussions
- Explain the resolution and why
- Include data/reasoning that led to decision
- Keep concise but informative

## Common Scenarios

### Adding a Feature

**Focus on:**
- What problem does this solve?
- Why this approach over alternatives?
- What patterns does it establish?
- What's next in this area?

**Example structure:**
```yaml
changes:
  - summary: "Add password reset with email tokens"
    change_type: "feature"
    scope: "auth"
  - summary: "Add password reset tests"
    change_type: "feature"
    scope: "tests"

context:
  problem_statement: "Users locked out need password recovery"
  
  decisions_made:
    - decision: "Use JWT tokens with 1-hour expiration"
      rationale: "Self-contained, no database lookup"
      alternatives_considered:
        - "Random tokens (rejected - requires database)"
      scope: "project"
  
  established_patterns:
    - pattern: "Security operations use time-limited tokens"
      scope: "project"

next_steps:
  - summary: "Add rate limiting to reset endpoint"
    priority: "high"
```

### Fixing a Bug

**Focus on:**
- What was broken and how was it discovered?
- What was the root cause?
- How did you fix it?
- How do we prevent similar bugs?

**Example structure:**
```yaml
changes:
  - summary: "Fix null pointer in user profile rendering"
    change_type: "fix"
    scope: "frontend"

context:
  problem_statement: "User profiles crash when avatar URL is null, affecting 5% of users"
  
  key_insights:
    - insight: "Avatar URLs can be null for new users"
      impact: "Need null checks before rendering user data"
      scope: "project"
  
  failed_approaches:
    - approach: "Tried providing default avatar in database"
      why_failed: "Existing users still had null values"
      lessons_learned: "Fix must handle existing data"
      scope: "checkpoint"

next_steps:
  - summary: "Audit other null pointer risks in rendering"
    priority: "high"
```

### Refactoring

**Focus on:**
- Why refactor now?
- What problem does it solve?
- What pattern does it establish?
- How do you ensure behavior unchanged?

**Example structure:**
```yaml
changes:
  - summary: "Extract validation into internal/validator"
    change_type: "refactor"
    scope: "internal/validator"
  - summary: "Add comprehensive validation tests"
    change_type: "feature"
    scope: "tests"

context:
  problem_statement: "Validation duplicated across 15 handlers, inconsistent rules"
  
  decisions_made:
    - decision: "Create validator package with functional API"
      rationale: "Simple, testable, composable"
      alternatives_considered:
        - "Third-party library (rejected - adds dependency)"
      scope: "checkpoint"
  
  established_patterns:
    - pattern: "Extract cross-cutting concerns when duplicated 3+ times"
      rationale: "Balance avoiding premature abstraction with reducing duplication"
      scope: "project"
```

## Workflow Tips

### Daily Practice

**Morning:**
```bash
checkpoint start  # Review status and next steps
```

**During work:**
- Work normally
- Mental note of decisions as you make them
- Group related changes

**After each logical unit:**
```bash
checkpoint check     # Create checkpoint
# Fill input file
checkpoint lint      # Validate
checkpoint commit    # Commit
```

**End of day:**
- Review `.checkpoint-project.yml` recommendations
- Merge valuable patterns into main document

### Team Coordination

**Project setup:**
1. One person runs `checkpoint init`
2. Commit checkpoint files
3. Team pulls and uses

**Daily:**
- Everyone checkpoints their work
- Shared patterns emerge organically
- Review others' checkpoints for learning

**Weekly:**
- Team reviews project recommendations
- Curator merges/prunes recommendations
- Align on patterns and conventions

### LLM Collaboration

**Before asking LLM:**
```bash
checkpoint start  # Share status with LLM
```

**After LLM makes changes:**
```bash
checkpoint check  # LLM fills checkpoint
```

**Your review checklist:**
- [ ] Summaries are specific
- [ ] Context explains why
- [ ] Decisions include alternatives
- [ ] Project patterns marked scope: project
- [ ] Next steps are concrete

## Maintenance

### Curate Project File

**Weekly:** Review `.checkpoint-project.yml` recommendations

**For each recommendation:**
- Should it be in main project document?
- Which section does it belong in?
- Does it conflict with existing patterns?
- Is it valuable enough to keep?

**Actions:**
- Merge valuable patterns into main document
- Refine unclear recommendations
- Delete recommendations that don't apply
- Keep recommendation queue under 20 items

### Audit Checkpoints

**Monthly:** Review recent checkpoints for quality

**Check for:**
- Are summaries specific?
- Is context valuable?
- Are patterns being captured?
- Is lint catching issues?

**Improve:**
- Update examples with team's best work
- Refine project patterns
- Share good checkpoints as examples

### Evolve Patterns

**As project grows:**
- Patterns become clearer
- Conventions solidify
- Anti-patterns identified

**Update:**
- `.checkpoint-project.yml` with new insights
- `.checkpoint/examples/` with real examples
- `.checkpoint/guides/` with learned practices

## Anti-Patterns to Avoid

### ✗ Batch Checkpointing

**Bad:** Working for days, then creating one huge checkpoint

**Why:** Lose detail, forget reasoning, context too vague

**Instead:** Checkpoint after each logical unit (feature, bug fix, refactor)

### ✗ Empty Context

**Bad:** Filling only changes[], leaving context empty

**Why:** Loses the most valuable information

**Instead:** Spend more time on context than on changes list

### ✗ Placeholder Text

**Bad:** Leaving "[FILL IN: ...]" in committed checkpoints

**Why:** Checkpoint is worthless without real information

**Instead:** Always run `checkpoint lint` before committing

### ✗ Copy-Paste Context

**Bad:** Copying context from examples without adapting

**Why:** Generic context provides no value

**Instead:** Write context specific to your changes

### ✗ Ignoring Failed Approaches

**Bad:** Only documenting what worked

**Why:** Lose valuable lessons about what doesn't work

**Instead:** Always document significant failed approaches

### ✗ Missing scope: project

**Bad:** Marking project-wide patterns as scope: checkpoint

**Why:** Patterns aren't elevated to project document

**Instead:** Generously use scope: project, curator will filter

## Measuring Success

### Good Checkpoint Indicators

- [ ] Can understand what changed without reading code
- [ ] Context explains why decisions were made
- [ ] Alternatives considered are documented
- [ ] Future developers can learn from it
- [ ] Patterns are marked for reuse
- [ ] Next steps are concrete

### Project Health Indicators

- [ ] Project file has 10+ established patterns
- [ ] Recommendations reviewed regularly (<20 pending)
- [ ] Team checkpoints consistently (most commits)
- [ ] Context quality improving over time
- [ ] Patterns emerge organically from checkpoints

### Personal Development

- [ ] Checkpoint habit established (daily practice)
- [ ] Context quality improving
- [ ] Capturing more insights
- [ ] Better at explaining decisions
- [ ] Learning from past checkpoints

---

These practices evolve with your project. Start simple, add detail as value becomes clear.