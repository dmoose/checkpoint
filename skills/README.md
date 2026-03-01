# Checkpoint & Guardrail Skills

This directory contains skill files that help LLMs work effectively with **checkpoint** (structured development history) and **guardrail** (project context and guidelines).

## Usage

Point your LLM at these files or include them in context:

- **checkpoint-workflow.md** — The core check/fill/commit cycle
- **guardrail-context.md** — Reading and contributing to project knowledge
- **filling-checkpoints.md** — Detailed guide for filling checkpoint-input correctly
- **session-handoff.md** — Handing off work between LLM sessions

## Quick Start

```
# Give an LLM the workflow skill
cat skills/checkpoint-workflow.md

# Give an LLM full context on a project
guardrail explain project
```

These are reference docs, not tutorials. Each file is designed to be consulted during work, not read end-to-end.
