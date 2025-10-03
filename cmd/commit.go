package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-llm/internal/changelog"
	"go-llm/internal/file"
	"go-llm/internal/git"
	"go-llm/internal/schema"
	"go-llm/pkg/config"
)

// Commit implements Phase 3: parse input, append to changelog, git commit, write status
func Commit(projectPath string) {
	// Validate git repository
	if ok, err := git.IsGitRepository(projectPath); !ok {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: git repository check failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "error: %s is not a git repository\n", projectPath)
		}
		os.Exit(1)
	}

	// Check if input file exists
	inputPath := filepath.Join(projectPath, config.InputFileName)
	if !file.Exists(inputPath) {
		fmt.Fprintf(os.Stderr, "error: input file not found at %s\n", inputPath)
		fmt.Fprintf(os.Stderr, "run 'checkpoint check %s' first\n", projectPath)
		os.Exit(1)
	}

	// Read and parse input file
	inputContent, err := file.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to read input file: %v\n", err)
		os.Exit(1)
	}

	entry, err := schema.ParseInputFile(inputContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to parse input file: %v\n", err)
		fmt.Fprintf(os.Stderr, "please fix the YAML and try again\n")
		os.Exit(1)
	}

	// Validate entry
	if err := schema.ValidateEntry(entry); err != nil {
		fmt.Fprintf(os.Stderr, "error: validation failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "please fill in all required fields\n")
		os.Exit(1)
	}

	// Fill timestamp if missing
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Render changelog document (without git_status/diff_file)
	doc, err := schema.RenderChangelogDocument(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to render changelog document: %v\n", err)
		os.Exit(1)
	}

	// Append to changelog (append-only)
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if err := changelog.AppendEntry(changelogPath, doc); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to append to changelog: %v\n", err)
		os.Exit(1)
	}

// Stage all changes so the changelog reflects committed code state
	if err := git.StageAll(projectPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to stage changes: %v\n", err)
		os.Exit(1)
	}

	// Generate commit message
	commitMsg := generateCommitMessage(entry)

	// Commit
	commitHash, err := git.Commit(projectPath, commitMsg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to commit: %v\n", err)
		fmt.Fprintf(os.Stderr, "changelog has been appended but not committed\n")
		os.Exit(1)
	}

// Update entry with commit hash
	entry.CommitHash = commitHash

	// Backfill commit_hash into the last document of the changelog (no re-commit)
	if err := changelog.UpdateLastDocument(changelogPath, func(e *schema.CheckpointEntry) *schema.CheckpointEntry {
		e.CommitHash = commitHash
		return e
	}); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to backfill commit_hash in changelog: %v\n", err)
	}

	// Write status file (for macOS app discovery)
	statusPath := filepath.Join(projectPath, config.StatusFileName)
	statusContent := generateStatusFile(entry, commitMsg)
	if err := file.WriteFile(statusPath, statusContent); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write status file: %v\n", err)
		// Non-fatal; status file is informational
	}

	// Clean up input and diff files
	if err := os.Remove(inputPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to remove input file: %v\n", err)
	}
	diffPath := filepath.Join(projectPath, config.DiffFileName)
	if file.Exists(diffPath) {
		if err := os.Remove(diffPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove diff file: %v\n", err)
		}
	}

	fmt.Printf("✓ Checkpoint committed successfully\n")
	fmt.Printf("Commit: %s\n", commitHash)
	fmt.Printf("Changes: %d\n", len(entry.Changes))
	for i, c := range entry.Changes {
		scope := c.Scope
		if scope == "" {
			scope = "general"
		}
		fmt.Printf("  %d. %s (%s) - %s\n", i+1, c.ChangeType, scope, c.Summary)
	}
}

// generateCommitMessage creates a commit message summarizing the checkpoint
func generateCommitMessage(entry *schema.CheckpointEntry) string {
	if len(entry.Changes) == 1 {
		c := entry.Changes[0]
		scope := c.Scope
		if scope != "" {
			return fmt.Sprintf("Checkpoint: %s (%s) - %s", c.ChangeType, scope, c.Summary)
		}
		return fmt.Sprintf("Checkpoint: %s - %s", c.ChangeType, c.Summary)
	}

	// Multiple changes: summarize by type and scope
	types := make(map[string]int)
	scopes := make(map[string]int)
	for _, c := range entry.Changes {
		types[c.ChangeType]++
		if c.Scope != "" {
			scopes[c.Scope]++
		}
	}

	typeList := make([]string, 0, len(types))
	for t, count := range types {
		if count == 1 {
			typeList = append(typeList, t)
		} else {
			typeList = append(typeList, fmt.Sprintf("%s(%d)", t, count))
		}
	}

	scopeList := make([]string, 0, len(scopes))
	for s := range scopes {
		scopeList = append(scopeList, s)
	}

	msg := fmt.Sprintf("Checkpoint: %d changes - %s", len(entry.Changes), strings.Join(typeList, ", "))
	if len(scopeList) > 0 {
		msg += fmt.Sprintf(" [%s]", strings.Join(scopeList, ", "))
	}
	return msg
}

// generateStatusFile creates status file content for macOS app discovery
func generateStatusFile(entry *schema.CheckpointEntry, commitMsg string) string {
	return fmt.Sprintf(`last_commit_hash: "%s"
last_commit_timestamp: "%s"
last_commit_message: "%s"
status: "success"
changes_count: %d
`, entry.CommitHash, entry.Timestamp, commitMsg, len(entry.Changes))
}
