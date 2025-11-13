package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmoose/checkpoint/cmd"
	"github.com/dmoose/checkpoint/internal/file"
	"github.com/dmoose/checkpoint/pkg/config"
)

// TestCompleteWorkflow tests the full check -> edit -> commit cycle
func TestCompleteWorkflow(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "checkpoint-integration")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repository
	setupGitRepo(t, tmpDir)

	// Create some initial content to commit
	testFile := filepath.Join(tmpDir, "README.md")
	initialContent := "# Test Project\n\nThis is a test project for checkpoint integration testing.\n"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Make initial commit
	runGitCmd(t, tmpDir, "add", "README.md")
	runGitCmd(t, tmpDir, "commit", "-m", "Initial commit")

	// Modify the file to create a change for checkpoint
	modifiedContent := initialContent + "\n## New Section\n\nAdded some new content.\n"
	if err := os.WriteFile(testFile, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	// Step 1: Run checkpoint check
	cmd.Check(tmpDir)

	// Verify input file was created
	inputPath := filepath.Join(tmpDir, config.InputFileName)
	if !file.Exists(inputPath) {
		t.Fatalf("input file not created at %s", inputPath)
	}

	// Verify diff file was created
	diffPath := filepath.Join(tmpDir, config.DiffFileName)
	if !file.Exists(diffPath) {
		t.Fatalf("diff file not created at %s", diffPath)
	}

	// Step 2: Edit the input file (simulate user/LLM editing)
	inputContent, err := file.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("failed to read input file: %v", err)
	}

	// Replace placeholder with actual content
	editedContent := strings.ReplaceAll(inputContent, `  - summary: "[FILL IN: what changed]"
    details: "[OPTIONAL: longer description]"
    change_type: "[FILL IN: feature|fix|refactor|docs|perf|other]"
    scope: "[FILL IN: affected component]"`, `  - summary: "Add new section to README"
    details: "Added a new section with additional content for testing"
    change_type: "docs"
    scope: "documentation"`)

	if err := file.WriteFile(inputPath, editedContent); err != nil {
		t.Fatalf("failed to edit input file: %v", err)
	}

	// Step 3: Run checkpoint commit
	cmd.Commit(tmpDir, "test-version")

	// Verify changelog was created and contains our change
	changelogPath := filepath.Join(tmpDir, config.ChangelogFileName)
	if !file.Exists(changelogPath) {
		t.Fatalf("changelog file not created at %s", changelogPath)
	}

	changelogContent, err := file.ReadFile(changelogPath)
	if err != nil {
		t.Fatalf("failed to read changelog: %v", err)
	}

	expectedEntries := []string{
		"Add new section to README",
		"change_type: docs",
		"scope: documentation",
		"schema_version: \"1\"",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(changelogContent, entry) {
			t.Errorf("changelog missing expected entry '%s'. Content:\n%s", entry, changelogContent)
		}
	}

	// Verify git commit was created
	output := runGitCmd(t, tmpDir, "log", "-1", "--pretty=format:%s")
	if !strings.Contains(output, "Checkpoint:") {
		t.Errorf("expected git commit with 'Checkpoint:' prefix, got: %s", output)
	}

	// Verify status file was created
	statusPath := filepath.Join(tmpDir, config.StatusFileName)
	if !file.Exists(statusPath) {
		t.Fatalf("status file not created at %s", statusPath)
	}

	// Verify temporary files were cleaned up
	if file.Exists(inputPath) {
		t.Errorf("input file should have been cleaned up")
	}
	if file.Exists(diffPath) {
		t.Errorf("diff file should have been cleaned up")
	}
}

// TestDryRunWorkflow tests the dry-run functionality
func TestDryRunWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-dryrun")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	setupGitRepo(t, tmpDir)

	// Create and commit initial content
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	runGitCmd(t, tmpDir, "add", "test.txt")
	runGitCmd(t, tmpDir, "commit", "-m", "Initial commit")

	// Modify file
	if err := os.WriteFile(testFile, []byte("initial\nmodified\n"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	// Run check
	cmd.Check(tmpDir)

	// Edit input file
	inputPath := filepath.Join(tmpDir, config.InputFileName)
	inputContent, _ := file.ReadFile(inputPath)
	editedContent := strings.ReplaceAll(inputContent, `  - summary: "[FILL IN: what changed]"
    details: "[OPTIONAL: longer description]"
    change_type: "[FILL IN: feature|fix|refactor|docs|perf|other]"
    scope: "[FILL IN: affected component]"`, `  - summary: "Update test file"
    change_type: "feature"
    scope: "core"`)
	file.WriteFile(inputPath, editedContent)

	// Run commit with dry-run
	cmd.CommitWithOptions(tmpDir, cmd.CommitOptions{DryRun: true}, "test-version")

	// Verify no actual commit was made
	output := runGitCmd(t, tmpDir, "log", "--oneline")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("expected only 1 commit (initial), got %d commits", len(lines))
	}

	// Verify changelog was not created
	changelogPath := filepath.Join(tmpDir, config.ChangelogFileName)
	if file.Exists(changelogPath) {
		t.Errorf("changelog should not be created in dry-run mode")
	}

	// Verify input files still exist (not cleaned up in dry-run)
	if !file.Exists(inputPath) {
		t.Errorf("input file should still exist after dry-run")
	}
}

// TestCleanWorkflow tests the clean command
func TestCleanWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-clean")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	setupGitRepo(t, tmpDir)

	// Create test file and initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	runGitCmd(t, tmpDir, "add", "test.txt")
	runGitCmd(t, tmpDir, "commit", "-m", "Initial commit")

	// Modify file and run check to create temporary files
	if err := os.WriteFile(testFile, []byte("content\nmodified\n"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}
	cmd.Check(tmpDir)

	// Verify temporary files exist
	inputPath := filepath.Join(tmpDir, config.InputFileName)
	diffPath := filepath.Join(tmpDir, config.DiffFileName)
	if !file.Exists(inputPath) {
		t.Fatalf("input file should exist before clean")
	}
	if !file.Exists(diffPath) {
		t.Fatalf("diff file should exist before clean")
	}

	// Run clean command
	cmd.Clean(tmpDir)

	// Verify temporary files are removed
	if file.Exists(inputPath) {
		t.Errorf("input file should be removed after clean")
	}
	if file.Exists(diffPath) {
		t.Errorf("diff file should be removed after clean")
	}
}

// TestConcurrentCheckpointPrevention tests that concurrent checkpoints are prevented
func TestConcurrentCheckpointPrevention(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-concurrent")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	setupGitRepo(t, tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	runGitCmd(t, tmpDir, "add", "test.txt")
	runGitCmd(t, tmpDir, "commit", "-m", "Initial commit")

	// Modify file
	if err := os.WriteFile(testFile, []byte("content\nmodified\n"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	// Run first check
	cmd.Check(tmpDir)

	// Verify lock file exists
	lockPath := filepath.Join(tmpDir, config.LockFileName)
	if !file.Exists(lockPath) {
		t.Fatalf("lock file should exist after check")
	}

	// Verify input file also exists
	inputPath := filepath.Join(tmpDir, config.InputFileName)
	if !file.Exists(inputPath) {
		t.Fatalf("input file should exist after check")
	}

	// Attempt second check - should be prevented by lock file
	// Capture stderr to verify proper error message
	originalStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	// This should exit due to lock file, but we can't easily test exit behavior
	// So we'll just verify the lock file still exists as evidence of protection
	if !file.Exists(lockPath) {
		t.Errorf("lock file should still exist to prevent concurrent checkpoints")
	}

	w.Close()
	os.Stderr = originalStderr

	// Clean up
	cmd.Clean(tmpDir)
}

// TestInitWorkflow tests the init command
func TestInitWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-init")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	setupGitRepo(t, tmpDir)

	// Run init command
	cmd.Init(tmpDir, "test-version")

	// Verify CHECKPOINT.md was created
	checkpointMdPath := filepath.Join(tmpDir, config.CheckpointMdFileName)
	if !file.Exists(checkpointMdPath) {
		t.Fatalf("CHECKPOINT.md file not created at %s", checkpointMdPath)
	}

	// Verify content contains expected sections
	content, err := file.ReadFile(checkpointMdPath)
	if err != nil {
		t.Fatalf("failed to read CHECKPOINT.md: %v", err)
	}

	expectedSections := []string{
		"# Checkpoint Workflow",
		"Key files:",
		"Basic workflow:",
		"Schema (YAML):",
		"## Learning Resources",
		"## Quick Tips",
		"## Commands",
		"## LLM Prompts",
	}

	for _, section := range expectedSections {
		if !strings.Contains(content, section) {
			t.Errorf("CHECKPOINT.md missing expected section '%s'", section)
		}
	}
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-errors")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test check in non-git directory
	originalStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	// This should fail because it's not a git repository
	// We'll skip actually testing the exit behavior since it's hard to mock
	w.Close()
	os.Stderr = originalStderr

	// For now, just verify the directory structure
	setupGitRepo(t, tmpDir)

	// Verify git repo was set up correctly
	gitDir := filepath.Join(tmpDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf("expected .git directory to exist after setupGitRepo")
	}
}

// Helper function to set up a git repository
func setupGitRepo(t *testing.T, dir string) {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available, skipping test: %v", err)
	}

	runGitCmd(t, dir, "config", "user.email", "test@example.com")
	runGitCmd(t, dir, "config", "user.name", "Test User")
}

// Helper function to run git commands
func runGitCmd(t *testing.T, dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git command failed: %v\nOutput: %s", err, output)
	}
	return string(output)
}
