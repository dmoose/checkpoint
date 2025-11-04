# checkpoint (Go CLI)

Checkpoint is a small, append-only changelog tool that helps you capture what changed when you commit code, and enables a macOS Swift app to generate daily summaries across projects.

It pairs a Go CLI with a simple YAML schema. Each "checkpoint" is one Git commit and one YAML document appended to a repository-local changelog. The macOS app discovers projects by the presence of the status file and reads the changelog to build per-day summaries.

## Architecture

- Go CLI (this repo)
  - `start`: validates readiness and shows next steps from last checkpoint
  - `check`: creates an editable input file with embedded instructions, current `git status`, and a path to a diff file
  - `commit`: validates the input, appends a YAML document to the changelog, stages all changes, creates a Git commit, then back-fills the commit hash into the last changelog document
  - `lint`: validates checkpoint input before committing
  - `init`: writes `CHECKPOINT.md` with workflow and schema guidance, creates `.checkpoint/` directory with examples, guides, and prompts
  - `examples`: shows example checkpoint entries demonstrating best practices
  - `guide`: displays detailed guides for checkpoint usage
  - `prompt`: shows LLM prompts from project library with variable substitution
  - `summary`: displays project overview and recent activity

- Files created in the repo
  - `.checkpoint-input` (untracked): the editable input with instructions and placeholders
  - `.checkpoint-diff` (untracked): context for the LLM/user (staged and unstaged diff)
  - `.checkpoint-changelog.yaml` (tracked): append-only YAML, one document per checkpoint with an array of changes
  - `.checkpoint-context.yml` (tracked): append-only context log capturing reasoning and decisions
  - `.checkpoint-project.yml` (tracked): project-wide patterns and conventions (human-curated)
  - `.checkpoint-status.yaml` (untracked): last-commit metadata with project identity for discovery
  - `.checkpoint/` (tracked): directory containing examples, guides, prompts, and templates

- macOS Swift app (separate project)
  - Recursively discovers projects by `.checkpoint-status.yaml`
  - Uses `project_id` and `path_hash` from status file for project identification and deduplication
  - Reads `.checkpoint-changelog.yaml` (multi-document YAML)
  - Buckets changes by local day and generates per-project and cross-project summaries

## Data model (YAML)

### Changelog Document

The `.checkpoint-changelog.yaml` file contains:
1. A meta document (first document, created once)
2. One checkpoint document per commit

**Meta Document (first document in changelog):**

```yaml
---
schema_version: "1"
document_type: "meta"
project_id: "01ARZ3NDEKTSV4RRFFQ69G5FAV"  # ULID, unique project identifier
path_hash: "abcdef1234567890"              # SHA256 hash of project path (first 16 chars)
created_at: "2025-01-01T12:00:00Z"
tool_version: "1.0.0"
languages: []                               # Detected project languages
```

**Checkpoint Document (one per commit):**

```yaml
---
schema_version: "1"
timestamp: "2025-10-22T16:00:00Z"
commit_hash: "abc123..."  # backfilled after commit
changes:
  - summary: "Implement feature X"
    details: "Optional longer description"
    change_type: "feature"   # feature|fix|refactor|docs|perf|other
    scope: "module/subsystem"
next_steps:
  - summary: "Plan follow-up work"
    details: "Optional"
    priority: "low|med|high"
    scope: "affected component"
```

### Status File

The `.checkpoint-status.yaml` file contains project identity and last commit metadata for discovery:

```yaml
project_id: "01ARZ3NDEKTSV4RRFFQ69G5FAV"  # Mirrored from changelog meta
path_hash: "abcdef1234567890"              # Mirrored from changelog meta
last_commit_hash: "abc123..."
last_commit_timestamp: "2025-10-22T16:00:00Z"
last_commit_message: "Checkpoint: feature - Add new feature"
status: "success"
changes_count: 3
next_steps:
  - summary: "Follow-up task"
    priority: "high"
    scope: "module"
```

The `project_id` and `path_hash` are copied from the changelog meta document during each commit, enabling:
- Unique project identification without parsing the full changelog
- Detection of moved/renamed projects via path_hash comparison
- Project deduplication across different discovery locations

## Workflow

1) Start session (optional but recommended)

- Run: `checkpoint start` to see project status, next steps, and validate readiness

2) Create an input

- Run: `checkpoint check [path]`
- Edit `.checkpoint-input`: describe all changes in `changes[]`, `context`, and any `next_steps[]`

3) Check your work (optional but recommended)

- Run: `checkpoint lint [path]` to catch obvious mistakes before commit

4) Commit

- Run: `checkpoint commit [flags] [path]`
  - Stages all changes (`git add -A`)
  - Appends a YAML document to `.checkpoint-changelog.yaml`
  - Appends context to `.checkpoint-context.yml`
  - Updates `.checkpoint-project.yml` with recommendations
  - Creates a single Git commit
  - Back-fills `commit_hash` in the last changelog document (no second commit)

5) Summaries

- Run: `checkpoint summary` to see project overview and recent activity
- macOS app reads `.checkpoint-changelog.yaml` and generates daily summaries

## Commands

- `checkpoint start [path]`
  - Validates readiness and shows next steps from last checkpoint
  - Checks git status and checkpoint initialization
  - Displays planned work

- `checkpoint check [path]`
  - Generates `.checkpoint-input` and `.checkpoint-diff`
  - Guards against concurrent checkpoints with lock files

- `checkpoint lint [path]`
  - Check checkpoint input for obvious mistakes and issues before commit
  - Validates input file and catches placeholder text, vague summaries, validation errors

- `checkpoint commit [path]`
  - Validates input, appends to `.checkpoint-changelog.yaml`, commits, and back-fills commit hash
  - Flags:
    - `-n, --dry-run`       Show the commit message and staged files without committing
    - `--changelog-only`    Stage only the changelog instead of all changes

- `checkpoint summary [path]`
  - Show project overview and recent activity
  - Display checkpoints, status, next steps, and patterns
  - Flags:
    - `--json`              Output machine-readable JSON format

- `checkpoint init [path]`
  - Writes `CHECKPOINT.md` with usage guidance
  - Creates `.checkpoint/` directory with examples, guides, and prompts
  - Safe to run multiple times (won't overwrite existing data files)

- `checkpoint clean [path]`
  - Removes `.checkpoint-input` and `.checkpoint-diff` to abort and restart

- `checkpoint examples [category] [path]`
  - Show example checkpoint entries and best practices
  - Available categories: feature, bugfix, refactor, context, anti-patterns

- `checkpoint guide [topic] [path]`
  - Show detailed guides and documentation
  - Available topics: first-time-user, llm-workflow, best-practices

- `checkpoint prompt [id] [path]`
  - Show LLM prompts from project library with variable substitution
  - Usage: `checkpoint prompt [id] [--var key=value]`
  - Examples: session-start, fill-checkpoint, implement-feature, fix-bug, code-review

- `checkpoint mcp [--root DIR]...`
  - Start MCP stdio server (read-only) for multi-project access by project_id
  - Roots precedence: `--root` (repeatable) > `CHECKPOINT_ROOTS` env (comma-separated) > `~/.config/checkpoint/config.json`
  - Tool: `project` with params `{ project_id }` returns structured project info; responses echo `project_id`

## Validation

Before committing, the tool validates:
- Required fields (`schema_version`, `timestamp`, and at least one item in `changes[]`)
- `change_type` must be one of: feature, fix, refactor, docs, perf, other
- Obvious mistakes:
  - Placeholder text like `[FILL IN ...]`
  - Overlong summaries (>80 chars)
  - Invalid `next_steps[].priority` (must be low|med|high)

If validation fails, errors are printed and the commit is aborted.

## Design choices

- Append-only changelog: one YAML document per checkpoint; corruption is isolated to the tail
- Backfill commit hash: the hash is a convenience written into the last document after the commit (without another commit)
- Stage-all by default: the changelog reflects exactly what was committed
- Minimal dependencies: standard library + `yaml.v3`

## Error Handling

The tool provides helpful error messages with hints for common issues:
- Missing input files suggest running `checkpoint check` first
- YAML parsing errors suggest checking syntax or running `checkpoint clean` to restart
- Git repository issues provide specific hints for resolution
- File permission problems indicate which files need attention

## LLM Prompts

The prompts system provides a library of curated LLM prompts for common development tasks:

```bash
checkpoint prompt                          # List available prompts
checkpoint prompt fill-checkpoint          # Show checkpoint fill prompt
checkpoint prompt implement-feature \
  --var feature_name="Auth" \
  --var priority="high"                    # Show prompt with variables
```

Prompts support variable substitution:
- Automatic variables: `{{project_name}}`, `{{project_path}}`
- Global variables: defined in `.checkpoint/prompts/prompts.yaml`
- User variables: provided via `--var` flags

Customize prompts by editing files in `.checkpoint/prompts/`.

## Typical .gitignore entries

```
.checkpoint-input
.checkpoint-diff
.checkpoint-status.yaml
.checkpoint-lock
```

These files are intentionally tracked:
- `.checkpoint-changelog.yaml` - changelog history
- `.checkpoint-context.yml` - context log
- `.checkpoint-project.yml` - project patterns
- `.checkpoint/` - examples, guides, and prompts
