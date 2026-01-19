# Git

Version control system for tracking changes.

## Purpose

Track code changes, collaborate with others, and maintain project history.

## Common Commands

```bash
git status              # Show working tree status
git diff                # Show unstaged changes
git diff --staged       # Show staged changes
git add <file>          # Stage file for commit
git add -A              # Stage all changes
git commit -m "msg"     # Commit staged changes
git log --oneline -10   # Show recent commits
git branch              # List branches
git checkout <branch>   # Switch branches
git pull                # Fetch and merge remote
git push                # Push to remote
```

## Tips

- Use descriptive commit messages
- Commit often, push regularly
- Review changes before committing with git diff
- Use branches for features/experiments

## With Checkpoint

Always use `checkpoint commit` instead of raw `git commit` to maintain the changelog.
