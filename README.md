# checkpoint (Go CLI)

Checkpoint is a small, append-only changelog tool that helps you capture what changed when you commit code, and enables a macOS Swift app to generate daily summaries across projects.

It pairs a Go CLI with a simple YAML schema. Each "checkpoint" is one Git commit and one YAML document appended to a repository-local changelog. The macOS app discovers projects by the presence of the status file and reads the changelog to build per-day summaries.

## Architecture

- Go CLI (this repo)
  - `check`: creates an editable input file with embedded instructions, current `git status`, and a path to a diff file
  - `commit`: validates the input, appends a YAML document to the changelog, stages all changes, creates a Git commit, then back-fills the commit hash into the last changelog document
  - `init`: writes `CHECKPOINT.md` with workflow and schema guidance

- Files created in the repo
  - `.checkpoint-input` (untracked): the editable input with instructions and placeholders
  - `.checkpoint-diff` (untracked): context for the LLM/user (staged and unstaged diff)
  - `.checkpoint-changelog.yaml` (tracked): append-only YAML, one document per checkpoint with an array of changes
  - `.checkpoint-status.yaml` (untracked): last-commit metadata for discovery

- macOS Swift app (separate project)
  - Recursively discovers projects by `.checkpoint-status.yaml`
  - Reads `.checkpoint-changelog.yaml` (multi-document YAML)
  - Buckets changes by local day and generates per-project and cross-project summaries

## Data model (YAML)

One YAML document per checkpoint:

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

## Workflow

1) Create an input

- Run: `checkpoint check [path]`
- Edit `.checkpoint-input`: describe all changes in `changes[]` and any `next_steps[]`

2) Commit

- Run: `checkpoint commit [flags] [path]`
  - Stages all changes (`git add -A`)
  - Appends a YAML document to `.checkpoint-changelog.yaml`
  - Creates a single Git commit
  - Back-fills `commit_hash` in the last changelog document (no second commit)

3) Summaries (macOS app)

- The app reads `.checkpoint-changelog.yaml` and generates daily summaries

## Commands

- `checkpoint check [path]`
  - Generates `.checkpoint-input` and `.checkpoint-diff`

- `checkpoint commit [path]`
  - Validates input, appends to `.checkpoint-changelog.yaml`, commits, and back-fills commit hash
  - Flags:
    - `-n, --dry-run`       Show the commit message and exit without committing
    - `--changelog-only`    Stage only the changelog instead of all changes

- `checkpoint init [path]`
  - Writes `CHECKPOINT.md` with usage guidance

## Validation

Before committing, the tool validates:
- Required fields (`schema_version`, `timestamp`, and at least one item in `changes[]`)
- `change_type` must be one of: feature, fix, refactor, docs, perf, other
- Obvious mistakes:
  - Placeholder text like `[FILL IN ...]`
  - Overlong summaries (>100 chars)
  - Invalid `next_steps[].priority` (must be low|med|high)

If validation fails, errors are printed and the commit is aborted.

## Design choices

- Append-only changelog: one YAML document per checkpoint; corruption is isolated to the tail
- Backfill commit hash: the hash is a convenience written into the last document after the commit (without another commit)
- Stage-all by default: the changelog reflects exactly what was committed
- Minimal dependencies: standard library + `yaml.v3`

## Typical .gitignore entries

```
.checkpoint-input
.checkpoint-diff
.checkpoint-status.yaml
```

The changelog (`.checkpoint-changelog.yaml`) is intentionally tracked.

## Future work

- `verify` command to lint the changelog and report issues
- Optional strict mode (scope allowlist)
- Template override per repo for prompts and fields
- TOML support if needed
