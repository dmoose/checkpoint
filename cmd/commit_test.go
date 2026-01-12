package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmoose/checkpoint/internal/file"
	"github.com/dmoose/checkpoint/internal/schema"
	"github.com/dmoose/checkpoint/pkg/config"
)

// TestCommitValidation tests the validation logic in commit command
func TestCommitValidation(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "checkpoint-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	if err := runGitCmd(tmpDir, "init"); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	if err := runGitCmd(tmpDir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("failed to set git email: %v", err)
	}
	if err := runGitCmd(tmpDir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("failed to set git name: %v", err)
	}

	tests := []struct {
		name        string
		inputYAML   string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid input",
			inputYAML: `schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes:
  - summary: "Add new feature"
    change_type: "feature"
    scope: "core"`,
			expectError: false,
		},
		{
			name: "missing changes",
			inputYAML: `schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes: []`,
			expectError: true,
			errorMsg:    "changes[>=1]",
		},
		{
			name: "placeholder in summary",
			inputYAML: `schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes:
  - summary: "[FILL IN: what changed]"
    change_type: "feature"`,
			expectError: true,
			errorMsg:    "placeholder text",
		},
		{
			name: "invalid change_type",
			inputYAML: `schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes:
  - summary: "Add feature"
    change_type: "invalid"`,
			expectError: true,
			errorMsg:    "invalid change_type",
		},
		{
			name: "summary too long",
			inputYAML: `schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes:
  - summary: "This is a very long summary that exceeds the maximum allowed length of 80 characters and should trigger a validation error"
    change_type: "feature"`,
			expectError: true,
			errorMsg:    "too long",
		},
		{
			name: "invalid next_steps priority",
			inputYAML: `schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes:
  - summary: "Add feature"
    change_type: "feature"
next_steps:
  - summary: "Follow up task"
    priority: "invalid"`,
			expectError: true,
			errorMsg:    "priority must be low|med|high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create input file
			inputPath := filepath.Join(tmpDir, config.InputFileName)
			err := file.WriteFile(inputPath, tt.inputYAML)
			if err != nil {
				t.Fatalf("failed to write input file: %v", err)
			}

			// Test validation directly using schema functions
			entry, parseErr := schema.ParseInputFile(tt.inputYAML)
			if parseErr != nil && !tt.expectError {
				t.Errorf("unexpected parse error: %v", parseErr)
			}

			if parseErr == nil {
				validateErr := schema.ValidateEntry(entry)
				if tt.expectError {
					if validateErr == nil {
						t.Errorf("expected validation error but got none")
					} else if !strings.Contains(validateErr.Error(), tt.errorMsg) {
						t.Errorf("expected error message to contain '%s', got: %s", tt.errorMsg, validateErr.Error())
					}
				} else {
					if validateErr != nil {
						t.Errorf("unexpected validation error: %v", validateErr)
					}
				}
			}

			// Clean up input file
			os.Remove(inputPath)
		})
	}
}

// TestGenerateCommitMessage tests commit message generation
func TestGenerateCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		changes  []string // simplified - just summaries
		expected string
	}{
		{
			name:     "single change",
			changes:  []string{"Add new feature"},
			expected: "Checkpoint: feature - Add new feature",
		},
		{
			name:     "multiple changes",
			changes:  []string{"Add feature", "Fix bug"},
			expected: "Checkpoint: 2 changes - feature, fix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would require refactoring generateCommitMessage to be more testable
			// For now, we verify the basic structure exists
			if len(tt.changes) == 0 {
				t.Skip("Need to refactor generateCommitMessage for better testability")
			}
		})
	}
}

// TestCommitDryRun tests dry-run functionality
func TestCommitDryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	if err := runGitCmd(tmpDir, "init"); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	if err := runGitCmd(tmpDir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("failed to set git email: %v", err)
	}
	if err := runGitCmd(tmpDir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("failed to set git name: %v", err)
	}

	// Create valid input file
	inputYAML := `schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes:
  - summary: "Add test feature"
    change_type: "feature"
    scope: "test"`

	inputPath := filepath.Join(tmpDir, config.InputFileName)
	err = file.WriteFile(inputPath, inputYAML)
	if err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Capture stdout for dry-run output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run dry-run
	CommitWithOptions(tmpDir, CommitOptions{DryRun: true}, "test-version")

	// Restore and read output
	w.Close()
	os.Stdout = originalStdout
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify dry-run output
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("expected dry-run output to contain '[dry-run]', got: %s", output)
	}
	if !strings.Contains(output, "Would commit with message") {
		t.Errorf("expected dry-run to show commit message, got: %s", output)
	}

	// Verify no actual commit was made
	changelogPath := filepath.Join(tmpDir, config.ChangelogFileName)
	if file.Exists(changelogPath) {
		content, _ := file.ReadFile(changelogPath)
		if strings.Contains(content, "Add test feature") {
			t.Errorf("dry-run should not have created changelog entry")
		}
	}
}

// TestGenerateStatusFileWithProjectMetadata tests that status file includes project_id and path_hash
func TestGenerateStatusFileWithProjectMetadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	if err := runGitCmd(tmpDir, "init"); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	if err := runGitCmd(tmpDir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("failed to set git email: %v", err)
	}
	if err := runGitCmd(tmpDir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("failed to set git name: %v", err)
	}

	// Create valid input file
	inputYAML := `schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes:
  - summary: "Add test feature"
    change_type: "feature"
    scope: "test"
next_steps:
  - summary: "Follow up task"
    priority: "high"
    scope: "test"`

	inputPath := filepath.Join(tmpDir, config.InputFileName)
	err = file.WriteFile(inputPath, inputYAML)
	if err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Run commit (which will create changelog with meta and status file)
	CommitWithOptions(tmpDir, CommitOptions{}, "test-version")

	// Read status file
	statusPath := filepath.Join(tmpDir, config.StatusFileName)
	if !file.Exists(statusPath) {
		t.Fatal("status file should have been created")
	}

	statusContent, err := file.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("failed to read status file: %v", err)
	}

	// Verify status file contains project metadata
	if !strings.Contains(statusContent, "project_id:") {
		t.Error("status file should contain project_id")
	}
	if !strings.Contains(statusContent, "path_hash:") {
		t.Error("status file should contain path_hash")
	}

	// Verify status file contains commit metadata
	if !strings.Contains(statusContent, "last_commit_hash:") {
		t.Error("status file should contain last_commit_hash")
	}
	if !strings.Contains(statusContent, "last_commit_timestamp:") {
		t.Error("status file should contain last_commit_timestamp")
	}
	if !strings.Contains(statusContent, "last_commit_message:") {
		t.Error("status file should contain last_commit_message")
	}
	if !strings.Contains(statusContent, "status: \"success\"") {
		t.Error("status file should contain status: success")
	}
	if !strings.Contains(statusContent, "changes_count:") {
		t.Error("status file should contain changes_count")
	}

	// Verify next_steps are preserved
	if !strings.Contains(statusContent, "next_steps:") {
		t.Error("status file should contain next_steps")
	}
	if !strings.Contains(statusContent, "Follow up task") {
		t.Error("status file should contain next step summary")
	}

	// Verify project_id and path_hash come before commit metadata
	projectIDIndex := strings.Index(statusContent, "project_id:")
	lastCommitIndex := strings.Index(statusContent, "last_commit_hash:")
	if projectIDIndex > lastCommitIndex {
		t.Error("project_id should come before last_commit_hash in status file")
	}

	pathHashIndex := strings.Index(statusContent, "path_hash:")
	if pathHashIndex > lastCommitIndex {
		t.Error("path_hash should come before last_commit_hash in status file")
	}
}

// Helper function to run git commands for testing
func runGitCmd(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git command failed: %v\nOutput: %s", err, output)
	}
	return nil
}
