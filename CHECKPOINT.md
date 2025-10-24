# Checkpoint Workflow

This repository uses an append-only changelog to capture LLM-assisted development work.

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
1. Run: checkpoint check [path]
2. Open .checkpoint-input and fill:
   - changes[] - what changed
   - context - why and how (problem, decisions, insights, conversation)
   - next_steps[] - planned work
3. Run: checkpoint commit [path]
   - Stages all changes, creates a commit, backfills commit_hash
   - Appends to changelog, context, and project files
4. Periodically: review .checkpoint-project.yml recommendations and curate main document

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

LLM guidance:
- Derive distinct changes from git_status and .checkpoint-diff
- Keep summaries <80 chars; present tense; consistent scope names
- Fill context section with reasoning and decision-making process
- Mark context items with scope: project if they represent project-wide patterns
- Review project_context and recent_context in input for consistency
- Do not modify schema_version or timestamp; leave commit_hash empty
