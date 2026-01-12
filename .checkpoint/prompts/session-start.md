# Development Session Start

I'm working on {{project_name}} ({{primary_language}}).

## Current Status

Run `checkpoint start` to see project status and next steps.

## Task

Let's work on: {{task_description}}

## Process

### Basic Workflow
1. Implement the changes
2. Test thoroughly
3. When done, I'll run `checkpoint check`
4. You'll analyze changes and fill `.checkpoint-input`
5. I'll review and run `checkpoint commit`

### With Session Planning (for complex work)
1. Run `checkpoint plan` to create a planning session
2. Fill in goals, approach, and next actions
3. Work through the plan, updating progress
4. Run `checkpoint check` when ready
5. Fill `.checkpoint-input`
6. Run `checkpoint commit` (clears session)

Or use `checkpoint session handoff` to preserve session for later.

## Session Commands

```bash
checkpoint plan              # Create planning session
checkpoint session           # View current session
checkpoint session handoff   # Prepare for handoff to next LLM
```

## Project Patterns

Check `.checkpoint-project.yml` for established patterns and conventions.
