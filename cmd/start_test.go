package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmoose/checkpoint/internal/file"
	"github.com/dmoose/checkpoint/pkg/config"
)

func TestStartCommand(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "checkpoint-start-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repository
	setupGitRepo(t, tmpDir)

	// Create changelog to simulate initialized checkpoint
	changelogPath := filepath.Join(tmpDir, config.ChangelogFileName)
	changelogContent := `---
schema_version: "1"
document_type: meta
project_id: test123
---
schema_version: "1"
timestamp: "2025-01-01T10:00:00Z"
commit_hash: "abc123"
changes:
  - summary: "Initial checkpoint"
    change_type: "feature"
    scope: "test"
next_steps: []
`
	if err := file.WriteFile(changelogPath, changelogContent); err != nil {
		t.Fatalf("failed to create changelog: %v", err)
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	success := startInternal(tmpDir)

	w.Close()
	os.Stdout = originalStdout

	// Read output
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !success {
		t.Errorf("expected start to succeed")
	}

	// Verify output contains key indicators
	if !strings.Contains(output, "CHECKPOINT START") {
		t.Errorf("expected 'CHECKPOINT START' in output")
	}
	if !strings.Contains(output, "✓ Git repository detected") {
		t.Errorf("expected git repository confirmation")
	}
	if !strings.Contains(output, "✓ Checkpoint initialized") {
		t.Errorf("expected checkpoint initialization confirmation")
	}
	if !strings.Contains(output, "✓ No checkpoint in progress") {
		t.Errorf("expected no checkpoint in progress confirmation")
	}
	if !strings.Contains(output, "READY TO WORK") {
		t.Errorf("expected 'READY TO WORK' in output")
	}
}

func TestStartWithoutGitRepo(t *testing.T) {
	// Create temporary directory WITHOUT git
	tmpDir, err := os.MkdirTemp("", "checkpoint-start-nogit")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Capture stdout and stderr
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	success := startInternal(tmpDir)

	w.Close()
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	// Read output
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify error handling
	if success {
		t.Errorf("expected start to fail without git repository")
	}
	if !strings.Contains(output, "✗ Not a git repository") {
		t.Errorf("expected git repository error in output")
	}
	if !strings.Contains(output, "Cannot start") {
		t.Errorf("expected failure message in output")
	}
}

func TestStartWithoutCheckpointInit(t *testing.T) {
	// Create temporary directory with git but no checkpoint
	tmpDir, err := os.MkdirTemp("", "checkpoint-start-noinit")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repository
	setupGitRepo(t, tmpDir)

	// Capture output
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	success := startInternal(tmpDir)

	w.Close()
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if success {
		t.Errorf("expected start to fail without checkpoint initialization")
	}
	if !strings.Contains(output, "✗ Checkpoint not initialized") {
		t.Errorf("expected checkpoint not initialized error")
	}
	if !strings.Contains(output, "checkpoint init") {
		t.Errorf("expected hint to run checkpoint init")
	}
}

func TestStartWithCheckpointInProgress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-start-inprogress")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	setupGitRepo(t, tmpDir)

	// Create changelog
	changelogPath := filepath.Join(tmpDir, config.ChangelogFileName)
	if err := file.WriteFile(changelogPath, "---\nschema_version: \"1\"\n"); err != nil {
		t.Fatalf("failed to create changelog: %v", err)
	}

	// Create lock file to simulate checkpoint in progress
	lockPath := filepath.Join(tmpDir, config.LockFileName)
	if err := file.WriteFile(lockPath, "pid=123\n"); err != nil {
		t.Fatalf("failed to create lock file: %v", err)
	}

	// Capture output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	success := startInternal(tmpDir)

	w.Close()
	os.Stdout = originalStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !success {
		t.Errorf("expected start to succeed")
	}

	if !strings.Contains(output, "⚠ Checkpoint in progress") {
		t.Errorf("expected checkpoint in progress warning")
	}
	if !strings.Contains(output, "checkpoint commit") || !strings.Contains(output, "checkpoint clean") {
		t.Errorf("expected hints for resolving in-progress checkpoint")
	}
}

func TestStartWithNextSteps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-start-nextsteps")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	setupGitRepo(t, tmpDir)

	// Create changelog
	changelogPath := filepath.Join(tmpDir, config.ChangelogFileName)
	if err := file.WriteFile(changelogPath, "---\nschema_version: \"1\"\n"); err != nil {
		t.Fatalf("failed to create changelog: %v", err)
	}

	// Create status file with next steps
	statusPath := filepath.Join(tmpDir, config.StatusFileName)
	statusContent := `last_commit_hash: "abc123"
last_commit_timestamp: "2025-01-01T10:00:00Z"
status: "success"
next_steps:
  - summary: "High priority task"
    priority: "high"
    scope: "test"
  - summary: "Medium priority task"
    priority: "med"
  - summary: "Low priority task"
    priority: "low"
`
	if err := file.WriteFile(statusPath, statusContent); err != nil {
		t.Fatalf("failed to create status file: %v", err)
	}

	// Capture output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	success := startInternal(tmpDir)

	w.Close()
	os.Stdout = originalStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !success {
		t.Errorf("expected start to succeed with checkpoint in progress (warning only)")
	}

	if !strings.Contains(output, "NEXT STEPS") {
		t.Errorf("expected NEXT STEPS section in output")
	}
	if !strings.Contains(output, "[HIGH]") {
		t.Errorf("expected high priority task")
	}
	if !strings.Contains(output, "High priority task") {
		t.Errorf("expected high priority task summary")
	}
	if !strings.Contains(output, "[MED]") {
		t.Errorf("expected medium priority task")
	}
	if !strings.Contains(output, "[LOW]") {
		t.Errorf("expected low priority task")
	}

	// Check that high priority comes before low priority
	highIdx := strings.Index(output, "[HIGH]")
	lowIdx := strings.Index(output, "[LOW]")
	if highIdx > lowIdx {
		t.Errorf("high priority tasks should appear before low priority tasks")
	}
}

func TestCountCheckpoints(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "single checkpoint",
			content: `---
schema_version: "1"
document_type: meta
---
schema_version: "1"
timestamp: "2025-01-01T10:00:00Z"
changes: []
next_steps: []
`,
			expected: 1,
		},
		{
			name: "multiple checkpoints",
			content: `---
schema_version: "1"
document_type: meta
---
schema_version: "1"
timestamp: "2025-01-01T10:00:00Z"
---
schema_version: "1"
timestamp: "2025-01-02T10:00:00Z"
---
schema_version: "1"
timestamp: "2025-01-03T10:00:00Z"
`,
			expected: 3,
		},
		{
			name: "only meta document",
			content: `---
schema_version: "1"
document_type: meta
`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := countCheckpoints(tt.content)
			if count != tt.expected {
				t.Errorf("countCheckpoints() = %d, want %d", count, tt.expected)
			}
		})
	}
}

func TestCountPendingRecommendations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-start-recommendations")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	projectPath := filepath.Join(tmpDir, config.ProjectFileName)

	// Test with no project file
	count := countPendingRecommendations(projectPath)
	if count != 0 {
		t.Errorf("expected 0 recommendations when file doesn't exist, got %d", count)
	}

	// Create project file with recommendations
	projectContent := `---
schema_version: "1"
project_name: "test"
---
schema_version: "1"
document_type: recommendations
timestamp: "2025-01-01T10:00:00Z"
---
schema_version: "1"
document_type: recommendations
timestamp: "2025-01-02T10:00:00Z"
`
	if err := file.WriteFile(projectPath, projectContent); err != nil {
		t.Fatalf("failed to create project file: %v", err)
	}

	count = countPendingRecommendations(projectPath)
	if count != 2 {
		t.Errorf("expected 2 recommendations, got %d", count)
	}
}

// Helper function for tests
func setupGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git for commits
	exec.Command("git", "-C", dir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", dir, "config", "user.name", "Test User").Run()
}
