package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/checkpoint/internal/changelog"
	"github.com/dmoose/checkpoint/internal/detect"
	"github.com/dmoose/checkpoint/internal/file"
	"github.com/dmoose/checkpoint/internal/project"
	"github.com/dmoose/checkpoint/internal/templates"
	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
)

var initOpts struct {
	template      string
	listTemplates bool
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initOpts.template, "template", "", "Use a specific template")
	initCmd.Flags().BoolVar(&initOpts.listTemplates, "list-templates", false, "List available templates")
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize checkpoint in a project",
	Long: `Creates .checkpoint/ directory structure and CHECKPOINT.md.
Auto-detects project language and sets up config files.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		if len(args) > 0 {
			projectPath = args[0]
		}
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		InitWithOptions(absPath, Version, InitOptions{Template: initOpts.template, ListTemplates: initOpts.listTemplates})
	},
}

// InitOptions holds flags for the init command
type InitOptions struct {
	Template      string // template name to use
	ListTemplates bool   // list available templates
}

// createDefaultPrompts creates the default prompts.yaml and prompt template files
// Only creates files that don't already exist
func createDefaultPrompts(promptsDir string, projectName string) error {
	created := 0
	skipped := 0

	// Create prompts.yaml
	promptsYaml := filepath.Join(promptsDir, "prompts.yaml")
	if file.Exists(promptsYaml) {
		skipped++
	} else {
		promptsContent := `schema_version: "1"

variables:
  project_name: "` + projectName + `"
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
`
		if err := file.WriteFile(promptsYaml, promptsContent); err != nil {
			return fmt.Errorf("failed to create prompts.yaml: %w", err)
		}
		created++
	}

	// Create session-start.md
	sessionStartPath := filepath.Join(promptsDir, "session-start.md")
	if file.Exists(sessionStartPath) {
		skipped++
	} else {
		sessionStartContent := `# Development Session Start

I'm working on {{project_name}} ({{primary_language}}).

## Current Status

Run ` + "`checkpoint start`" + ` to see project status and next steps.

## Task

Let's work on: {{task_description}}

## Process

1. Implement the changes
2. Test thoroughly
3. When done, I'll run ` + "`checkpoint check`" + `
4. You'll analyze changes and fill ` + "`.checkpoint-input`" + `
5. I'll review and run ` + "`checkpoint commit`" + `

## Project Patterns

Check ` + "`.checkpoint-project.yml`" + ` for established patterns and conventions.
`
		if err := file.WriteFile(sessionStartPath, sessionStartContent); err != nil {
			return fmt.Errorf("failed to create session-start.md: %w", err)
		}
		created++
	}

	// Create fill-checkpoint.md
	fillCheckpointPath := filepath.Join(promptsDir, "fill-checkpoint.md")
	if file.Exists(fillCheckpointPath) {
		skipped++
	} else {
		fillCheckpointContent := `# Fill Checkpoint Input

I've run ` + "`checkpoint check`" + ` which created:
- ` + "`.checkpoint-input`" + ` - Template for you to fill
- ` + "`.checkpoint-diff`" + ` - Full diff of changes

## Your Task

Fill ` + "`.checkpoint-input`" + ` with structured information about these changes.

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
- Mark ` + "`scope: project`" + ` for project-wide lessons
- Mark ` + "`scope: checkpoint`" + ` for specific details

**decisions_made:** Why this approach?
- List alternatives we considered
- Explain why we chose this
- Note constraints that influenced us
- Mark ` + "`scope: project`" + ` for architectural decisions

**established_patterns:** New conventions?
- Pattern description
- Why it works for this project
- Examples of where to apply
- Mark ` + "`scope: project`" + `

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

Review ` + "`.checkpoint-project.yml`" + ` for established patterns.

## Validation

Run ` + "`checkpoint lint`" + ` to validate your work.

## Focus

Capture **reasoning and decisions**, not just descriptions.
The "why" is more valuable than the "what".
`
		if err := file.WriteFile(fillCheckpointPath, fillCheckpointContent); err != nil {
			return fmt.Errorf("failed to create fill-checkpoint.md: %w", err)
		}
		created++
	}

	// Create implement-feature.md
	implementFeaturePath := filepath.Join(promptsDir, "implement-feature.md")
	if file.Exists(implementFeaturePath) {
		skipped++
	} else {
		implementFeatureContent := `# Implement Feature: {{feature_name}}

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
   - Check ` + "`.checkpoint-project.yml`" + ` for patterns

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

I'll run ` + "`checkpoint check`" + ` and you'll create the checkpoint entry explaining:
- What changed
- Why this approach
- What patterns were followed
- What's next
`
		if err := file.WriteFile(implementFeaturePath, implementFeatureContent); err != nil {
			return fmt.Errorf("failed to create implement-feature.md: %w", err)
		}
		created++
	}

	// Create fix-bug.md
	fixBugPath := filepath.Join(promptsDir, "fix-bug.md")
	if file.Exists(fixBugPath) {
		skipped++
	} else {
		fixBugContent := `# Fix Bug

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

I'll run ` + "`checkpoint check`" + ` and you'll create checkpoint entry explaining:
- What was broken
- Root cause
- How fixed
- How to prevent similar bugs
`
		if err := file.WriteFile(fixBugPath, fixBugContent); err != nil {
			return fmt.Errorf("failed to create fix-bug.md: %w", err)
		}
		created++
	}

	// Create code-review.md
	codeReviewPath := filepath.Join(promptsDir, "code-review.md")
	if file.Exists(codeReviewPath) {
		skipped++
	} else {
		codeReviewContent := `# Code Review

## Project: {{project_name}}

Language: {{primary_language}}

## Review Checklist

### Code Quality
- [ ] Clear, readable code
- [ ] Follows project conventions (check ` + "`.checkpoint-project.yml`" + `)
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
`
		if err := file.WriteFile(codeReviewPath, codeReviewContent); err != nil {
			return fmt.Errorf("failed to create code-review.md: %w", err)
		}
		created++
	}

	// Report what was done
	if created > 0 {
		fmt.Printf("✓ Created %d prompt file(s)\n", created)
	}
	if skipped > 0 {
		fmt.Printf("  Skipped %d existing prompt file(s)\n", skipped)
	}

	return nil
}

// updateProjectGitignore adds checkpoint artifact entries to the project's .gitignore
func updateProjectGitignore(projectPath string) {
	gitignorePath := filepath.Join(projectPath, ".gitignore")

	checkpointEntries := `
# Checkpoint artifacts (temporary files, not tracked)
.checkpoint-input
.checkpoint-diff
.checkpoint-lock
.checkpoint-status.yaml
.checkpoint-session.yaml
`

	// Check if .gitignore exists
	existingContent := ""
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existingContent = string(data)
	}

	// Check if checkpoint entries already exist
	if strings.Contains(existingContent, ".checkpoint-input") {
		// Already has checkpoint entries
		return
	}

	// Append checkpoint entries
	var newContent string
	if existingContent == "" {
		newContent = strings.TrimPrefix(checkpointEntries, "\n")
	} else {
		// Ensure there's a newline before our entries
		if !strings.HasSuffix(existingContent, "\n") {
			existingContent += "\n"
		}
		newContent = existingContent + checkpointEntries
	}

	if err := os.WriteFile(gitignorePath, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not update .gitignore: %v\n", err)
		return
	}

	if existingContent == "" {
		fmt.Println("✓ Created .gitignore with checkpoint artifacts")
	} else {
		fmt.Println("✓ Updated .gitignore with checkpoint artifacts")
	}
}

// createAutoDetectedConfigs creates config files based on auto-detected project info
func createAutoDetectedConfigs(checkpointDir string, projectPath string) {
	info := detect.DetectProject(projectPath)

	created := 0
	skipped := 0

	// Create project.yaml (check for legacy .yml too)
	projectYamlPath := filepath.Join(checkpointDir, config.ExplainProjectYaml)
	projectYmlLegacy := filepath.Join(checkpointDir, config.ExplainProjectYmlLegacy)
	if file.Exists(projectYamlPath) || file.Exists(projectYmlLegacy) {
		skipped++
	} else {
		var sb strings.Builder
		sb.WriteString("schema_version: \"1\"\n\n")
		sb.WriteString(fmt.Sprintf("name: %s\n", info.Name))
		if info.Description != "" {
			sb.WriteString(fmt.Sprintf("description: %s\n", info.Description))
		} else {
			sb.WriteString("description: \"\" # TODO: Add project description\n")
		}
		sb.WriteString(fmt.Sprintf("language: %s\n", info.Language))
		if len(info.Languages) > 1 {
			sb.WriteString("additional_languages:\n")
			for _, lang := range info.Languages[1:] {
				sb.WriteString(fmt.Sprintf("  - %s\n", lang))
			}
		}
		if len(info.Frameworks) > 0 {
			sb.WriteString("frameworks:\n")
			for _, fw := range info.Frameworks {
				sb.WriteString(fmt.Sprintf("  - %s\n", fw))
			}
		}
		sb.WriteString("\narchitecture:\n")
		sb.WriteString("  # TODO: Describe high-level architecture\n")
		sb.WriteString("  # pattern: MVC|microservices|monolith|cli|library\n")
		sb.WriteString("  # key_directories:\n")
		sb.WriteString("  #   - path: src/\n")
		sb.WriteString("  #     purpose: Source code\n")

		if err := file.WriteFile(projectYamlPath, sb.String()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create project.yaml: %v\n", err)
		} else {
			created++
		}
	}

	// Create tools.yaml (check for legacy .yml too)
	toolsYamlPath := filepath.Join(checkpointDir, config.ExplainToolsYaml)
	toolsYmlLegacy := filepath.Join(checkpointDir, config.ExplainToolsYmlLegacy)
	if file.Exists(toolsYamlPath) || file.Exists(toolsYmlLegacy) {
		skipped++
	} else {
		var sb strings.Builder
		sb.WriteString("schema_version: \"1\"\n\n")

		cmdCount := 0
		if info.BuildCmd != "" {
			sb.WriteString("# Build commands\n")
			sb.WriteString("build:\n")
			sb.WriteString("  default:\n")
			sb.WriteString(fmt.Sprintf("    command: %s\n", info.BuildCmd))
			sb.WriteString("    notes: Build the project\n\n")
			cmdCount++
		}
		if info.TestCmd != "" {
			sb.WriteString("# Test commands\n")
			sb.WriteString("test:\n")
			sb.WriteString("  default:\n")
			sb.WriteString(fmt.Sprintf("    command: %s\n", info.TestCmd))
			sb.WriteString("    notes: Run tests\n\n")
			cmdCount++
		}
		if info.LintCmd != "" {
			sb.WriteString("# Lint commands\n")
			sb.WriteString("lint:\n")
			sb.WriteString("  default:\n")
			sb.WriteString(fmt.Sprintf("    command: %s\n", info.LintCmd))
			sb.WriteString("    notes: Run linter\n\n")
			cmdCount++
		}
		if info.FormatCmd != "" {
			sb.WriteString("  format:\n")
			sb.WriteString(fmt.Sprintf("    command: %s\n", info.FormatCmd))
			sb.WriteString("    notes: Format code\n\n")
		}
		if info.DevCmd != "" {
			sb.WriteString("# Run commands\n")
			sb.WriteString("run:\n")
			sb.WriteString("  dev:\n")
			sb.WriteString(fmt.Sprintf("    command: %s\n", info.DevCmd))
			sb.WriteString("    notes: Run in development mode\n\n")
			cmdCount++
		}
		if info.CleanCmd != "" {
			sb.WriteString("# Maintenance commands\n")
			sb.WriteString("maintenance:\n")
			sb.WriteString("  clean:\n")
			sb.WriteString(fmt.Sprintf("    command: %s\n", info.CleanCmd))
			sb.WriteString("    notes: Clean build artifacts\n")
			cmdCount++
		}

		if cmdCount == 0 {
			sb.WriteString("# No commands detected. Add commands like:\n")
			sb.WriteString("# build:\n")
			sb.WriteString("#   default:\n")
			sb.WriteString("#     command: your-build-command\n")
			sb.WriteString("#     notes: Build the project\n")
			sb.WriteString("#\n")
			sb.WriteString("# test:\n")
			sb.WriteString("#   default:\n")
			sb.WriteString("#     command: your-test-command\n")
			sb.WriteString("#     notes: Run tests\n")
		}

		if err := file.WriteFile(toolsYamlPath, sb.String()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create tools.yaml: %v\n", err)
		} else {
			created++
		}
	}

	// Create guidelines.yaml (check for legacy .yml too)
	guidelinesYamlPath := filepath.Join(checkpointDir, config.ExplainGuidelinesYaml)
	guidelinesYmlLegacy := filepath.Join(checkpointDir, config.ExplainGuidelinesYmlLegacy)
	if file.Exists(guidelinesYamlPath) || file.Exists(guidelinesYmlLegacy) {
		skipped++
	} else {
		var sb strings.Builder
		sb.WriteString("schema_version: \"1\"\n\n")
		sb.WriteString("# Naming conventions\n")
		sb.WriteString("naming:\n")
		sb.WriteString("  # Add your naming conventions here\n")
		sb.WriteString("  # Example:\n")
		sb.WriteString("  # functions:\n")
		sb.WriteString("  #   exported: PascalCase\n")
		sb.WriteString("  #   internal: camelCase\n")
		sb.WriteString("\n")
		sb.WriteString("# Project structure guidelines\n")
		sb.WriteString("structure:\n")
		sb.WriteString("  # Describe how to add new components\n")
		sb.WriteString("  # Example:\n")
		sb.WriteString("  # new_command: |\n")
		sb.WriteString("  #   1. Create cmd/name.go\n")
		sb.WriteString("  #   2. Add to main.go switch\n")
		sb.WriteString("\n")
		sb.WriteString("# Rules to follow\n")
		sb.WriteString("rules:\n")
		sb.WriteString("  # - Run tests before committing\n")
		sb.WriteString("  # - All errors need user-friendly messages\n")
		sb.WriteString("\n")
		sb.WriteString("# Anti-patterns to avoid\n")
		sb.WriteString("avoid:\n")
		sb.WriteString("  # - Global state\n")
		sb.WriteString("  # - Silent failures\n")
		sb.WriteString("\n")
		sb.WriteString("# Design principles\n")
		sb.WriteString("principles:\n")
		sb.WriteString("  # - Keep it simple\n")
		sb.WriteString("  # - Prefer composition over inheritance\n")

		if err := file.WriteFile(guidelinesYamlPath, sb.String()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create guidelines.yaml: %v\n", err)
		} else {
			created++
		}
	}

	// Create skills.yaml (check for legacy .yml too)
	skillsYamlPath := filepath.Join(checkpointDir, config.ExplainSkillsYaml)
	skillsYmlLegacy := filepath.Join(checkpointDir, config.ExplainSkillsYmlLegacy)
	if file.Exists(skillsYamlPath) || file.Exists(skillsYmlLegacy) {
		skipped++
	} else {
		var sb strings.Builder
		sb.WriteString("schema_version: \"1\"\n\n")
		sb.WriteString("# Skills available for this project\n")
		sb.WriteString("# Local skills are in .checkpoint/skills/\n")
		sb.WriteString("# Global skills are in ~/.config/checkpoint/skills/\n")
		sb.WriteString("\n")
		sb.WriteString("# Uncomment to include global skills:\n")
		sb.WriteString("# global:\n")
		sb.WriteString("#   - git\n")
		sb.WriteString("#   - ripgrep\n")

		if err := file.WriteFile(skillsYamlPath, sb.String()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create skills.yaml: %v\n", err)
		} else {
			created++
		}
	}

	// Report what was detected and created
	if created > 0 || skipped > 0 {
		if info.Language != "" {
			fmt.Printf("✓ Detected: %s project", info.Language)
			if len(info.Frameworks) > 0 {
				fmt.Printf(" (%s)", strings.Join(info.Frameworks, ", "))
			}
			fmt.Println()
		}
		if created > 0 {
			fmt.Printf("✓ Created %d config file(s) with auto-detected settings\n", created)
		}
		if skipped > 0 {
			fmt.Printf("  Skipped %d existing config file(s)\n", skipped)
		}
	}
}

// ListTemplates prints available templates
func ListTemplates() {
	tmplList, err := templates.ListTemplates()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing templates: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available templates:")
	fmt.Println()
	for _, t := range tmplList {
		fmt.Printf("  %-12s  %s\n", t.Name, t.Description)
		if t.Source != "embedded" {
			fmt.Printf("               (from %s)\n", t.Source)
		}
	}
	fmt.Println()
	fmt.Println("Usage: checkpoint init --template <name> [path]")
}

// Init creates a CHECKPOINT.md file with practical instructions and theory
func Init(projectPath string, version string) {
	InitWithOptions(projectPath, version, InitOptions{})
}

// InitWithOptions creates checkpoint files with optional template
func InitWithOptions(projectPath string, version string, opts InitOptions) {
	// Handle --list-templates
	if opts.ListTemplates {
		ListTemplates()
		return
	}
	// Create .checkpoint/ directory structure
	checkpointDir := filepath.Join(projectPath, ".checkpoint")
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating .checkpoint directory: %v\n", err)
		os.Exit(1)
	}

	// Create subdirectories
	subdirs := []string{"examples", "guides", "prompts", "skills"}
	for _, subdir := range subdirs {
		path := filepath.Join(checkpointDir, subdir)
		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating .checkpoint/%s directory: %v\n", subdir, err)
			os.Exit(1)
		}
	}

	// Apply template if specified, otherwise auto-detect
	projectName := filepath.Base(projectPath)
	if opts.Template != "" {
		tmpl, err := templates.GetTemplate(opts.Template)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			fmt.Fprintf(os.Stderr, "hint: Run 'checkpoint init --list-templates' to see available templates\n")
			os.Exit(1)
		}
		if err := templates.ApplyTemplate(tmpl, projectPath, projectName); err != nil {
			fmt.Fprintf(os.Stderr, "error applying template: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Applied template '%s'\n", opts.Template)
	} else {
		// Auto-detect project settings and create config files
		createAutoDetectedConfigs(checkpointDir, projectPath)
	}

	// Update project .gitignore with checkpoint artifacts
	updateProjectGitignore(projectPath)

	// Create .gitignore for .checkpoint directory (to preserve empty dirs)
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

- **prompts/** - LLM prompt templates
  - Session start prompts
  - Checkpoint filling prompts
  - Feature implementation and bug fix prompts

- **skills/** - Skill definitions for LLM context
  - Local project-specific skills
  - References to global skills

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

	// Initialize changelog with meta document (only if it doesn't exist)
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if !file.Exists(changelogPath) {
		if err := changelog.InitializeChangelog(changelogPath, version); err != nil {
			fmt.Fprintf(os.Stderr, "error initializing changelog: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created %s\n", config.ChangelogFileName)
	} else {
		fmt.Printf("  %s already exists (skipped)\n", config.ChangelogFileName)
	}

	// Initialize project file (only if it doesn't exist)
	projectFilePath := filepath.Join(projectPath, config.ProjectFileName)
	if !file.Exists(projectFilePath) {
		if err := project.InitializeProjectFile(projectFilePath, projectName, nil); err != nil {
			fmt.Fprintf(os.Stderr, "error initializing project file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created %s\n", config.ProjectFileName)
	} else {
		fmt.Printf("  %s already exists (skipped)\n", config.ProjectFileName)
	}

	// Create default prompts
	promptsDir := filepath.Join(checkpointDir, "prompts")
	if err := createDefaultPrompts(promptsDir, projectName); err != nil {
		fmt.Fprintf(os.Stderr, "error creating default prompts: %v\n", err)
		os.Exit(1)
	}

	path := filepath.Join(projectPath, "CHECKPOINT.md")
	checkpointMdExists := file.Exists(path)
	content := `# Checkpoint Workflow - Quick Reference

This repository uses an append-only changelog to capture LLM-assisted development work.

**First time here?** Run ` + "`checkpoint guide first-time-user`" + ` for a complete walkthrough.

**Using with LLM?** Run ` + "`checkpoint guide llm-workflow`" + ` for integration patterns.

**Need examples?** Run ` + "`checkpoint examples`" + ` to see well-structured checkpoints.

Key files:
- .checkpoint-input: Edit this during a checkpoint to describe changes and context
- .checkpoint-diff: Git diff context for the LLM
- .checkpoint-changelog.yaml: Append-only YAML changelog; one document per checkpoint with changes[]
- .checkpoint-context.yml: Append-only context log; captures reasoning and decisions
- .checkpoint-project.yml: Project-wide patterns and conventions (human-curated)
- .checkpoint-status.yaml: Last commit metadata with project identity for discovery

Concepts:
- One checkpoint equals one Git commit. The tool stages ALL changes when committing.
- The changelog starts with a meta document containing project_id and path_hash for identity.
- The changelog document is appended before commit; after commit, the tool backfills commit_hash into the last document without another commit.
- Each document contains an array of changes; use concise summaries and optional details.
- Status file mirrors project_id and path_hash from the changelog meta for efficient discovery.
- Context captures the "why" behind decisions to maintain continuity across LLM sessions.
- Project file aggregates project-wide patterns; human curates from checkpoint recommendations.

Basic workflow:
1. Start: checkpoint start
   - Check project status and see next steps
2. Make changes: (code as usual)
3. Prepare: checkpoint check [path]
   - Generates .checkpoint-input and .checkpoint-diff
4. Fill input: Open .checkpoint-input and describe:
   - changes[] - what changed (be specific!)
   - context - why and how (problem, decisions, insights)
   - next_steps[] - planned work
5. Validate: checkpoint lint (optional but recommended)
6. Commit: checkpoint commit [path]
   - Stages all changes, creates a commit, backfills commit_hash
   - Appends to changelog, context, and project files
7. Periodically: review .checkpoint-project.yml recommendations and curate

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

## Learning Resources

**Comprehensive Guides:**
- ` + "`checkpoint guide first-time-user`" + ` - Complete walkthrough for newcomers
- ` + "`checkpoint guide llm-workflow`" + ` - LLM integration patterns and workflow
- ` + "`checkpoint guide best-practices`" + ` - Best practices for effective checkpoints

**Examples:**
- ` + "`checkpoint examples`" + ` - List all available examples
- ` + "`checkpoint examples feature`" + ` - See good feature checkpoint
- ` + "`checkpoint examples bugfix`" + ` - See good bug fix checkpoint
- ` + "`checkpoint examples anti-patterns`" + ` - Learn what to avoid

**All resources are in ` + "`.checkpoint/`" + ` directory:**
- ` + "`.checkpoint/examples/`" + ` - Example checkpoints
- ` + "`.checkpoint/guides/`" + ` - Detailed documentation
- ` + "`.checkpoint/prompts/`" + ` - LLM prompt templates

## Quick Tips

**For LLMs:**
- Use ` + "`checkpoint prompt fill-checkpoint`" + ` to get instructions for filling checkpoint entries
- Derive distinct changes from git_status and .checkpoint-diff
- Keep summaries <80 chars; present tense; consistent scope names
- Fill context section with reasoning and decision-making process
- Mark context items with scope: project if they represent project-wide patterns
- Run ` + "`checkpoint lint`" + ` before finishing
- See ` + "`checkpoint guide llm-workflow`" + ` for detailed instructions

**For Humans:**
- Run ` + "`checkpoint start`" + ` at the beginning of each session
- Be specific in change summaries (not "fix bug" but "fix null pointer in user profile")
- Explain WHY in context, not just WHAT changed
- Document alternatives you considered
- Mark project-wide patterns with scope: project
- See ` + "`checkpoint guide best-practices`" + ` for more tips

## Commands

- ` + "`checkpoint start`" + ` - Check readiness and show next steps
- ` + "`checkpoint check`" + ` - Generate input files
- ` + "`checkpoint lint`" + ` - Validate before committing
- ` + "`checkpoint commit`" + ` - Commit with checkpoint metadata
- ` + "`checkpoint examples [category]`" + ` - View examples
- ` + "`checkpoint guide [topic]`" + ` - View guides
- ` + "`checkpoint prompt [id]`" + ` - View LLM prompts
- ` + "`checkpoint summary`" + ` - Show project overview
- ` + "`checkpoint clean`" + ` - Abort and restart
- ` + "`checkpoint init`" + ` - Initialize checkpoint in project
- ` + "`checkpoint help`" + ` - Show all commands

## LLM Prompts

Use the prompts system for consistent, high-quality interactions:

` + "```bash" + `
checkpoint prompt                          # List available prompts
checkpoint prompt fill-checkpoint          # Get checkpoint fill instructions
checkpoint prompt implement-feature \
  --var feature_name="Auth" \
  --var priority="high"                    # Feature implementation with variables
` + "```" + `

Prompts support variable substitution:
- Automatic: ` + "`{{project_name}}`" + `, ` + "`{{project_path}}`" + `
- Global: defined in ` + "`.checkpoint/prompts/prompts.yaml`" + `
- User: provided via ` + "`--var`" + ` flags

Customize prompts by editing files in ` + "`.checkpoint/prompts/`" + `.
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing CHECKPOINT.md: %v\n", err)
		os.Exit(1)
	}
	if checkpointMdExists {
		fmt.Printf("✓ Updated CHECKPOINT.md\n")
	} else {
		fmt.Printf("✓ Created CHECKPOINT.md\n")
	}
	fmt.Printf("\n✓ Checkpoint initialization complete\n")
	fmt.Printf("  .checkpoint/ directory structure is ready\n")
	fmt.Printf("\nNext: Run 'checkpoint start' to begin\n")
}
