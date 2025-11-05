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
