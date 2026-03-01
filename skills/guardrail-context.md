# Skill: Guardrail Context

## Reading Project Context

Use `guardrail explain` to understand a project before working on it:

```bash
guardrail explain project      # Full project overview
guardrail explain tools        # Available tools and commands
guardrail explain guidelines   # Coding patterns and rules to follow
guardrail explain skills       # Available skills and workflows
guardrail explain learnings    # Accumulated project knowledge
```

Always start a session with `guardrail explain project` to understand what you're working with.

## Following Guidelines

Guidelines live in `.checkpoint/guidelines.yaml`. They define:

- **Coding patterns** — How to structure code in this project
- **Naming conventions** — Variable, function, file naming rules
- **Error handling** — How errors should be reported
- **Testing requirements** — What needs tests and how to write them
- **Avoidances** — Things explicitly NOT to do

```bash
guardrail explain guidelines   # Read all active guidelines
```

Follow these as hard rules, not suggestions. They represent accumulated project decisions.

## Checking Available Tools

Tools and commands are defined in `.checkpoint/tools.yaml`:

```bash
guardrail explain tools        # See what's available
```

This tells you build commands, test commands, lint commands, and any project-specific scripts.

## Capturing New Knowledge

When you learn something during a session, capture it:

```bash
# A pattern or rule to follow
guardrail learn --guideline "Always use table-driven tests for validation logic"

# Something to avoid
guardrail learn --avoid "Don't use init() for registering commands; use explicit AddCommand"

# A general principle
guardrail learn --principle "Prefer stdlib over external deps unless the external dep saves significant complexity"
```

These get stored in `.checkpoint/` and are available to future sessions.

## The ai_authority Section

Projects can define what an LLM can do autonomously vs what needs human approval. Check this with:

```bash
guardrail explain project
```

Look for the `ai_authority` section. Typical rules:

- **Autonomous**: Run tests, format code, fix lint errors, read files
- **Needs approval**: Delete files, change public APIs, modify CI config, add dependencies
- **Never**: Push to main, force push, delete branches, modify credentials

If no `ai_authority` section exists, default to conservative: ask before destructive or irreversible actions.

## All Commands

| Command | Purpose |
|---------|---------|
| `guardrail explain [topic]` | Read project context (project, tools, guidelines, skills, learnings) |
| `guardrail learn` | Capture a new guideline, avoidance, or principle |
| `guardrail skill [name]` | Read or manage skill files |
| `guardrail doctor` | Check project setup health |
| `guardrail guide` | Display active guidelines |
| `guardrail examples` | Show example configurations |
| `guardrail prompt` | Load prompts from `.checkpoint/prompts/` |
