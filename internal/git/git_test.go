package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsGitRepository(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test non-git directory - should return false with no error
	ok, err := IsGitRepository(tmpDir)
	if ok {
		t.Errorf("expected false for non-git directory")
	}
	if err != nil {
		t.Errorf("unexpected error for non-git directory: %v", err)
	}

	// Test non-existent directory - should return error
	nonExistentDir := filepath.Join(tmpDir, "does-not-exist")
	ok, err = IsGitRepository(nonExistentDir)
	if ok {
		t.Errorf("expected false for non-existent directory")
	}
	if err == nil {
		t.Errorf("expected error for non-existent directory")
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available, skipping test: %v", err)
	}

	// Test git directory
	ok, err = IsGitRepository(tmpDir)
	if !ok {
		t.Errorf("expected true for git directory")
	}
	if err != nil {
		t.Errorf("unexpected error for git directory: %v", err)
	}

	// Test subdirectory of git repo
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	ok, err = IsGitRepository(subDir)
	if !ok {
		t.Errorf("expected true for subdirectory of git repo")
	}
	if err != nil {
		t.Errorf("unexpected error for subdirectory: %v", err)
	}
}

func TestGetDiff(t *testing.T) {
	// Create temporary git repository
	tmpDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Create a file and add some content
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("initial content\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Add and commit the file
	runGitCmd(t, tmpDir, "add", "test.txt")
	runGitCmd(t, tmpDir, "commit", "-m", "Initial commit")

	// Modify the file
	err = os.WriteFile(testFile, []byte("initial content\nmodified content\n"), 0644)
	if err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	// Test unstaged diff
	diff, err := GetDiff(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to get unstaged diff: %v", err)
	}
	if !strings.Contains(diff, "modified content") {
		t.Errorf("expected diff to contain 'modified content', got: %s", diff)
	}

	// Stage the file
	runGitCmd(t, tmpDir, "add", "test.txt")

	// Test staged diff
	diff, err = GetDiff(tmpDir, true)
	if err != nil {
		t.Fatalf("failed to get staged diff: %v", err)
	}
	if !strings.Contains(diff, "modified content") {
		t.Errorf("expected staged diff to contain 'modified content', got: %s", diff)
	}
}

func TestStageFile(t *testing.T) {
	tmpDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Create a test file
	testFile := "test.txt"
	testPath := filepath.Join(tmpDir, testFile)
	err := os.WriteFile(testPath, []byte("test content\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Stage the file
	err = StageFile(tmpDir, testFile)
	if err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}

	// Verify file is staged
	output := runGitCmd(t, tmpDir, "status", "--porcelain")
	if !strings.Contains(output, "A  test.txt") {
		t.Errorf("expected file to be staged, git status output: %s", output)
	}
}

func TestStageAll(t *testing.T) {
	tmpDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Create multiple test files
	files := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	for _, file := range files {
		filePath := filepath.Join(tmpDir, file)
		os.MkdirAll(filepath.Dir(filePath), 0755)
		err := os.WriteFile(filePath, []byte("test content\n"), 0644)
		if err != nil {
			t.Fatalf("failed to write test file %s: %v", file, err)
		}
	}

	// Stage all files
	err := StageAll(tmpDir)
	if err != nil {
		t.Fatalf("failed to stage all files: %v", err)
	}

	// Verify all files are staged
	output := runGitCmd(t, tmpDir, "status", "--porcelain")
	for _, file := range files {
		if !strings.Contains(output, file) {
			t.Errorf("expected %s to be staged, git status output: %s", file, output)
		}
	}
}

func TestCommit(t *testing.T) {
	tmpDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Create and stage a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	runGitCmd(t, tmpDir, "add", "test.txt")

	// Test commit
	commitMsg := "Test commit message"
	hash, err := Commit(tmpDir, commitMsg)
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	if hash == "" {
		t.Errorf("expected non-empty commit hash")
	}

	if len(hash) != 40 { // Standard Git SHA-1 hash length
		t.Errorf("expected 40-character hash, got %d characters: %s", len(hash), hash)
	}

	// Verify commit message
	output := runGitCmd(t, tmpDir, "log", "-1", "--pretty=format:%s")
	if strings.TrimSpace(output) != commitMsg {
		t.Errorf("expected commit message '%s', got '%s'", commitMsg, output)
	}
}

func TestGetStatus(t *testing.T) {
	tmpDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Empty repository should have empty status
	status, err := GetStatus(tmpDir)
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	if status != "" {
		t.Errorf("expected empty status for empty repo, got: %s", status)
	}

	// Create some files in different states
	// Untracked file
	untracked := filepath.Join(tmpDir, "untracked.txt")
	os.WriteFile(untracked, []byte("untracked\n"), 0644)

	// Modified file (first commit it, then modify)
	modified := filepath.Join(tmpDir, "modified.txt")
	os.WriteFile(modified, []byte("original\n"), 0644)
	runGitCmd(t, tmpDir, "add", "modified.txt")
	runGitCmd(t, tmpDir, "commit", "-m", "Add modified.txt")
	os.WriteFile(modified, []byte("original\nmodified\n"), 0644)

	// Staged file
	staged := filepath.Join(tmpDir, "staged.txt")
	os.WriteFile(staged, []byte("staged\n"), 0644)
	runGitCmd(t, tmpDir, "add", "staged.txt")

	// Get status
	status, err = GetStatus(tmpDir)
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	// Verify status contains expected entries
	expectedEntries := []string{
		"?? untracked.txt", // Untracked
		" M modified.txt",  // Modified
		"A  staged.txt",    // Added/Staged
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(status, entry) {
			t.Errorf("expected status to contain '%s', got: %s", entry, status)
		}
	}
}

// Helper function to set up a git repository for testing
func setupGitRepo(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Skipf("git not available, skipping test: %v", err)
	}

	// Set git config for testing
	runGitCmd(t, tmpDir, "config", "user.email", "test@example.com")
	runGitCmd(t, tmpDir, "config", "user.name", "Test User")

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// Helper function to run git commands for testing
func runGitCmd(t *testing.T, dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git command failed: %v", err)
	}
	return string(output)
}
