# Prompts System Design Specification

**Version:** 1.0  
**Status:** Ready for Implementation  
**Target:** Checkpoint 6

## Overview

The prompts system provides a project-wide library of curated LLM prompts for common development tasks. Checkpoint manages these prompts as project resources, allowing teams to maintain consistent, high-quality prompts that evolve with the project.

## Problem Statement

When co-developing with LLM agents, developers repeatedly craft similar prompts for common tasks:
- Starting development sessions
- Filling checkpoint entries
- Implementing features following project patterns
- Reviewing code against standards
- Debugging with project context

Without a prompt library:
- Quality varies (forgotten context, missing patterns)
- Repetition (retyping similar prompts)
- Inconsistency (different team members use different approaches)
- Lost knowledge (good prompts not captured)

## Solution

A prompt management system integrated with checkpoint:
- **Storage**: `.checkpoint/prompts/` directory with `prompts.yaml` index
- **Command**: `checkpoint prompt [id]` to list and display prompts
- **Variables**: Simple substitution of project info and user-provided values
- **Version control**: Prompts tracked in git, evolve with project

## User Stories

### Story 1: Developer uses checkpoint prompt
```bash
$ checkpoint prompt
# Lists available prompts by category

$ checkpoint prompt fill-checkpoint
# Outputs prompt with variables substituted
# Developer copies to LLM
```

### Story 2: Developer creates custom prompt
```bash
# Edit .checkpoint/prompts/prompts.yaml
# Add new prompt entry
# Create .checkpoint/prompts/my-prompt.md
# Use with: checkpoint prompt my-prompt
```

### Story 3: Developer uses variables
```bash
$ checkpoint prompt new-feature --var feature_name="user auth" --var priority="high"
# Variables substituted in template
```

### Story 4: Team evolves prompts
```bash
# Developer improves prompt based on experience
# Edit .checkpoint/prompts/implement-feature.md
# Commit changes
# Team benefits from improved prompt
```

## File Structure

```
.checkpoint/
├── prompts/
│   ├── prompts.yaml              # Index with metadata
│   ├── session-start.md          # Prompt templates
│   ├── fill-checkpoint.md
│   ├── implement-feature.md
│   ├── fix-bug.md
│   ├── code-review.md
│   └── (custom prompts...)
├── examples/
└── guides/
```

## Schema: prompts.yaml

```yaml
schema_version: "1"

# Global variables available to all prompts
variables:
  project_name: "checkpoint"
  primary_language: "Go"

# Prompt definitions
prompts:
  - id: session-start
    name: "Start Development Session"
    category: checkpoint
    description: "Orient LLM at beginning of work session"
    file: session-start.md
    variables:
      - task_description
    
  - id: fill-checkpoint
    name: "Fill Checkpoint Input"
    category: checkpoint
    description: "Analyze changes and create checkpoint entry"
    file: fill-checkpoint.md
    
  - id: implement-feature
    name: "Implement Feature"
    category: development
    description: "Implement new feature following project patterns"
    file: implement-feature.md
    variables:
      - feature_name
      - feature_description
      - priority
```

### Schema Fields

**Top level:**
- `schema_version` (string, required): Schema version (currently "1")
- `variables` (map, optional): Global variables available to all prompts
- `prompts` (array, required): List of prompt definitions

**Prompt definition:**
- `id` (string, required): Unique identifier (used in commands)
- `name` (string, required): Human-readable name
- `category` (string, required): Category (checkpoint, development, project-specific)
- `description` (string, required): Brief description of purpose
- `file` (string, required): Filename in .checkpoint/prompts/ directory
- `variables` (array, optional): Expected variables (for documentation)

## Prompt Template Format

Prompts are Markdown files with variable substitution.

### Example: fill-checkpoint.md

```markdown
# Fill Checkpoint Input

I've run `checkpoint check` which created:
- `.checkpoint-input` - Template to fill
- `.checkpoint-diff` - Full diff of changes

## Your Task

Analyze the changes and fill `.checkpoint-input` with:

### 1. Changes Section
- Read the diff carefully
- Group related changes into logical units
- Write specific summaries (<80 chars, present tense)
- Good: "Add JWT authentication middleware"
- Bad: "Update stuff"

### 2. Context Section
Explain WHY these changes were made:

**problem_statement:** What problem are we solving?

**key_insights:** What did we learn?
- Use `scope: project` for project-wide lessons
- Use `scope: checkpoint` for specific details

**decisions_made:** Why this approach?
- List alternatives considered
- Explain constraints

### 3. Next Steps
What should be done next? Prioritize as high/med/low.

## Project Info

Project: {{project_name}}
Language: {{primary_language}}

## Validation

Run `checkpoint lint` to check your work.
```

### Example: new-feature.md (with variables)

```markdown
# Implement Feature: {{feature_name}}

{{feature_description}}

Priority: {{priority}}

## Project: {{project_name}}

Follow these project patterns:
- Pattern 1
- Pattern 2

## Implementation Steps

1. Review existing similar features
2. Design approach
3. Implement with tests
4. Update documentation

## When Done

Run `checkpoint check` to create checkpoint entry.
```

## Variable System

### Variable Sources (Priority Order)

1. **Command line** (`--var key=value`): Highest priority, user-provided
2. **Global variables** (in prompts.yaml): Project-wide defaults
3. **Automatic** (checkpoint-determined): project_name, project_path

### Automatic Variables

Checkpoint automatically provides:
- `{{project_name}}` - Directory name (e.g., "checkpoint")
- `{{project_path}}` - Absolute path to project

### Variable Substitution

Simple string replacement: `{{variable_name}}` → value

**Rules:**
- Case-sensitive
- Unknown variables → empty string (no error)
- Variables must be valid identifiers ([a-z_][a-z0-9_]*)

**Example:**
```
Template: "Project: {{project_name}} Language: {{primary_language}}"
Variables: {project_name: "myapp", primary_language: "Python"}
Result:   "Project: myapp Language: Python"
```

## Command Interface

### `checkpoint prompt` - List prompts

```bash
$ checkpoint prompt
# or
$ checkpoint prompts

CHECKPOINT PROMPTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Available prompts:

Checkpoint Workflow:
  session-start          Start development session
  fill-checkpoint        Fill checkpoint input
  review-recommendations Review project recommendations

Development:
  implement-feature      Implement new feature
  fix-bug               Fix bug with investigation
  code-review           Review code changes

Usage: checkpoint prompt <id>
       checkpoint prompt <id> --var key=value
```

### `checkpoint prompt <id>` - Show prompt

```bash
$ checkpoint prompt fill-checkpoint

# Fill Checkpoint Input
# 
# I've run `checkpoint check` which created:
# - `.checkpoint-input` - Template to fill
# ...
```

### `checkpoint prompt <id> --var` - With variables

```bash
$ checkpoint prompt implement-feature \
  --var feature_name="User Authentication" \
  --var feature_description="Add JWT-based auth" \
  --var priority="high"

# Implement Feature: User Authentication
#
# Add JWT-based auth
#
# Priority: high
# ...
```

## Implementation Tasks

### Task 1: Create prompts package (internal/prompts)

**File:** `internal/prompts/prompts.go`

**Functions:**
```go
// Load prompts.yaml and parse
func LoadPromptsConfig(promptsDir string) (*PromptsConfig, error)

// Get list of all prompts
func ListPrompts(config *PromptsConfig) []PromptInfo

// Get specific prompt by ID
func GetPrompt(config *PromptsConfig, id string) (*Prompt, error)

// Load prompt template from file
func LoadPromptTemplate(promptsDir string, filename string) (string, error)

// Substitute variables in template
func SubstituteVariables(template string, vars map[string]string) string
```

**Structs:**
```go
type PromptsConfig struct {
    SchemaVersion string
    Variables     map[string]string
    Prompts       []PromptDefinition
}

type PromptDefinition struct {
    ID          string
    Name        string
    Category    string
    Description string
    File        string
    Variables   []string
}

type PromptInfo struct {
    ID          string
    Name        string
    Category    string
    Description string
}

type Prompt struct {
    Definition PromptDefinition
    Template   string
}
```

### Task 2: Create prompt command (cmd/prompt.go)

**File:** `cmd/prompt.go`

**Functions:**
```go
// Main entry point
func Prompt(projectPath string, promptID string, vars map[string]string)

// List all prompts
func listPrompts(promptsDir string)

// Show specific prompt
func showPrompt(promptsDir string, promptID string, vars map[string]string)

// Build variables map (automatic + global + user)
func buildVariables(projectPath string, globalVars map[string]string, userVars map[string]string) map[string]string
```

**Output format:**
- List: Grouped by category, aligned columns
- Show: Output raw markdown with variables substituted

### Task 3: Wire into main.go

**In main.go:**
```go
case "prompt", "prompts":
    // Parse --var flags
    vars := parseVarFlags(args)
    
    // Determine prompt ID (first positional arg)
    promptID := ""
    if len(positional) > 0 {
        promptID = positional[0]
    }
    
    cmd.Prompt(absPath, promptID, vars)
```

**Flag parsing:**
```go
func parseVarFlags(args []string) map[string]string {
    vars := make(map[string]string)
    
    for i := 0; i < len(args); i++ {
        if args[i] == "--var" && i+1 < len(args) {
            // Parse key=value
            parts := strings.SplitN(args[i+1], "=", 2)
            if len(parts) == 2 {
                vars[parts[0]] = parts[1]
            }
            i++ // Skip next arg
        }
    }
    
    return vars
}
```

### Task 4: Update help

**In cmd/help.go:**

Add to COMMANDS section:
```
prompt      Show LLM prompts from project library
            Display prompts with variable substitution
            Usage: checkpoint prompt [id] [--var key=value]
```

Add to EXAMPLES section:
```
checkpoint prompt                   # List available prompts
checkpoint prompt fill-checkpoint   # Show checkpoint fill prompt
checkpoint prompt new-feature --var feature_name="Auth" # With variables
```

### Task 5: Create initial prompts

**Files to create:**
1. `.checkpoint/prompts/prompts.yaml` - Index file
2. `.checkpoint/prompts/session-start.md`
3. `.checkpoint/prompts/fill-checkpoint.md`
4. `.checkpoint/prompts/implement-feature.md`
5. `.checkpoint/prompts/fix-bug.md`
6. `.checkpoint/prompts/code-review.md`

See "Initial Prompt Library" section below for content.

### Task 6: Update checkpoint init

**In cmd/init.go:**

When creating `.checkpoint/` directory:
```go
// Create prompts subdirectory
promptsDir := filepath.Join(checkpointDir, "prompts")
if err := os.MkdirAll(promptsDir, 0755); err != nil {
    // handle error
}

// Create prompts.yaml
promptsYaml := filepath.Join(promptsDir, "prompts.yaml")
if err := writeDefaultPromptsYaml(promptsYaml); err != nil {
    // handle error
}

// Create default prompt files
if err := writeDefaultPromptFiles(promptsDir); err != nil {
    // handle error
}
```

### Task 7: Tests

**Files:**
- `internal/prompts/prompts_test.go` - Unit tests for prompts package
- `cmd/prompt_test.go` - Command tests

**Test coverage:**
- Load prompts.yaml (valid, invalid, missing)
- List prompts (empty, populated)
- Get prompt by ID (exists, doesn't exist)
- Variable substitution (simple, multiple, missing vars)
- Command integration (list, show, with vars)

## Initial Prompt Library

### prompts.yaml

```yaml
schema_version: "1"

variables:
  project_name: "checkpoint"
  primary_language: "Go"

prompts:
  - id: session-start
    name: "Start Development Session"
    category: checkpoint
    description: "Orient LLM at beginning of work session"
    file: session-start.md
    variables:
      - task_description
    
  - id: fill-checkpoint
    name: "Fill Checkpoint Input"
    category: checkpoint
    description: "Analyze changes and create checkpoint entry"
    file: fill-checkpoint.md
    
  - id: implement-feature
    name: "Implement Feature"
    category: development
    description: "Implement new feature following project patterns"
    file: implement-feature.md
    variables:
      - feature_name
      - feature_description
      - priority
      
  - id: fix-bug
    name: "Fix Bug"
    category: development
    description: "Investigate and fix bug with testing"
    file: fix-bug.md
    variables:
      - bug_description
      
  - id: code-review
    name: "Code Review"
    category: development
    description: "Review code changes against project standards"
    file: code-review.md
```

### session-start.md

```markdown
# Development Session Start

I'm working on {{project_name}} ({{primary_language}}).

## Current Status

Run `checkpoint start` to see project status and next steps.

## Task

Let's work on: {{task_description}}

## Process

1. Implement the changes
2. Test thoroughly
3. When done, I'll run `checkpoint check`
4. You'll analyze changes and fill `.checkpoint-input`
5. I'll review and run `checkpoint commit`

## Project Patterns

Check `.checkpoint-project.yml` for established patterns and conventions.
```

### fill-checkpoint.md

```markdown
# Fill Checkpoint Input

I've run `checkpoint check` which created:
- `.checkpoint-input` - Template for you to fill
- `.checkpoint-diff` - Full diff of changes

## Your Task

Fill `.checkpoint-input` with structured information about these changes.

### Changes Section

List each distinct change:
- Summary: <80 chars, present tense, specific
- Details: Explain what and why (optional)
- Type: feature|fix|refactor|docs|perf|other
- Scope: Component affected

**Examples:**
- Good: "Add JWT authentication middleware to API"
- Bad: "Update auth code"

### Context Section (CRITICAL)

This is the most valuable part. Explain:

**problem_statement:** What problem did we solve?

**key_insights:** What did we learn?
- Mark `scope: project` for project-wide lessons
- Mark `scope: checkpoint` for specific details

**decisions_made:** Why this approach?
- List alternatives we considered
- Explain why we chose this
- Note constraints that influenced us
- Mark `scope: project` for architectural decisions

**established_patterns:** New conventions?
- Pattern description
- Why it works for this project
- Examples of where to apply
- Mark `scope: project`

**failed_approaches:** What didn't work?
- What we tried
- Why it failed
- Lessons learned

### Next Steps

What should happen next? Include:
- Summary of task
- Priority: high|med|low
- Scope/component

## Project Context

Project: {{project_name}}
Language: {{primary_language}}

Review `.checkpoint-project.yml` for established patterns.

## Validation

Run `checkpoint lint` to validate your work.

## Focus

Capture **reasoning and decisions**, not just descriptions.
The "why" is more valuable than the "what".
```

### implement-feature.md

```markdown
# Implement Feature: {{feature_name}}

{{feature_description}}

Priority: {{priority}}

## Project: {{project_name}}

Language: {{primary_language}}

## Process

1. **Understand requirements**
   - Clarify any ambiguities
   - Identify edge cases

2. **Review existing code**
   - Find similar features
   - Check `.checkpoint-project.yml` for patterns

3. **Design approach**
   - Consider alternatives
   - Think about testing
   - Plan error handling

4. **Implement**
   - Follow project patterns
   - Write clear code
   - Add comments for complex logic

5. **Test**
   - Unit tests for logic
   - Integration tests if needed
   - Edge cases and errors

6. **Document**
   - Update relevant docs
   - Add code comments
   - Update API docs if needed

## When Done

I'll run `checkpoint check` and you'll create the checkpoint entry explaining:
- What changed
- Why this approach
- What patterns were followed
- What's next
```

### fix-bug.md

```markdown
# Fix Bug

{{bug_description}}

## Project: {{project_name}}

## Investigation Process

1. **Reproduce the bug**
   - Minimal reproduction case
   - Identify conditions

2. **Understand root cause**
   - Use debugger/logging
   - Trace execution
   - Identify where it breaks

3. **Design fix**
   - Address root cause, not symptoms
   - Consider edge cases
   - Think about similar bugs

4. **Implement fix**
   - Minimal change to fix issue
   - Follow project patterns
   - Add defensive checks

5. **Test**
   - Verify fix works
   - Add regression test
   - Test edge cases

6. **Document**
   - Explain root cause in checkpoint
   - Document prevention strategy

## When Done

I'll run `checkpoint check` and you'll create checkpoint entry explaining:
- What was broken
- Root cause
- How fixed
- How to prevent similar bugs
```

### code-review.md

```markdown
# Code Review

## Project: {{project_name}}

Language: {{primary_language}}

## Review Checklist

### Code Quality
- [ ] Clear, readable code
- [ ] Follows project conventions (check `.checkpoint-project.yml`)
- [ ] Appropriate comments
- [ ] No obvious bugs
- [ ] Good error handling

### Design
- [ ] Appropriate abstractions
- [ ] Not over-engineered
- [ ] Fits with existing architecture
- [ ] Considers edge cases

### Testing
- [ ] Tests included
- [ ] Tests follow project patterns
- [ ] Edge cases covered
- [ ] Error cases tested

### Documentation
- [ ] Code comments where needed
- [ ] API docs updated
- [ ] README updated if needed

### Security/Performance
- [ ] No security issues
- [ ] Performance considerations
- [ ] Resource management

## Feedback Format

Provide feedback as:
- **MUST FIX**: Critical issues
- **SHOULD FIX**: Important improvements
- **CONSIDER**: Suggestions
- **GOOD**: Call out well-done parts

Be specific and constructive.
```

## Integration with Existing Commands

### Optional: checkpoint check integration

**Future enhancement:** After generating input, optionally output fill-checkpoint prompt:

```go
// In cmd/check.go, after successful generation:
if shouldOutputPrompt() {
    fmt.Println()
    fmt.Println("📋 LLM PROMPT:")
    fmt.Println(strings.Repeat("━", 60))
    
    // Load and display fill-checkpoint prompt
    if prompt := loadPrompt("fill-checkpoint"); prompt != "" {
        fmt.Println(prompt)
    }
    
    fmt.Println(strings.Repeat("━", 60))
    fmt.Println()
}
```

**Flag:** `checkpoint check --prompt` or `--no-prompt`

**Note:** This is optional and can be added after basic prompts system is working.

## Error Handling

**Missing prompts directory:**
```
Error: Prompts directory not found at .checkpoint/prompts/
Hint: Run 'checkpoint init' to initialize
```

**Invalid prompts.yaml:**
```
Error: Failed to parse .checkpoint/prompts/prompts.yaml
Details: yaml: line 5: mapping values are not allowed in this context
```

**Prompt ID not found:**
```
Error: Prompt 'unknown-id' not found
Run 'checkpoint prompt' to see available prompts
```

**Missing prompt file:**
```
Error: Prompt file 'missing.md' not found in .checkpoint/prompts/
Check prompts.yaml configuration
```

**Variable missing (non-fatal):**
```
Warning: Variable '{{undefined_var}}' not provided, using empty string
```

## Success Criteria

- [ ] `checkpoint prompt` lists all available prompts
- [ ] `checkpoint prompt <id>` displays prompt with substitution
- [ ] `checkpoint prompt <id> --var key=value` works correctly
- [ ] Automatic variables (project_name, project_path) substituted
- [ ] Global variables from prompts.yaml work
- [ ] `checkpoint init` creates prompts directory and files
- [ ] All tests pass
- [ ] Documentation updated (CHECKPOINT.md, guides)
- [ ] Examples work as documented

## Future Enhancements (Out of Scope)

These are NOT part of initial implementation:

1. **Command output variables:** `{{cmd:checkpoint start}}`
2. **Environment variables:** `{{env:USER}}`
3. **Conditional logic:** `{{#if}}...{{/if}}`
4. **Loops:** `{{#each}}...{{/each}}`
5. **Filters:** `{{variable|uppercase}}`
6. **Default values:** `{{variable|default:"fallback"}}`
7. **Auto-prompt from check:** `checkpoint check --prompt`
8. **Prompt categories filter:** `checkpoint prompt --category development`
9. **Prompt validation:** Check that expected variables are provided
10. **Prompt templates:** Inherit from base templates

## References

- Examples command: `cmd/examples.go` - Similar list/show pattern
- Guide command: `cmd/guide.go` - Similar structure
- Project file schema: `.checkpoint-project.yml` - Reference for structure

## Questions for Implementer

If any of these are unclear, ask before proceeding:

1. Should unknown variables be an error or warning?
2. Should we validate that expected variables are provided?
3. How should we handle multi-line variable values?
4. Should prompts support includes (one prompt references another)?
5. Should there be a `checkpoint prompt --edit <id>` command?

**Recommendation:** Keep it simple. Unknown vars → empty string (no error). No validation initially. No includes. No edit command (users edit files directly).