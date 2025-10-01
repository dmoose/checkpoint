## Document 1: System Architecture & Design Specification

# LLM-Assisted Dev Project Tracking System - Architecture Specification

## 1. Overview

A two-component system for structured change tracking across LLM-assisted development projects:

- **Go CLI Tool:** Manages checkpoint creation and changelog commits with enforced consistency
- **macOS Swift App:** Processes changelogs to generate daily knowledge summaries and planning interface

The system treats changelogs as append-only, timestamped narratives that can be back-edited with commit hashes on subsequent cycles. The workflow is human-curated (agent assists, user decides).

---

## 2. Workflow

### 2.1 User Checkpoint Cycle

```
User signals "I want to checkpoint"
  ↓
User runs: go-tool check [path]
  ↓
Tool generates input file with:
  - Embedded LLM instructions
  - Git status snapshot
  - Placeholder schema fields
  - Path to git diff file
  ↓
User shares input file content with LLM / LLM reads file
  ↓
LLM fills schema fields following instructions
  ↓
User reviews LLM output, edits if needed
  ↓
User runs: go-tool commit [path]
  ↓
Tool validates input file exists
  ↓
Tool parses input → renders changelog entry → appends to changelog
  ↓
Tool commits (changelog + code changes)
  ↓
Tool extracts commit hash
  ↓
Tool writes status file (hash + metadata)
  ↓
Tool removes input file + diff file
  ↓
Cycle complete
```

### 2.2 Mac App Summary Generation Cycle

```
User opens Mac app or automated trigger fires
  ↓
Mac app checks for lock file in output directory
  ↓
If lock exists: warn user, offer option to delete and proceed
  ↓
Mac app writes lock file
  ↓
For each configured project directory:
  - Read changelog file
  - Find most recent entry timestamp from last run
  - Collect all entries newer than that timestamp
  - Generate per-project daily markdown summary
  ↓
Generate cross-project daily summary from all project summaries
  ↓
If Apple Intelligence available: generate narrative summaries
  ↓
If not available: present structured data directly
  ↓
Write summary files to output directory
  ↓
Remove lock file
  ↓
Cycle complete
```

---

## 3. File Structure & Formats

### 3.1 Project Directory Layout

```
{project-root}/
├── .git/
├── src/
├── [code files]
├── .changelog          # Append-only changelog file (YAML or TOML)
├── .checkpoint-input   # Ephemeral input file (YAML or TOML) [.gitignore]
├── .checkpoint-diff    # Temp diff file for context (text) [.gitignore]
└── .checkpoint-status  # Status from last successful commit (YAML or TOML) [.gitignore]
```

All checkpoint-related files are `.gitignore`'d by default.

### 3.2 Changelog File (.changelog)

- **Format:** YAML or TOML
- **Structure:** Append-only with entries separated by delimiters
- **Entry Contents:** Schema-defined fields (version identifier included in each entry)
- **File Behavior:** Never truncated, only appended to
- **Parsing Strategy:** Mac app reads from end of file backward by date to find resume point

Example (YAML):

```yaml
---
schema_version: "1"
timestamp: "2025-10-21T16:15:00-07:00"
summary: "Implemented user authentication flow"
details: "Added JWT token validation middleware..."
change_type: "feature"
scope: "auth"
commit_hash: "abc123def456"
---
schema_version: "1"
timestamp: "2025-10-21T16:05:00-07:00"
summary: "Refactored database query layer"
details: "Moved query logic to separate module..."
change_type: "refactor"
scope: "database"
commit_hash: "xyz789uvw012"
---
```

### 3.3 Input File (.checkpoint-input)

- **Format:** YAML or TOML (same as changelog)
- **Lifetime:** Created by `check`, consumed by `commit`, deleted after
- **Contents:**
  - LLM instruction prompt at top
  - Git status snapshot
  - Schema fields with placeholders
  - Reference to diff file path
- **Editability:** User can edit between LLM generation and commit

Example (YAML):

```yaml
# INSTRUCTIONS FOR LLM:
# Fill in the fields below to describe the changes made.
# Follow the existing schema structure.
# Be concise but descriptive.
# Do not modify the schema_version or timestamp.

schema_version: "1"
timestamp: "2025-10-21T16:15:00-07:00"

# Git status output (informational):
git_status: |
  M  src/auth.go
  M  src/config.go
  ?? tests/auth_test.go

# Reference to diff context (path to git diff output):
diff_file: ".checkpoint-diff"

# SCHEMA FIELDS - Fill these in:
summary: "[FILL IN: one-line summary of changes]"
details: "[FILL IN: longer description of what was changed and why]"
change_type: "[FILL IN: feature|fix|refactor|docs|perf|other]"
scope: "[FILL IN: component or module affected]"
commit_hash: ""  # Leave empty; tool will fill this
```

### 3.4 Diff File (.checkpoint-diff)

- **Format:** Text (git diff output)
- **Lifetime:** Created by `check`, removed by `commit`
- **Purpose:** Context for LLM to understand code changes

### 3.5 Status File (.checkpoint-status)

- **Format:** YAML or TOML
- **Lifetime:** Written by `commit`, persists until next cycle
- **Purpose:** Informational; next input file can reference for context
- **Contents:** Commit hash, timestamp, any metadata useful for next checkpoint

Example:

```yaml
last_commit_hash: "abc123def456"
last_commit_timestamp: "2025-10-21T16:15:00-07:00"
last_commit_message: "Checkpoint: feature - Implemented user authentication flow"
status: "success"
```

---

## 4. Schema Definition

### 4.1 Schema Versioning

- Each changelog entry includes `schema_version` field
- Tool generates entries with current version
- Mac app is forgiving of schema mismatches (fields can be missing)
- Schema evolution handled by adding optional fields, never removing required ones
- Tool doesn't parse old entries; only generates current version

### 4.2 Base Schema Fields (v1)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `schema_version` | string | yes | Version identifier (e.g., "1") |
| `timestamp` | ISO8601 | yes | When checkpoint was created |
| `summary` | string | yes | One-line summary |
| `details` | string | no | Longer narrative |
| `change_type` | enum | yes | feature, fix, refactor, docs, perf, other |
| `scope` | string | no | Component or module affected |
| `commit_hash` | string | no | Git commit hash (filled by tool) |

Future versions can add fields like `breaking_changes`, `ticket_id`, `authors`, etc. without breaking existing parsing.

---

## 5. Go Tool Specification

### 5.1 Command: `check`

**Purpose:** Generate input file for LLM

**Invocation:**
```
go-tool check [optional-path]
```

**Behavior:**
- If no path given and current directory is git root, use current directory
- If no path given and current directory is not git root, error
- If path given, change to that directory (must be git root)
- Validate it's a git repository (fail if not)
- Check if input file already exists (fail if it does; another checkpoint in progress)
- Parse `git status`
- Generate `git diff` output
- Load internal schema template
- Create input file with:
  - LLM instructions at top
  - Git status embedded
  - Diff file reference
  - Placeholder schema fields
- Create diff file alongside input file
- Output success message with path to input file and next steps

**Error Handling:**
- Not a git repository → error message + exit
- Input file already exists → error message (checkpoint already in progress) + exit
- Unable to read git status/diff → error message + exit

### 5.2 Command: `commit`

**Purpose:** Parse input, create changelog entry, commit code, write status

**Invocation:**
```
go-tool commit [optional-path]
```

**Behavior:**
- If no path given and current directory is git root, use current directory
- Change to target directory
- Validate it's a git repository (fail if not)
- Check if input file exists (fail if not; no checkpoint in progress)
- Parse input file as YAML/TOML
- Validate all required schema fields are present
- Render changelog entry via Go template using parsed fields
- Append changelog entry to `.changelog` file
- Stage both input file (or its removal) and `.changelog`
- Craft commit message from schema fields (template-based)
- Execute `git commit` with message
- Extract commit hash from git log
- Update input file with commit hash (or create new structure)
- Write status file with commit hash and metadata
- Delete input file
- Delete diff file
- Output success message with commit hash

**Error Handling:**
- Input file doesn't exist → error message (no checkpoint in progress) + exit
- Input file parsing fails → error message with parsing details, leave input file + exit
- Required fields missing → error message with field list, leave input file + exit
- Git commit fails → error message, attempt cleanup (depends on failure mode)
- Any other failure → error message, leave input file for manual inspection + exit

### 5.3 Command: `help`

**Purpose:** Display usage documentation

**Output:** Usage for `check` and `commit` subcommands

### 5.4 Internal Schema Template

Compiled into the tool. Used by `check` to generate input file. Defines:
- Required and optional fields
- Field types
- Validation rules
- LLM instruction prompt
- Go template for rendering changelog entries

---

## 6. macOS Swift App Specification

### 6.1 Functional Requirements

- Scan configured project directories for `.changelog` files
- Parse YAML/TOML changelog entries
- Generate per-project daily markdown summaries
- Generate cross-project daily summary
- Integrate Apple Intelligence for narrative generation (if available)
- Allow user to write/edit PLAN.md files for next steps
- Implement simple lock file mechanism for safe concurrent operation
- Run on-demand (user-triggered)

### 6.2 Architecture Components

#### 6.2.1 Configuration

- User specifies project directories to monitor
- User specifies output directory for summary files
- Store configuration in app preferences or config file

#### 6.2.2 Changelog Parser

- Read `.changelog` file
- Parse YAML/TOML entries
- Handle mixed schema versions gracefully (forgiving of missing fields)
- Extract timestamp, summary, details, change_type, scope, commit_hash

#### 6.2.3 Resume Logic

- Store last processed timestamp per project
- On each run, find all entries timestamped after last run
- If changelog deleted, regenerate by processing entire file
- Use filename-based sorting (date format YYYY-MM-DD sorts naturally)

#### 6.2.4 Summary Generation

- Per-project: Group entries by date, create markdown file
- Cross-project: Aggregate all daily changes across projects
- If Apple Intelligence available: generate narrative from structured data
- If not available: render structured data as formatted markdown

#### 6.2.5 Lock File Management

- On startup: check if `.checkpoint-lock` exists in output directory
- If exists: warn user, offer option to delete and proceed
- Write lock file at start of processing
- Delete lock file at end of processing
- Simple timeout: if lock file older than 2 hours, consider it stale

#### 6.2.6 Planning Interface

- Allow user to create/edit PLAN.md in each project directory
- PLAN.md can be git-ignored or tracked (user choice)
- Contents: freeform markdown for next steps / planning
- No enforced schema (placeholder for future)

### 6.3 Output Files

**Per-Project Daily Summary:**
- Filename: `{ProjectName}_YYYY-MM-DD.md`
- Location: Configured output directory
- Contents: Date, list of changes, Apple Intelligence summary if available

**Cross-Project Daily Summary:**
- Filename: `SUMMARY_YYYY-MM-DD.md`
- Location: Configured output directory
- Contents: All projects' changes that day, aggregated view

**Lock File:**
- Filename: `.checkpoint-lock`
- Location: Output directory
- Contents: Timestamp of lock creation (informational)

### 6.4 User Interaction

- Open app → shows configuration and summary generation options
- Click "Generate Today's Summary" → process all projects → show results
- Click "View Project Details" → show per-project summary
- Click "Edit Planning Notes" → open PLAN.md editor
- Lock file warning → inform of conflict, offer resolution

---

## 7. Data Flow Diagrams

### 7.1 Checkpoint Creation Flow

```
User → go-tool check
  ↓
Tool validates git root
Tool checks input file doesn't exist
Tool runs git status + git diff
Tool generates input file + diff file
  ↓
User sees: "Input file created at .checkpoint-input"
User shares/reviews input with LLM
LLM fills schema fields
User edits if needed
User → go-tool commit
```

### 7.2 Commit & Changelog Update Flow

```
go-tool commit
  ↓
Tool validates input file exists
Tool parses input file (YAML/TOML)
Tool validates required fields
Tool renders changelog entry via Go template
Tool appends entry to .changelog
Tool stages .changelog + code changes
Tool commits to git
Tool extracts commit hash from git log
Tool writes .checkpoint-status with hash
Tool removes input file + diff file
  ↓
User sees: "Commit successful: {hash}"
Cycle complete
```

### 7.3 Mac App Summary Generation Flow

```
Mac app user triggers summary generation
  ↓
App checks output dir for lock file
If lock exists: warn + offer delete/proceed
  ↓
App writes lock file
  ↓
For each configured project:
  Read .changelog file
  Find last processed timestamp (from app state or file)
  Collect entries since last timestamp
  Group by date
  Generate per-project markdown
  ↓
Aggregate all changes into cross-project summary
  ↓
If Apple Intelligence available: generate narratives
If not: render structured data
  ↓
Write summary files to output dir
App removes lock file
  ↓
User sees: "Summary files generated: {list of files}"
```

---

## 8. Error Scenarios & Recovery

### 8.1 Input File Stale (User Forgets to Commit)

- **Scenario:** User creates input file, leaves it for a day, tries to create new checkpoint
- **Tool Behavior:** `check` fails because input file exists
- **Recovery:** User must manually delete input file (with confirmation) or call `commit` to process it

### 8.2 Commit Fails After Parsing Success

- **Scenario:** Git commit fails (conflicts, permissions, etc.)
- **Tool Behavior:** Input file is left in place for user inspection
- **Recovery:** User resolves git issue, tries `commit` again

### 8.3 Diff File Parsing Fails

- **Scenario:** Diff file corrupted or unreadable
- **Tool Behavior:** Error message, input file left in place
- **Recovery:** User deletes files manually or tries `check` again

### 8.4 Lock File Stale (Mac App Crashed)

- **Scenario:** Mac app crashed, lock file remains
- **Tool Behavior:** Mac app detects stale lock (>2 hours old), warns user
- **Recovery:** User deletes lock and proceeds

### 8.5 Mixed Schema Versions

- **Scenario:** Old v1 entries exist, app receives v2 schema
- **Tool Behavior:** Mac app ignores missing v2 fields, gracefully handles v1 fields
- **Recovery:** No action needed; backwards compatible by design

---

## 9. Implementation Priorities

### Phase 1: Go Tool Foundation
- Implement `check` and `commit` subcommands
- YAML support (simpler than TOML to start)
- Schema v1 with base fields
- Git integration (status, diff, commit)

### Phase 2: Go Tool Refinement
- TOML support if needed
- Better error messages
- Testing and validation
- Documentation

### Phase 3: Mac App
- UI for configuration
- Changelog parsing
- Summary generation (without Apple Intelligence initially)
- Lock file management

### Phase 4: Mac App Enhancement
- Apple Intelligence integration
- Planning interface
- Persistent state management
- Summary history/archive

---

## 10. Configuration & Customization

### 10.1 Future Extensibility

- Schema versioning supports adding fields without breaking existing logic
- Template system allows customizing Go template output format
- Mac app configuration supports per-project custom summary styles
- Planning interface can evolve to structured format without breaking existing workflows

### 10.2 Current Scope (MVP)

- Single schema version (v1)
- Standard Go template for changelog rendering
- Standard markdown format for summaries
- No per-project customization

---
