package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dmoose/checkpoint/internal/changelog"
	"github.com/dmoose/checkpoint/internal/context"
	"github.com/dmoose/checkpoint/internal/file"
	"github.com/dmoose/checkpoint/internal/git"
	"github.com/dmoose/checkpoint/internal/project"
	"github.com/dmoose/checkpoint/internal/schema"
	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
)

var commitOpts struct {
	dryRun        bool
	changelogOnly bool
	keepSession   bool
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().BoolVarP(&commitOpts.dryRun, "dry-run", "n", false, "Show commit message and staged files without committing")
	commitCmd.Flags().BoolVar(&commitOpts.changelogOnly, "changelog-only", false, "Stage only changelog instead of all changes")
	commitCmd.Flags().BoolVar(&commitOpts.keepSession, "keep-session", false, "Preserve session file after commit (default: cleared)")
}

var commitCmd = &cobra.Command{
	Use:   "commit [path]",
	Short: "Parse input, append to changelog, stage changes, and git commit",
	Long: `Validates input, creates YAML document, stages files, commits.
Then backfills commit hash into the last changelog document.`,
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
		CommitWithOptions(absPath, CommitOptions{
			DryRun:        commitOpts.dryRun,
			ChangelogOnly: commitOpts.changelogOnly,
			KeepSession:   commitOpts.keepSession,
		}, Version)
	},
}

type CommitOptions struct {
	DryRun        bool
	ChangelogOnly bool
	KeepSession   bool
}

// Commit implements Phase 3: parse input, append to changelog, git commit, write status
func Commit(projectPath string, version string) {
	CommitWithOptions(projectPath, CommitOptions{}, version)
}

func CommitWithOptions(projectPath string, opts CommitOptions, version string) {
	// Validate git repository
	if ok, err := git.IsGitRepository(projectPath); !ok {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: git repository check failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "hint: ensure you're in a git repository and have proper permissions\n")
		} else {
			fmt.Fprintf(os.Stderr, "error: %s is not a git repository\n", projectPath)
			fmt.Fprintf(os.Stderr, "hint: run 'git init' to initialize a repository\n")
		}
		os.Exit(1)
	}

	// Check if input file exists
	inputPath := filepath.Join(projectPath, config.InputFileName)
	if !file.Exists(inputPath) {
		fmt.Fprintf(os.Stderr, "error: input file not found at %s\n", inputPath)
		fmt.Fprintf(os.Stderr, "hint: run 'checkpoint check %s' to generate the input file\n", projectPath)
		fmt.Fprintf(os.Stderr, "or run 'checkpoint clean %s' to restart if needed\n", projectPath)
		os.Exit(1)
	}

	// Read and parse input file
	inputContent, err := file.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to read input file: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: check file permissions or try running 'checkpoint check %s' again\n", projectPath)
		os.Exit(1)
	}

	entry, err := schema.ParseInputFile(inputContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to parse input file: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: check YAML syntax in %s\n", inputPath)
		fmt.Fprintf(os.Stderr, "or run 'checkpoint clean %s' to restart\n", projectPath)
		os.Exit(1)
	}

	// Validate entry (comprehensive validation)
	if err := schema.ValidateEntry(entry); err != nil {
		fmt.Fprintf(os.Stderr, "error: validation failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: edit %s to fix the issues above\n", inputPath)
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
	// Generate commit message
	commitMsg := generateCommitMessage(entry)

	// Handle dry-run before making any changes
	if opts.DryRun {
		fmt.Printf("[dry-run] Would commit with message:\n%s\n", commitMsg)
		fmt.Printf("\n[dry-run] Files that would be staged:\n")
		if opts.ChangelogOnly {
			fmt.Printf("  - %s\n", config.ChangelogFileName)
		} else {
			fmt.Printf("  - All modified and untracked files (git add -A)\n")
		}
		return
	}

	// Initialize changelog with meta document if it doesn't exist
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if err := changelog.InitializeChangelog(changelogPath, version); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialize changelog: %v\n", err)
		os.Exit(1)
	}

	// Append to changelog (append-only)
	if err := changelog.AppendEntry(changelogPath, doc); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to append to changelog: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: check write permissions for %s\n", changelogPath)
		os.Exit(1)
	}

	// Append context entry
	contextPath := filepath.Join(projectPath, config.ContextFileName)
	contextEntry := context.CreateContextEntry(entry.Timestamp, entry.Context)
	if err := context.AppendContextEntry(contextPath, contextEntry); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to append context entry: %v\n", err)
	}

	// Generate project recommendations from context
	projectFilePath := filepath.Join(projectPath, config.ProjectFileName)
	recommendations := generateProjectRecommendations(entry.Context)
	if recommendations != nil {
		if err := project.AppendRecommendations(projectFilePath, entry.Timestamp,
			recommendations.Additions, recommendations.Updates, recommendations.Deletions); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to append project recommendations: %v\n", err)
		}
	}

	// Stage changes per options
	if opts.ChangelogOnly {
		if err := git.StageFile(projectPath, config.ChangelogFileName); err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to stage changelog: %v\n", err)
			fmt.Fprintf(os.Stderr, "hint: ensure git is working and the repository is not corrupted\n")
			os.Exit(1)
		}
	} else {
		if err := git.StageAll(projectPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to stage changes: %v\n", err)
			fmt.Fprintf(os.Stderr, "hint: check for git issues or run 'git status' to see what's wrong\n")
			os.Exit(1)
		}
	}

	// Commit
	var commitHash string
	commitHash, err = git.Commit(projectPath, commitMsg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to commit: %v\n", err)
		fmt.Fprintf(os.Stderr, "warning: changelog has been appended but not committed\n")
		fmt.Fprintf(os.Stderr, "hint: fix git issues and run 'checkpoint commit %s' again\n", projectPath)
		os.Exit(1)
	}

	// Update/backfill commit hash in changelog
	entry.CommitHash = commitHash
	if err := changelog.UpdateLastDocument(changelogPath, func(e *schema.CheckpointEntry) *schema.CheckpointEntry {
		e.CommitHash = commitHash
		return e
	}); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to backfill commit_hash in changelog: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: the commit succeeded, but you may need to manually add the commit hash\n")
	}

	// Read meta document for project metadata
	var projectID, pathHash string
	if meta, err := changelog.ReadMetaDocument(changelogPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to read meta document: %v\n", err)
	} else if meta != nil {
		projectID = meta.ProjectID
		pathHash = meta.PathHash
	}

	// Write status file (for macOS app discovery) with project metadata
	statusPath := filepath.Join(projectPath, config.StatusFileName)
	statusContent := generateStatusFile(entry, commitMsg, projectID, pathHash)
	if err := file.WriteFile(statusPath, statusContent); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write status file: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: this is non-fatal, but the macOS app may not discover this project\n")
	}

	// Clean up input, diff, and lock files
	if err := os.Remove(inputPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to remove input file: %v\n", err)
	}
	diffPath := filepath.Join(projectPath, config.DiffFileName)
	if file.Exists(diffPath) {
		if err := os.Remove(diffPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove diff file: %v\n", err)
		}
	}
	lockPath := filepath.Join(projectPath, config.LockFileName)
	if file.Exists(lockPath) {
		if err := os.Remove(lockPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove lock file: %v\n", err)
		}
	}

	// Clear session file unless --keep-session is set
	sessionPath := filepath.Join(projectPath, sessionFileName)
	if file.Exists(sessionPath) {
		if opts.KeepSession {
			fmt.Println("Session preserved (--keep-session)")
		} else {
			if err := os.Remove(sessionPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to remove session file: %v\n", err)
			} else {
				fmt.Println("Session cleared.")
			}
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

// generateProjectRecommendations extracts project-scoped items from checkpoint context
func generateProjectRecommendations(ctx context.CheckpointContext) *struct {
	Additions project.ProjectAdditions
	Updates   project.ProjectUpdates
	Deletions []project.ProjectDeletion
} {
	var additions project.ProjectAdditions
	var updates project.ProjectUpdates
	var deletions []project.ProjectDeletion
	hasRecommendations := false

	// Extract project-scoped insights
	for _, insight := range ctx.KeyInsights {
		if insight.Scope == "project" {
			additions.KeyInsights = append(additions.KeyInsights, project.Insight{
				Insight:   insight.Insight,
				Rationale: insight.Impact,
			})
			hasRecommendations = true
		}
	}

	// Extract project-scoped patterns
	for _, pattern := range ctx.EstablishedPatterns {
		if pattern.Scope == "project" {
			additions.EstablishedPatterns = append(additions.EstablishedPatterns, project.Pattern{
				Pattern:   pattern.Pattern,
				Rationale: pattern.Rationale,
				Examples:  pattern.Examples,
			})
			hasRecommendations = true
		}
	}

	// Extract project-scoped failed approaches
	for _, failed := range ctx.FailedApproaches {
		if failed.Scope == "project" {
			additions.FailedApproaches = append(additions.FailedApproaches, project.FailedApproach{
				Approach:       failed.Approach,
				WhyFailed:      failed.WhyFailed,
				LessonsLearned: failed.LessonsLearned,
			})
			hasRecommendations = true
		}
	}

	// Extract project-scoped decisions as design principles
	for _, decision := range ctx.DecisionsMade {
		if decision.Scope == "project" {
			additions.DesignPrinciples = append(additions.DesignPrinciples, project.Principle{
				Principle: decision.Decision,
				Rationale: decision.Rationale,
			})
			hasRecommendations = true
		}
	}

	if !hasRecommendations {
		return nil
	}

	return &struct {
		Additions project.ProjectAdditions
		Updates   project.ProjectUpdates
		Deletions []project.ProjectDeletion
	}{
		Additions: additions,
		Updates:   updates,
		Deletions: deletions,
	}
}

// generateStatusFile creates status file content for macOS app discovery
func generateStatusFile(entry *schema.CheckpointEntry, commitMsg string, projectID string, pathHash string) string {
	// include next_steps to be surfaced in the next check
	var b strings.Builder

	// Project metadata (if available)
	if projectID != "" {
		b.WriteString(fmt.Sprintf("project_id: \"%s\"\n", projectID))
	}
	if pathHash != "" {
		b.WriteString(fmt.Sprintf("path_hash: \"%s\"\n", pathHash))
	}

	// Commit metadata
	b.WriteString(fmt.Sprintf("last_commit_hash: \"%s\"\n", entry.CommitHash))
	b.WriteString(fmt.Sprintf("last_commit_timestamp: \"%s\"\n", entry.Timestamp))
	b.WriteString(fmt.Sprintf("last_commit_message: \"%s\"\n", commitMsg))
	b.WriteString("status: \"success\"\n")
	b.WriteString(fmt.Sprintf("changes_count: %d\n", len(entry.Changes)))
	if len(entry.NextSteps) > 0 {
		b.WriteString("next_steps:\n")
		for _, ns := range entry.NextSteps {
			b.WriteString("  - summary: \"")
			b.WriteString(strings.ReplaceAll(ns.Summary, "\"", "'"))
			b.WriteString("\"\n")
			if ns.Details != "" {
				b.WriteString("    details: \"" + strings.ReplaceAll(ns.Details, "\"", "'") + "\"\n")
			}
			if ns.Priority != "" {
				b.WriteString("    priority: \"" + ns.Priority + "\"\n")
			}
			if ns.Scope != "" {
				b.WriteString("    scope: \"" + ns.Scope + "\"\n")
			}
		}
	}
	return b.String()
}
