# Skill: Session Handoff

How to start, maintain, and hand off work between LLM sessions.

## Starting a Session

Run these at the beginning of every session:

```bash
# 1. Get project overview and current state
checkpoint start .

# 2. Understand the project
guardrail explain project

# 3. Read recent history for context
checkpoint search --recent 5

# 4. Check for handoff from previous session
cat .checkpoint-session.yaml 2>/dev/null
```

The handoff file (`.checkpoint-session.yaml`) contains:
- What the previous session was working on
- Decisions that were made
- What's pending / next steps
- Any blockers or open questions

## During a Session

- **Follow project guidelines**: `guardrail explain guidelines`
- **Check tools before guessing**: `guardrail explain tools`
- **Capture learnings as you go**: `guardrail learn --guideline "..."`
- **Checkpoint after each logical unit** — don't accumulate hours of uncheckpointed work
- **Track the key_exchanges** — note when a discussion leads to a decision

## Before Ending a Session

Create a handoff for the next session:

```bash
checkpoint session handoff
```

This captures:
- Current work state
- What was accomplished
- What's still in progress
- Pending decisions or blockers
- Next steps with priorities

### Manual Handoff Context

If `checkpoint session handoff` isn't available, ensure this information is captured in your last checkpoint's `next_steps` and `key_exchanges`:

```yaml
next_steps:
  - step: "Implement the retry wrapper tests — mock server is set up in test_helpers.go"
    priority: high
  - step: "Review whether the timeout config should be per-endpoint"
    priority: medium

key_exchanges:
  - topic: "Session wrap-up"
    summary: "Completed retry logic. Tests pass. Human wants integration tests next session. Config refactor deferred."
    participants: [human, claude]
```

## What Makes a Good Handoff

A good handoff lets the next session start working within 2 minutes. Include:

- **What's done**: Completed work since session start
- **What's in progress**: Partially completed work and its state
- **What's next**: Prioritized list of upcoming tasks
- **Key decisions**: Anything the next session needs to respect
- **Blockers**: Open questions that need human input
- **File locations**: Where the relevant code lives

## What to Avoid

- Don't leave uncheckpointed work — always checkpoint before handoff
- Don't assume the next session remembers anything — be explicit
- Don't write a novel — bullet points with specifics beat paragraphs
- Don't skip the `checkpoint start` at session beginning — it's how you pick up context
