# Checkpoint Workflow - Quick Reference

This repository uses an append-only changelog to capture LLM-assisted development work.

**First time here?** Run `checkpoint guide first-time-user` for a complete walkthrough.

**Using with LLM?** Run `checkpoint guide llm-workflow` for integration patterns.

**Need examples?** Run `checkpoint examples` to see well-structured checkpoints.

Key files:
- .checkpoint-input: Edit this during a checkpoint to describe changes and context
- .checkpoint-diff: Git diff context for the LLM
- .checkpoint-changelog.yaml: Append-only YAML changelog; one document per checkpoint with changes[]
- .checkpoint-context.yml: Append-only context log; captures reasoning and decisions
- .checkpoint-project.yml: Project-wide patterns and conventions (human-curated)
- .checkpoint-status.yaml: Last commit metadata for discovery

Concepts:
- One checkpoint equals one Git commit. The tool stages ALL changes when committing.
- The changelog document is appended before commit; after commit, the tool backfills commit_hash into the last document without another commit.
- Each document contains an array of changes; use concise summaries and optional details.
- Context captures the "why" behind decisions to maintain continuity across LLM sessions.
- Project file aggregates project-wide patterns; human curates from checkpoint recommendations.

Basic workflow:
1. Start: checkpoint start
   - Check project status and see next steps
2. Make changes: (code as usual)
3. Prepare: checkpoint check [path]
   - Generates .checkpoint-input and .checkpoint-diff
4. Fill input: Open .checkpoint-input and describe:
   - changes[] - what changed (be specific!)
   - context - why and how (problem, decisions, insights)
   - next_steps[] - planned work
5. Validate: checkpoint lint (optional but recommended)
6. Commit: checkpoint commit [path]
   - Stages all changes, creates a commit, backfills commit_hash
   - Appends to changelog, context, and project files
7. Periodically: review .checkpoint-project.yml recommendations and curate

Schema (YAML):
---
schema_version: "1"
timestamp: "<auto>"
commit_hash: "<filled after commit>"
changes:
  - summary: "<what changed>"
    details: "<optional longer description>"
    change_type: "feature|fix|refactor|docs|perf|other"
    scope: "<component>"
context:
  problem_statement: "<what problem are we solving>"
  key_insights: [...]
  decisions_made: [...]
  established_patterns: [...]
  conversation_context: [...]
next_steps:
  - summary: "<planned action>"
    details: "<optional>"
    priority: "low|med|high"
    scope: "<component>"

## Learning Resources

**Comprehensive Guides:**
- `checkpoint guide first-time-user` - Complete walkthrough for newcomers
- `checkpoint guide llm-workflow` - LLM integration patterns and workflow
- `checkpoint guide best-practices` - Best practices for effective checkpoints

**Examples:**
- `checkpoint examples` - List all available examples
- `checkpoint examples feature` - See good feature checkpoint
- `checkpoint examples bugfix` - See good bug fix checkpoint
- `checkpoint examples anti-patterns` - Learn what to avoid

**All resources are in `.checkpoint/` directory:**
- `.checkpoint/examples/` - Example checkpoints
- `.checkpoint/guides/` - Detailed documentation
- `.checkpoint/prompts/` - LLM prompt templates

## Quick Tips

**For LLMs:**
- Use `checkpoint prompt fill-checkpoint` to get instructions for filling checkpoint entries
- Derive distinct changes from git_status and .checkpoint-diff
- Keep summaries <80 chars; present tense; consistent scope names
- Fill context section with reasoning and decision-making process
- Mark context items with scope: project if they represent project-wide patterns
- Run `checkpoint lint` before finishing
- See `checkpoint guide llm-workflow` for detailed instructions

**For Humans:**
- Run `checkpoint start` at the beginning of each session
- Be specific in change summaries (not "fix bug" but "fix null pointer in user profile")
- Explain WHY in context, not just WHAT changed
- Document alternatives you considered
- Mark project-wide patterns with scope: project
- See `checkpoint guide best-practices` for more tips

## Commands

- `checkpoint start` - Check readiness and show next steps
- `checkpoint check` - Generate input files
- `checkpoint lint` - Validate before committing
- `checkpoint commit` - Commit with checkpoint metadata
- `checkpoint examples [category]` - View examples
- `checkpoint guide [topic]` - View guides
- `checkpoint prompt [id]` - View LLM prompts
- `checkpoint summary` - Show project overview
- `checkpoint clean` - Abort and restart
- `checkpoint init` - Initialize checkpoint in project
- `checkpoint help` - Show all commands

## LLM Prompts

Use the prompts system for consistent, high-quality interactions:

```bash
checkpoint prompt                          # List available prompts
checkpoint prompt fill-checkpoint          # Get checkpoint fill instructions
checkpoint prompt implement-feature \
  --var feature_name="Auth" \
  --var priority="high"                    # Feature implementation with variables
```

Prompts support variable substitution:
- Automatic: `{{project_name}}`, `{{project_path}}`
- Global: defined in `.checkpoint/prompts/prompts.yaml`
- User: provided via `--var` flags

Customize prompts by editing files in `.checkpoint/prompts/`.
