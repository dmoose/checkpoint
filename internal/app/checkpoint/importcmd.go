package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/checkpoint/internal/changelog"
	"github.com/dmoose/checkpoint/internal/context"
	"github.com/dmoose/checkpoint/internal/file"
	"github.com/dmoose/checkpoint/internal/git"
	"github.com/dmoose/checkpoint/internal/schema"
	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
)

var importOpts struct {
	since     string
	last      int
	all       bool
	noEnrich  bool
	threshold int
}

func init() {
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(importCommitCmd)
	importCmd.Flags().StringVar(&importOpts.since, "since", "", "Import commits since date (e.g. 2024-01-01)")
	importCmd.Flags().IntVar(&importOpts.last, "last", 0, "Import last N commits")
	importCmd.Flags().BoolVar(&importOpts.all, "all", false, "Import all commits")
	importCmd.Flags().BoolVar(&importOpts.noEnrich, "no-enrich", false, "Skip enrichment, create mechanical entries only")
	importCmd.Flags().IntVar(&importOpts.threshold, "enrich-threshold", 0, "Only enrich commits with more than N lines changed (0 = enrich all)")
}

var importCmd = &cobra.Command{
	Use:   "import [commit...] [commit-range]",
	Short: "Import historical commits into checkpoint changelog",
	Long: `Create checkpoint entries from existing git history.

Without --no-enrich, generates checkpoint-input and .checkpoint-diff for each
commit so an LLM can fill in context. With --no-enrich, creates mechanical
entries marked as imports.

Examples:
  checkpoint import --last 10                  # Import last 10 commits
  checkpoint import --since 2024-01-01         # Import since date
  checkpoint import --last 10 --no-enrich      # Mechanical only, no LLM needed
  checkpoint import --last 20 --enrich-threshold 50  # Only enrich large commits
  checkpoint import abc123 def456              # Import specific commits`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		runImport(absPath, args)
	},
}

func runImport(projectPath string, args []string) {
	// Validate git repository
	if ok, err := git.IsGitRepository(projectPath); !ok {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: git repository check failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "error: %s is not a git repository\n", projectPath)
		}
		os.Exit(1)
	}

	// Get commits to import
	commits, err := getCommitsToImport(projectPath, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(commits) == 0 {
		fmt.Println("No commits to import.")
		return
	}

	// Reverse to chronological order (git log returns newest first)
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	// Initialize changelog if needed
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if err := changelog.InitializeChangelog(changelogPath, Build.Version); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialize changelog: %v\n", err)
		os.Exit(1)
	}

	// Check which commits are already imported
	existingHashes := getExistingCommitHashes(changelogPath)

	fmt.Printf("Found %d commit(s) to import\n\n", len(commits))

	imported := 0
	skipped := 0
	enrichable := 0

	for i, commit := range commits {
		// Skip already-imported commits
		if existingHashes[commit.Hash] {
			skipped++
			continue
		}

		// Get file stats
		numstat, _ := git.GetCommitNumStat(projectPath, commit.Hash)
		filesChanged := schema.ParseNumStat(numstat)
		totalLines := countTotalLines(filesChanged)

		shortHash := commit.Hash[:min(8, len(commit.Hash))]

		if importOpts.noEnrich {
			// Mechanical import — no LLM needed
			entry := buildMechanicalEntry(commit, filesChanged)

			doc, err := schema.RenderChangelogDocument(entry)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  error rendering entry for %s: %v\n", shortHash, err)
				continue
			}

			if err := changelog.AppendEntry(changelogPath, doc); err != nil {
				fmt.Fprintf(os.Stderr, "  error appending entry for %s: %v\n", shortHash, err)
				continue
			}

			// Also append a minimal context entry
			contextPath := filepath.Join(projectPath, config.ContextFileName)
			ctxEntry := context.CreateContextEntry(commit.Timestamp, entry.Context)
			_ = context.AppendContextEntry(contextPath, ctxEntry)

			fmt.Printf("  [%d/%d] %s %s (mechanical)\n", i+1, len(commits), shortHash, truncate(commit.Subject, 60))
			imported++
		} else {
			// Interactive enrichment mode
			shouldEnrich := importOpts.threshold == 0 || totalLines > importOpts.threshold

			if !shouldEnrich {
				// Below threshold — do mechanical import
				entry := buildMechanicalEntry(commit, filesChanged)
				doc, _ := schema.RenderChangelogDocument(entry)
				_ = changelog.AppendEntry(changelogPath, doc)

				contextPath := filepath.Join(projectPath, config.ContextFileName)
				ctxEntry := context.CreateContextEntry(commit.Timestamp, entry.Context)
				_ = context.AppendContextEntry(contextPath, ctxEntry)

				fmt.Printf("  [%d/%d] %s %s (mechanical, %d lines)\n", i+1, len(commits), shortHash, truncate(commit.Subject, 50), totalLines)
				imported++
				continue
			}

			// Generate input file for LLM enrichment
			diff, _ := git.GetCommitDiff(projectPath, commit.Hash)

			inputContent := generateImportInput(commit, filesChanged)
			inputPath := filepath.Join(projectPath, config.InputFileName)
			diffPath := filepath.Join(projectPath, config.DiffFileName)

			if err := file.WriteFile(inputPath, inputContent); err != nil {
				fmt.Fprintf(os.Stderr, "  error writing input file: %v\n", err)
				continue
			}
			if err := file.WriteFile(diffPath, diff); err != nil {
				fmt.Fprintf(os.Stderr, "  error writing diff file: %v\n", err)
				continue
			}

			enrichable++
			fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("IMPORT %d/%d: %s\n", i+1, len(commits), shortHash)
			fmt.Printf("Date:    %s\n", commit.Timestamp)
			fmt.Printf("Author:  %s\n", commit.Author)
			fmt.Printf("Subject: %s\n", commit.Subject)
			if commit.Body != "" {
				fmt.Printf("Body:    %s\n", truncate(commit.Body, 80))
			}
			fmt.Printf("Files:   %d changed (+%d/-%d lines)\n", len(filesChanged), countAdditions(filesChanged), countDeletions(filesChanged))
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
			fmt.Printf("Generated:\n")
			fmt.Printf("  Input: %s\n", inputPath)
			fmt.Printf("  Diff:  %s\n", diffPath)
			fmt.Printf("\nHave your LLM fill checkpoint-input, then run:\n")
			fmt.Printf("  checkpoint import-commit\n\n")
			fmt.Printf("Or skip this commit:\n")
			fmt.Printf("  checkpoint clean\n\n")
			fmt.Printf("Remaining: %d commit(s) after this one\n", len(commits)-i-1)

			// Stop here — user/LLM fills input, then runs import-commit
			return
		}
	}

	fmt.Printf("\nImport complete: %d imported, %d skipped (already in changelog)\n", imported, skipped)
	if enrichable > 0 {
		fmt.Printf("  %d commit(s) ready for enrichment\n", enrichable)
	}
}

// import-commit: append the filled input for a historical commit, then continue import

var importCommitCmd = &cobra.Command{
	Use:    "import-commit",
	Short:  "Commit a filled import entry and continue importing",
	Hidden: true, // internal command used during import workflow
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		runImportCommit(absPath)
	},
}

func runImportCommit(projectPath string) {
	inputPath := filepath.Join(projectPath, config.InputFileName)
	if !file.Exists(inputPath) {
		fmt.Fprintf(os.Stderr, "error: no checkpoint-input found\n")
		fmt.Fprintf(os.Stderr, "hint: run 'checkpoint import' first to generate input for a historical commit\n")
		os.Exit(1)
	}

	// Read and parse
	inputContent, err := file.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to read input file: %v\n", err)
		os.Exit(1)
	}

	entry, err := schema.ParseInputFile(inputContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to parse input file: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: check YAML syntax in %s\n", inputPath)
		os.Exit(1)
	}

	if err := schema.ValidateEntry(entry); err != nil {
		fmt.Fprintf(os.Stderr, "error: validation failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: edit %s to fix the issues above\n", inputPath)
		os.Exit(1)
	}

	// Append to changelog (no git commit — the commit already exists)
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if err := changelog.InitializeChangelog(changelogPath, Build.Version); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialize changelog: %v\n", err)
		os.Exit(1)
	}

	doc, err := schema.RenderChangelogDocument(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to render changelog document: %v\n", err)
		os.Exit(1)
	}

	if err := changelog.AppendEntry(changelogPath, doc); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to append to changelog: %v\n", err)
		os.Exit(1)
	}

	// Append context entry
	contextPath := filepath.Join(projectPath, config.ContextFileName)
	ctxEntry := context.CreateContextEntry(entry.Timestamp, entry.Context)
	if err := context.AppendContextEntry(contextPath, ctxEntry); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to append context entry: %v\n", err)
	}

	// Clean up
	_ = os.Remove(inputPath)
	diffPath := filepath.Join(projectPath, config.DiffFileName)
	_ = os.Remove(diffPath)

	shortHash := entry.CommitHash
	if len(shortHash) > 8 {
		shortHash = shortHash[:8]
	}

	fmt.Printf("✓ Imported %s: %s\n", shortHash, entry.Changes[0].Summary)
	fmt.Printf("\nRun 'checkpoint import' again to continue with the next commit.\n")
}

// getCommitsToImport determines which commits to import based on flags and args
func getCommitsToImport(projectPath string, args []string) ([]git.CommitInfo, error) {
	if len(args) > 0 {
		// Specific commits or range
		return git.GetLog(projectPath, args...)
	}

	var gitArgs []string

	if importOpts.all {
		// All commits
	} else if importOpts.since != "" {
		gitArgs = append(gitArgs, "--since="+importOpts.since)
	} else if importOpts.last > 0 {
		gitArgs = append(gitArgs, fmt.Sprintf("-n%d", importOpts.last))
	} else {
		return nil, fmt.Errorf("specify --since, --last, --all, or commit hashes\nusage: checkpoint import --last 10")
	}

	return git.GetLog(projectPath, gitArgs...)
}

// getExistingCommitHashes reads the changelog and returns a set of already-imported commit hashes
func getExistingCommitHashes(changelogPath string) map[string]bool {
	hashes := make(map[string]bool)
	if !file.Exists(changelogPath) {
		return hashes
	}

	content, err := file.ReadFile(changelogPath)
	if err != nil {
		return hashes
	}

	// Simple scan for commit_hash fields
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "commit_hash:") {
			hash := strings.TrimSpace(strings.TrimPrefix(line, "commit_hash:"))
			hash = strings.Trim(hash, "\"' ")
			if hash != "" {
				hashes[hash] = true
			}
		}
	}

	return hashes
}

// buildMechanicalEntry creates a checkpoint entry from git commit data without LLM enrichment
func buildMechanicalEntry(commit git.CommitInfo, filesChanged []schema.FileChange) *schema.CheckpointEntry {
	changeType := inferChangeType(commit.Subject)
	scope := inferScope(filesChanged)

	return &schema.CheckpointEntry{
		SchemaVersion: schema.SchemaVersion,
		Timestamp:     commit.Timestamp,
		CommitHash:    commit.Hash,
		Import:        true,
		FilesChanged:  filesChanged,
		Changes: []schema.Change{
			{
				Summary:    truncate(commit.Subject, schema.MaxSummaryLength),
				Details:    commit.Body,
				ChangeType: changeType,
				Scope:      scope,
			},
		},
		Context: context.CheckpointContext{
			ProblemStatement: commit.Subject,
		},
	}
}

// generateImportInput creates a checkpoint-input template for a historical commit
func generateImportInput(commit git.CommitInfo, filesChanged []schema.FileChange) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`# IMPORT MODE: Enriching historical commit %s
# Date: %s | Author: %s
# Original message: %s
#
# Fill in the context section based on the diff in .checkpoint-diff.
# The LLM should analyze the diff and infer:
#   - What problem was being solved
#   - What decisions were made and why
#   - What alternatives might have been considered
#   - Any patterns established
#
# When done, run: checkpoint import-commit
# To skip: checkpoint clean

`, commit.Hash[:min(8, len(commit.Hash))], commit.Timestamp, commit.Author, commit.Subject))

	b.WriteString(fmt.Sprintf("schema_version: \"%s\"\n", schema.SchemaVersion))
	b.WriteString(fmt.Sprintf("timestamp: \"%s\"\n", commit.Timestamp))
	b.WriteString(fmt.Sprintf("commit_hash: \"%s\"\n", commit.Hash))
	b.WriteString("import: true\n\n")

	// File changes
	if len(filesChanged) > 0 {
		b.WriteString("files_changed:\n")
		for _, f := range filesChanged {
			b.WriteString(fmt.Sprintf("  - path: \"%s\"\n    additions: %d\n    deletions: %d\n", f.Path, f.Additions, f.Deletions))
		}
		b.WriteString("\n")
	}

	// Pre-fill changes from commit message
	changeType := inferChangeType(commit.Subject)
	scope := inferScope(filesChanged)

	b.WriteString("changes:\n")
	b.WriteString(fmt.Sprintf("  - summary: \"%s\"\n", truncate(commit.Subject, schema.MaxSummaryLength)))
	if commit.Body != "" {
		b.WriteString(fmt.Sprintf("    details: \"%s\"\n", strings.ReplaceAll(commit.Body, "\"", "'")))
	}
	b.WriteString(fmt.Sprintf("    change_type: \"%s\"\n", changeType))
	if scope != "" {
		b.WriteString(fmt.Sprintf("    scope: \"%s\"\n", scope))
	}

	// Context template for enrichment
	b.WriteString(fmt.Sprintf(`
context:
  problem_statement: "[Infer from diff: what problem was this commit solving?]"

  key_insights:
    - insight: "[What can be learned from this change?]"
      scope: "checkpoint"

  decisions_made:
    - decision: "[What approach was chosen? Why?]"
      rationale: "[Infer from the diff]"
      alternatives_considered:
        - "[What else could have been done?]"
      scope: "checkpoint"

  failed_approaches: []

  key_exchanges: []

next_steps: []
`))

	return b.String()
}

// inferChangeType guesses the change type from a commit message
func inferChangeType(subject string) string {
	lower := strings.ToLower(subject)

	// Conventional commits
	if strings.HasPrefix(lower, "feat") {
		return "feature"
	}
	if strings.HasPrefix(lower, "fix") {
		return "fix"
	}
	if strings.HasPrefix(lower, "refactor") {
		return "refactor"
	}
	if strings.HasPrefix(lower, "docs") || strings.HasPrefix(lower, "doc:") {
		return "docs"
	}
	if strings.HasPrefix(lower, "perf") {
		return "perf"
	}

	// Checkpoint-style commits
	if strings.Contains(lower, "checkpoint:") {
		parts := strings.SplitN(lower, "-", 2)
		for _, ct := range []string{"feature", "fix", "refactor", "docs", "perf"} {
			if strings.Contains(parts[0], ct) {
				return ct
			}
		}
	}

	// Keyword inference
	if strings.Contains(lower, "fix") || strings.Contains(lower, "bug") || strings.Contains(lower, "patch") {
		return "fix"
	}
	if strings.Contains(lower, "add") || strings.Contains(lower, "implement") || strings.Contains(lower, "create") {
		return "feature"
	}
	if strings.Contains(lower, "refactor") || strings.Contains(lower, "extract") || strings.Contains(lower, "rename") || strings.Contains(lower, "move") {
		return "refactor"
	}
	if strings.Contains(lower, "doc") || strings.Contains(lower, "readme") || strings.Contains(lower, "comment") {
		return "docs"
	}
	if strings.Contains(lower, "update") || strings.Contains(lower, "upgrade") {
		return "other"
	}

	return "other"
}

// inferScope guesses a scope from the changed files
func inferScope(files []schema.FileChange) string {
	if len(files) == 0 {
		return ""
	}

	// Find common directory prefix
	dirs := make(map[string]int)
	for _, f := range files {
		parts := strings.Split(f.Path, "/")
		if len(parts) > 1 {
			dir := parts[0]
			if len(parts) > 2 {
				dir = parts[0] + "/" + parts[1]
			}
			dirs[dir]++
		}
	}

	// Return the most common directory
	bestDir := ""
	bestCount := 0
	for dir, count := range dirs {
		if count > bestCount {
			bestDir = dir
			bestCount = count
		}
	}

	return bestDir
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func countTotalLines(files []schema.FileChange) int {
	total := 0
	for _, f := range files {
		total += f.Additions + f.Deletions
	}
	return total
}

func countAdditions(files []schema.FileChange) int {
	total := 0
	for _, f := range files {
		total += f.Additions
	}
	return total
}

func countDeletions(files []schema.FileChange) int {
	total := 0
	for _, f := range files {
		total += f.Deletions
	}
	return total
}
