# First-Time User Guide

Welcome to checkpoint! This guide will walk you through using checkpoint for the first time.

## What is Checkpoint?

Checkpoint is a tool for capturing structured development history in a way that:
- Maintains continuity across LLM-assisted development sessions
- Creates searchable, structured changelogs
- Captures the "why" behind decisions, not just the "what"
- Works seamlessly with git commits

**Key concept:** One checkpoint = One git commit = One structured entry in your changelog

## Quick Start (5 minutes)

### 1. Initialize Checkpoint

In your project directory (must be a git repository):

```bash
checkpoint init
```

This creates:
- `CHECKPOINT.md` - Workflow documentation
- `.checkpoint-changelog.yaml` - Your append-only changelog
- `.checkpoint-context.yml` - Decision history
- `.checkpoint-project.yml` - Project patterns (you'll curate this)
- `.checkpoint/` - Examples, guides, and config files

**Auto-detection:** Init automatically detects your project's language and build tools from go.mod, package.json, Makefile, etc., and pre-populates config files.

### 2. Verify Setup

Check that everything is configured correctly:

```bash
checkpoint doctor
```

This validates your setup and suggests fixes for any issues:
- Missing config files
- Incomplete tool commands
- Setup problems

Use `checkpoint doctor --verbose` to see what was auto-detected.

### 3. Start a Work Session

Before beginning work:

```bash
checkpoint start
```

This shows:
- Project status checks
- Next steps from your last checkpoint
- Pending recommendations to review

### 4. Do Your Work

Make changes to your code as usual:
- Add features
- Fix bugs
- Refactor code
- Update documentation

Work normally - checkpoint doesn't interfere with your development flow.

### 5. Create a Checkpoint

When you're ready to commit:

```bash
checkpoint check
```

This creates:
- `.checkpoint-input` - Template for you to fill
- `.checkpoint-diff` - Full diff for reference

### 6. Fill the Input File

Open `.checkpoint-input` and describe your changes:

**Required sections:**
- `changes[]` - What changed (be specific!)
- `context` - Why you made these changes

**Tips:**
- Keep summaries under 80 characters
- Use present tense ("Add feature" not "Added feature")
- Explain WHY in context, not just WHAT
- Run `checkpoint examples` to see good examples

**Example:**
```yaml
changes:
  - summary: "Add password reset endpoint"
    details: "Implemented email-based reset with secure tokens"
    change_type: "feature"
    scope: "auth"

context:
  problem_statement: "Users locked out of accounts need password recovery"
  
  key_insights:
    - insight: "Time-limited tokens prevent replay attacks"
      impact: "Tokens expire in 1 hour, balance security with UX"
      scope: "checkpoint"
  
  decisions_made:
    - decision: "Use JWT tokens instead of random strings"
      rationale: "Self-contained, no database lookup needed"
      scope: "checkpoint"
```

### 7. Validate (Optional but Recommended)

Check your work before committing:

```bash
checkpoint lint
```

This catches common mistakes:
- Placeholder text left in
- Summaries too long (>80 chars)
- Missing required fields
- Vague descriptions

### 8. Commit

When satisfied:

```bash
checkpoint commit
```

This:
- Validates your input
- Stages ALL changes (`git add -A`)
- Creates a git commit
- Appends to the changelog
- Captures context for future reference

Done! Your changes are committed with rich metadata.

## Understanding the Files

### Tracked in Git (Permanent History)

**`.checkpoint-changelog.yaml`**
- Append-only changelog
- One YAML document per checkpoint
- Contains: changes, timestamps, commit hashes

**`.checkpoint-context.yml`**
- Append-only context history
- Captures reasoning and decisions
- Maintains development continuity

**`.checkpoint-project.yml`**
- Project-wide patterns and conventions
- Human-curated (you maintain this)
- Starts with recommendations, you review and merge

**`.checkpoint/` directory**
- Examples showing good checkpoints
- Guides like this one
- Tracked so whole team has same references

### Temporary Files (Not Tracked)

**`.checkpoint-input`**
- Template you fill during checkpoint
- Deleted after commit

**`.checkpoint-diff`**
- Full git diff for reference
- Deleted after commit

**`.checkpoint-status.yaml`**
- Last commit metadata
- Used by optional macOS app for discovery

## Common Workflows

### Solo Development

1. `checkpoint start` - See what's next
2. Make changes
3. `checkpoint check` - Create input
4. Fill input file
5. `checkpoint commit` - Commit changes

Repeat daily or after logical units of work.

### LLM-Assisted Development

**Your role:**
1. `checkpoint start` - Share status with LLM
2. Describe what you want to build
3. LLM makes changes
4. `checkpoint check` - LLM prepares checkpoint
5. Review the input file
6. `checkpoint commit` - You commit

**LLM's role:**
- Reads project patterns from `.checkpoint-project.yml`
- Makes code changes
- Fills `.checkpoint-input` with changes and context
- Runs `checkpoint lint` to validate

See `llm-workflow.md` for detailed LLM integration patterns.

### Team Development

**Setup:**
- One person runs `checkpoint init`
- Commit checkpoint files to repository
- Team members pull and start using

**Daily workflow:**
- Everyone uses checkpoint for their commits
- Shared changelog shows team activity
- Project patterns document grows organically

**Curation:**
- Periodically review `.checkpoint-project.yml` recommendations
- Merge valuable patterns into main project document
- Delete recommendations that don't apply

## Best Practices

### ✓ Do This

**Be Specific in Summaries**
- Good: "Add JWT token validation to auth middleware"
- Bad: "Update auth stuff"

**Explain WHY in Context**
- Good: "Chose Redis because we need sub-ms reads at 2000 QPS"
- Bad: "Used Redis for caching"

**Break Large Changes into Logical Units**
- Group related changes (e.g., feature + tests + docs)
- Separate unrelated changes into different checkpoints

**Mark Project Patterns**
```yaml
key_insights:
  - insight: "All API errors use standardized format"
    scope: "project"  # This applies project-wide
```

**Include Failed Approaches**
```yaml
failed_approaches:
  - approach: "Tried caching at database layer"
    why_failed: "Couldn't invalidate cache selectively"
    lessons_learned: "Cache at API layer for control"
```

### ✗ Avoid This

**Vague Descriptions**
- "Fix bug" - Which bug? How?
- "Update code" - What changed?
- "Improve performance" - How much? Where?

**Leaving Placeholders**
```yaml
summary: "[FILL IN: what changed]"  # Forgot to fill!
```

**Too Many Unrelated Changes**
- Don't mix: new feature + refactor + docs update + bug fix
- Each deserves its own checkpoint (or group logically)

**Missing Context**
- Changes without context lose value over time
- Future you won't remember why you made decisions

**Past Tense**
- Use: "Add feature" (present tense)
- Not: "Added feature" (past tense)

## Getting Help

### View Examples

See what good checkpoints look like:

```bash
checkpoint examples              # List categories
checkpoint examples feature      # See feature example
checkpoint examples bugfix       # See bugfix example
checkpoint examples anti-patterns # See common mistakes
```

### Read Guides

```bash
# View available guides
ls .checkpoint/guides/

# Read this guide
cat .checkpoint/guides/first-time-user.md
```

### Check Project Patterns

Your project's established patterns:

```bash
cat .checkpoint-project.yml
```

### View History and Next Steps

See recent checkpoint history:

```bash
checkpoint explain history    # Recent checkpoints, patterns, decisions
checkpoint explain next       # All outstanding next steps by priority
```

### Manage Sessions (LLM Handoff)

When switching between LLM sessions:

```bash
checkpoint session save "Working on auth feature"  # Save current state
checkpoint session                                  # Show session state
checkpoint session handoff                          # Generate handoff document
checkpoint session clear                            # Clear when done
```

### Quick Reference

The main workflow document:

```bash
cat CHECKPOINT.md
```

## Troubleshooting

**General Setup Issues**
- Run `checkpoint doctor` to diagnose problems
- It validates setup and suggests specific fixes

**"Not a git repository"**
- Checkpoint requires git
- Run `git init` first

**"Checkpoint not initialized"**
- Run `checkpoint init` first
- Then run `checkpoint doctor` to verify setup

**"Checkpoint in progress"**
- You have an unfinished checkpoint
- Options:
  - Continue: edit `.checkpoint-input` and `checkpoint commit`
  - Abort: `checkpoint clean` and start over

**"Validation failed"**
- Check error messages
- Common issues:
  - Placeholder text not filled in
  - Summary too long (>80 chars)
  - Invalid change_type or priority
- Run `checkpoint lint` for specific errors

**Input file too long to edit**
- Fixed in recent versions!
- Input file no longer embeds full project context
- If still having issues, check for very large diffs

## Next Steps

Once comfortable with basics:

1. **Explore Examples**: `checkpoint examples` - See various checkpoint types
2. **Learn Context Patterns**: `checkpoint examples context` - Effective context capture
3. **LLM Integration**: Read `.checkpoint/guides/llm-workflow.md`
4. **Project Curation**: Review `.checkpoint-project.yml` recommendations periodically
5. **Customize**: Add your own examples and patterns to `.checkpoint/`

## Philosophy

Checkpoint is designed around a few core principles:

**Append-Only History**
- Never modify past checkpoints
- Corruption isolated to newest entry
- Complete audit trail maintained

**Context is King**
- The "why" is more valuable than the "what"
- Future developers (including you) need context
- LLMs benefit enormously from captured decisions

**Human-LLM Collaboration**
- LLMs fill checkpoints based on changes
- Humans review and curate
- Tool facilitates, doesn't dictate

**Minimal Friction**
- Workflow fits into existing git habits
- One command to start, one to commit
- Templates and examples reduce cognitive load

**Progressive Enhancement**
- Works great with minimal effort
- Rich context capture available when valuable
- Project patterns emerge organically

---

Welcome to checkpoint! Start with `checkpoint start` and build from there.