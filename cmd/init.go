package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-llm/internal/changelog"
	"go-llm/internal/file"
	"go-llm/internal/project"
	"go-llm/pkg/config"
)

// Init creates a CHECKPOINT.md file with practical instructions and theory
func Init(projectPath string, version string) {
	// Create .checkpoint/ directory structure
	checkpointDir := filepath.Join(projectPath, ".checkpoint")
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating .checkpoint directory: %v\n", err)
		os.Exit(1)
	}

	// Create subdirectories
	subdirs := []string{"examples", "guides", "prompts", "templates"}
	for _, subdir := range subdirs {
		path := filepath.Join(checkpointDir, subdir)
		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating .checkpoint/%s directory: %v\n", subdir, err)
			os.Exit(1)
		}
	}

	// Create .gitignore for .checkpoint directory
	gitignorePath := filepath.Join(checkpointDir, ".gitignore")
	gitignoreContent := `# Checkpoint directory is tracked
# This file ensures the directory structure is preserved in git
`
	if err := file.WriteFile(gitignorePath, gitignoreContent); err != nil {
		fmt.Fprintf(os.Stderr, "error creating .checkpoint/.gitignore: %v\n", err)
		os.Exit(1)
	}

	// Create README.md in .checkpoint directory
	readmePath := filepath.Join(checkpointDir, "README.md")
	readmeContent := `# Checkpoint Support Files

This directory contains supporting materials for the checkpoint workflow.

## Directory Structure

- **examples/** - Example checkpoint entries showing best practices
  - Good examples of features, bug fixes, refactorings
  - Context examples showing effective decision capture
  - Anti-patterns to avoid

- **guides/** - Detailed documentation for checkpoint users
  - First-time user walkthrough
  - LLM integration patterns
  - Context writing guidelines
  - Best practices

- **prompts/** - LLM prompt templates (planned)
  - Session start prompts
  - Checkpoint filling prompts
  - Review and curation prompts

- **templates/** - Customizable templates (planned)
  - Custom input templates
  - Project-specific patterns

## Usage

These files are referenced by checkpoint commands and can be read directly:

- Run ` + "`checkpoint examples`" + ` to view examples
- Run ` + "`checkpoint guide [topic]`" + ` to view guides
- LLMs can read these files directly when filling checkpoint entries

## Maintenance

- This directory is tracked in git
- Add new examples as you develop useful patterns
- Update guides as the workflow evolves
- Customize for your project's specific needs
`
	if err := file.WriteFile(readmePath, readmeContent); err != nil {
		fmt.Fprintf(os.Stderr, "error creating .checkpoint/README.md: %v\n", err)
		os.Exit(1)
	}

	// Initialize changelog with meta document
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if err := changelog.InitializeChangelog(changelogPath, version); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing changelog: %v\n", err)
		os.Exit(1)
	}

	// Initialize project file
	projectFilePath := filepath.Join(projectPath, config.ProjectFileName)
	projectName := filepath.Base(projectPath)
	if err := project.InitializeProjectFile(projectFilePath, projectName, nil); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing project file: %v\n", err)
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
- .checkpoint-input: Edit this during a checkpoint to describe changes and context
- .checkpoint-diff: Git diff context for the LLM
- .checkpoint-changelog.yaml: Append-only YAML changelog; one document per checkpoint with changes[]
- .checkpoint-context.yml: Append-only context log; captures reasoning and decisions
- .checkpoint-project.yml: Project-wide patterns and conventions (human-curated)
- .checkpoint-status.yaml: Last commit metadata for discovery

Concepts:
- One checkpoint equals one Git commit. The tool stages ALL changes when committing.
- The changelog document is appended before commit; after commit, the tool backfills commit_hash into the last document without another commit.
- Each document contains an array of changes; use concise summaries and optional details.
- Context captures the "why" behind decisions to maintain continuity across LLM sessions.
- Project file aggregates project-wide patterns; human curates from checkpoint recommendations.

Basic workflow:
1. Run: checkpoint check [path]
2. Open .checkpoint-input and fill:
   - changes[] - what changed
   - context - why and how (problem, decisions, insights, conversation)
   - next_steps[] - planned work
3. Run: checkpoint commit [path]
   - Stages all changes, creates a commit, backfills commit_hash
   - Appends to changelog, context, and project files
4. Periodically: review .checkpoint-project.yml recommendations and curate main document

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
context:
  problem_statement: "<what problem are we solving>"
  key_insights: [...]
  decisions_made: [...]
  established_patterns: [...]
  conversation_context: [...]
next_steps:
  - summary: "<planned action>"
    details: "<optional>"
    priority: "low|med|high"
    scope: "<component>"

LLM guidance:
- Derive distinct changes from git_status and .checkpoint-diff
- Keep summaries <80 chars; present tense; consistent scope names
- Fill context section with reasoning and decision-making process
- Mark context items with scope: project if they represent project-wide patterns
- Review project_context and recent_context in input for consistency
- Do not modify schema_version or timestamp; leave commit_hash empty
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing CHECKPOINT.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Created CHECKPOINT.md\n")
	fmt.Printf("✓ Created .checkpoint/ directory structure\n")
	fmt.Printf("  - .checkpoint/examples/   (for example checkpoints)\n")
	fmt.Printf("  - .checkpoint/guides/     (for detailed documentation)\n")
	fmt.Printf("  - .checkpoint/prompts/    (for LLM prompt templates)\n")
	fmt.Printf("  - .checkpoint/templates/  (for customizable templates)\n")
	fmt.Printf("\nNext: Add files to .checkpoint/, then run 'checkpoint start'\n")
}
