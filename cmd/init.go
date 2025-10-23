package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-llm/internal/changelog"
	"go-llm/pkg/config"
)

// Init creates a CHECKPOINT.md file with practical instructions and theory
func Init(projectPath string, version string) {
	// Initialize changelog with meta document
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if err := changelog.InitializeChangelog(changelogPath, version); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing changelog: %v\n", err)
		os.Exit(1)
	}

	path := filepath.Join(projectPath, "CHECKPOINT.md")
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("CHECKPOINT.md already exists at %s\n", path)
		return
	}
	content := `# Checkpoint Workflow

This repository uses an append-only changelog to capture LLM-assisted development work.

Key files:
- .checkpoint-input: Edit this during a checkpoint to describe changes
- .checkpoint-diff: Git diff context for the LLM
- .checkpoint-changelog.yaml: Append-only YAML changelog; one document per checkpoint with changes[]
- .checkpoint-status.yaml: Last commit metadata for discovery

Concepts:
- One checkpoint equals one Git commit. The tool stages ALL changes when committing.
- The changelog document is appended before commit; after commit, the tool backfills commit_hash into the last document without another commit.
- Each document contains an array of changes; use concise summaries and optional details.

Basic workflow:
1. Run: checkpoint check [path]
2. Open .checkpoint-input and describe all changes in changes[] (and optional next_steps)
3. Run: checkpoint commit [path]
   - Stages all changes, creates a commit, backfills commit_hash in the last changelog doc

Schema (YAML):
---
schema_version: "1"
timestamp: "<auto>"
commit_hash: "<filled after commit>"
changes:
  - summary: "<what changed>"
    details: "<optional longer description>"
    change_type: "feature|fix|refactor|docs|perf|other"
    scope: "<component>"
next_steps:
  - summary: "<planned action>"
    details: "<optional>"
    priority: "low|med|high"
    scope: "<component>"

LLM guidance:
- Derive distinct changes from git_status and .checkpoint-diff
- Keep summaries <80 chars; present tense; consistent scope names
- Do not modify schema_version or timestamp; leave commit_hash empty
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing CHECKPOINT.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s\n", path)
}
