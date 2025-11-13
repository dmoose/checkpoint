# Future Features

### Changelog Verification
Validate changelog integrity for CI:
- Schema compliance
- Timestamps in order
- All entries (except last) have commit_hash
- Valid YAML structure

```bash
checkpoint verify  # or add to lint
```

### Git Hooks
Simple pre-commit hook to warn if committing without checkpoint:
```bash
checkpoint hooks install  # creates .git/hooks/pre-commit
```

### GitHub Action
Basic action for CI:
```yaml
- uses: dmoose/checkpoint-action@v1
  with:
    verify: true
```

## Maybe Later

### Archive Old Checkpoints
When changelog gets large (500+ entries), archive older entries:
```bash
checkpoint archive --before 2024-01-01
```
Creates `.checkpoint-archive/2023.yaml`, keeps recent in main file.

### Import from Git History
Bootstrap checkpoint from existing git history:
```bash
checkpoint import --since 2024-01-01
```
Creates basic entries from commit messages. One-time migration tool.
