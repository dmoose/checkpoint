# Checkpoint Workflow

This repository uses an append-only changelog to capture LLM-assisted development work.

Key files:
- .checkpoint-input: Edit this during a checkpoint to describe changes
- .checkpoint-diff: Git diff context for the LLM
- .checkpoint-changelog.yaml: Append-only YAML changelog; one document per checkpoint with changes[]
- .checkpoint-status.yaml: Last commit metadata for discovery

Concepts:
- One checkpoint equals one Git commit. The tool stages ALL changes when committing.
- The changelog document is appended before commit; after commit, the tool backfills commit_hash into the last document without another commit.
- Each document contains an array of changes; use concise summaries and optional details.

Basic workflow:
1. Human: Run: checkpoint check [path]
2. LLM: Fill .checkpoint-input with changes[] and next_steps, then run: checkpoint lint [path]
3. Human: Review and edit the input file as needed
4. Human: Run: checkpoint commit [path]
   - Stages all changes, creates a commit, backfills commit_hash in the last changelog doc
   - Cleans up input files, returning to step 1 for next checkpoint

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
next_steps:
  - summary: "<planned action>"
    details: "<optional>"
    priority: "low|med|high"
    scope: "<component>"

LLM guidance:
- Fill the changes[] array based on git_status and .checkpoint-diff context
- Derive distinct changes - group related file modifications into logical units
- Keep summaries <80 chars; present tense; consistent scope names
- Use the provided examples as a guide for good change entries
- Run 'checkpoint lint' after filling to catch obvious mistakes
- Do not modify schema_version or timestamp; leave commit_hash empty
- Remember: the human will review and edit your work before committing
