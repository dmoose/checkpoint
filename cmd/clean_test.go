package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-llm/internal/file"
	"go-llm/pkg/config"
)

func TestCleanCommand(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "checkpoint-clean-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test clean when no files exist (should not error)
	Clean(tmpDir)

	// Create checkpoint files manually
	inputPath := filepath.Join(tmpDir, config.InputFileName)
	diffPath := filepath.Join(tmpDir, config.DiffFileName)
	lockPath := filepath.Join(tmpDir, config.LockFileName)

	testContent := "test content"
	if err := file.WriteFile(inputPath, testContent); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}
	if err := file.WriteFile(diffPath, testContent); err != nil {
		t.Fatalf("failed to create diff file: %v", err)
	}
	if err := file.WriteFile(lockPath, testContent); err != nil {
		t.Fatalf("failed to create lock file: %v", err)
	}

	// Verify files exist before clean
	if !file.Exists(inputPath) {
		t.Fatalf("input file should exist before clean")
	}
	if !file.Exists(diffPath) {
		t.Fatalf("diff file should exist before clean")
	}
	if !file.Exists(lockPath) {
		t.Fatalf("lock file should exist before clean")
	}

	// Run clean command
	Clean(tmpDir)

	// Verify files are removed
	if file.Exists(inputPath) {
		t.Errorf("input file should be removed after clean")
	}
	if file.Exists(diffPath) {
		t.Errorf("diff file should be removed after clean")
	}
	if file.Exists(lockPath) {
		t.Errorf("lock file should be removed after clean")
	}
}

func TestCleanCommandOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-clean-output")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some checkpoint files
	inputPath := filepath.Join(tmpDir, config.InputFileName)
	diffPath := filepath.Join(tmpDir, config.DiffFileName)

	if err := file.WriteFile(inputPath, "test"); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}
	if err := file.WriteFile(diffPath, "test"); err != nil {
		t.Fatalf("failed to create diff file: %v", err)
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Clean(tmpDir)

	w.Close()
	os.Stdout = originalStdout

	// Read output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify output contains success message
	if !strings.Contains(output, "✓") || !strings.Contains(output, "cleaned") {
		t.Errorf("expected success message in output, got: %s", output)
	}
}

func TestCleanPartialFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-clean-partial")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create only input file (not diff file)
	inputPath := filepath.Join(tmpDir, config.InputFileName)
	if err := file.WriteFile(inputPath, "test"); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	// Run clean - should handle missing diff file gracefully
	Clean(tmpDir)

	// Verify input file is still removed
	if file.Exists(inputPath) {
		t.Errorf("input file should be removed even when diff file doesn't exist")
	}
}

func TestCleanPreservesOtherFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-clean-preserve")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create checkpoint files
	inputPath := filepath.Join(tmpDir, config.InputFileName)
	diffPath := filepath.Join(tmpDir, config.DiffFileName)

	// Create non-checkpoint files that should be preserved
	changelogPath := filepath.Join(tmpDir, config.ChangelogFileName)
	statusPath := filepath.Join(tmpDir, config.StatusFileName)
	otherPath := filepath.Join(tmpDir, "other-file.txt")

	testContent := "test content"
	files := []string{inputPath, diffPath, changelogPath, statusPath, otherPath}
	for _, filePath := range files {
		if err := file.WriteFile(filePath, testContent); err != nil {
			t.Fatalf("failed to create file %s: %v", filePath, err)
		}
	}

	// Run clean
	Clean(tmpDir)

	// Verify only checkpoint temporary files are removed
	if file.Exists(inputPath) {
		t.Errorf("input file should be removed")
	}
	if file.Exists(diffPath) {
		t.Errorf("diff file should be removed")
	}

	// Verify other files are preserved
	if !file.Exists(changelogPath) {
		t.Errorf("changelog file should be preserved")
	}
	if !file.Exists(statusPath) {
		t.Errorf("status file should be preserved")
	}
	if !file.Exists(otherPath) {
		t.Errorf("other file should be preserved")
	}
}

func TestCleanReadOnlyFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checkpoint-clean-readonly")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create input file and make it read-only
	inputPath := filepath.Join(tmpDir, config.InputFileName)
	if err := file.WriteFile(inputPath, "test"); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	// Make file read-only
	if err := os.Chmod(inputPath, 0444); err != nil {
		t.Fatalf("failed to make file read-only: %v", err)
	}

	// Capture stderr to check for warnings
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Clean(tmpDir)

	w.Close()
	os.Stderr = originalStderr

	// Read stderr output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	stderr := string(buf[:n])

	// On most systems, os.Remove can still remove read-only files
	// So we'll just check that clean completed without error
	// The important thing is that clean handles permission errors gracefully
	if strings.Contains(stderr, "failed to remove") {
		t.Errorf("clean should handle read-only files gracefully, got error: %s", stderr)
	}

	// Clean up by making file writable again
	os.Chmod(inputPath, 0644)
}
