package checkpoint

import (
	"testing"

	"github.com/dmoose/checkpoint/internal/git"
	"github.com/dmoose/checkpoint/internal/schema"
)

func TestInferChangeType(t *testing.T) {
	tests := []struct {
		subject  string
		expected string
	}{
		{"feat: add user auth", "feature"},
		{"fix: null pointer in profile", "fix"},
		{"refactor: extract validation", "refactor"},
		{"docs: update README", "docs"},
		{"perf: optimize query", "perf"},
		{"Fix memory leak", "fix"},
		{"Add new feature", "feature"},
		{"Implement caching", "feature"},
		{"Create user model", "feature"},
		{"Refactor database layer", "refactor"},
		{"Extract helper functions", "refactor"},
		{"Rename variables", "refactor"},
		{"Move files to new directory", "refactor"},
		{"Update dependencies", "other"},
		{"Document API endpoints", "docs"},
		{"Fix bug in parser", "fix"},
		{"Patch security vulnerability", "fix"},
		{"random commit message", "other"},
		{"Checkpoint: feature (api) - Add endpoint", "feature"},
		{"Checkpoint: fix (core) - Handle nil", "fix"},
		{"Checkpoint: 3 changes - refactor(2), fix", "fix"},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			got := inferChangeType(tt.subject)
			if got != tt.expected {
				t.Errorf("inferChangeType(%q) = %q, want %q", tt.subject, got, tt.expected)
			}
		})
	}
}

func TestInferScope(t *testing.T) {
	tests := []struct {
		name     string
		files    []schema.FileChange
		expected string
	}{
		{
			name:     "single directory",
			files:    []schema.FileChange{{Path: "cmd/main.go"}, {Path: "cmd/root.go"}},
			expected: "cmd",
		},
		{
			name:     "nested directory",
			files:    []schema.FileChange{{Path: "internal/git/git.go"}, {Path: "internal/git/git_test.go"}},
			expected: "internal/git",
		},
		{
			name:     "mixed directories",
			files:    []schema.FileChange{{Path: "cmd/main.go"}, {Path: "internal/schema/schema.go"}, {Path: "internal/schema/schema_test.go"}},
			expected: "internal/schema",
		},
		{
			name:     "root files",
			files:    []schema.FileChange{{Path: "main.go"}, {Path: "go.mod"}},
			expected: "",
		},
		{
			name:     "empty",
			files:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferScope(tt.files)
			if got != tt.expected {
				t.Errorf("inferScope() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a very long string that needs truncation", 20, "this is a very lo..."},
		{"has\nnewlines\nin it", 30, "has newlines in it"},
		{"", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.expected)
			}
		})
	}
}

func TestBuildMechanicalEntry(t *testing.T) {
	commit := git.CommitInfo{
		Hash:      "abc123def456",
		Timestamp: "2025-01-01T10:00:00Z",
		Author:    "Test",
		Subject:   "feat: add user authentication",
		Body:      "Implemented JWT-based auth",
	}
	files := []schema.FileChange{
		{Path: "internal/auth/auth.go", Additions: 100, Deletions: 0},
		{Path: "internal/auth/auth_test.go", Additions: 50, Deletions: 0},
	}

	entry := buildMechanicalEntry(commit, files)

	if entry.SchemaVersion != "1" {
		t.Errorf("expected schema_version 1, got %s", entry.SchemaVersion)
	}
	if !entry.Import {
		t.Error("expected import=true")
	}
	if entry.CommitHash != "abc123def456" {
		t.Errorf("expected commit hash abc123def456, got %s", entry.CommitHash)
	}
	if len(entry.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(entry.Changes))
	}
	if entry.Changes[0].ChangeType != "feature" {
		t.Errorf("expected change_type feature, got %s", entry.Changes[0].ChangeType)
	}
	if entry.Changes[0].Scope != "internal/auth" {
		t.Errorf("expected scope internal/auth, got %s", entry.Changes[0].Scope)
	}
	if entry.Context.ProblemStatement == "" {
		t.Error("expected non-empty problem_statement")
	}
}

func TestGenerateImportInput(t *testing.T) {
	commit := git.CommitInfo{
		Hash:      "abc123def456",
		Timestamp: "2025-01-01T10:00:00Z",
		Author:    "Test User",
		Subject:   "Add user model",
	}
	files := []schema.FileChange{
		{Path: "models/user.go", Additions: 50, Deletions: 0},
	}

	input := generateImportInput(commit, files)

	// Verify key sections are present
	expectations := []string{
		"IMPORT MODE",
		"abc123de",
		"schema_version:",
		"commit_hash:",
		"import: true",
		"files_changed:",
		"models/user.go",
		"changes:",
		"Add user model",
		"change_type:",
		"context:",
		"problem_statement:",
		"key_insights:",
		"decisions_made:",
		"import-commit",
	}

	for _, exp := range expectations {
		if !strContains(input, exp) {
			t.Errorf("import input missing expected content: %q", exp)
		}
	}
}

func strContains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCountTotalLines(t *testing.T) {
	files := []schema.FileChange{
		{Additions: 10, Deletions: 5},
		{Additions: 20, Deletions: 3},
	}
	if got := countTotalLines(files); got != 38 {
		t.Errorf("countTotalLines() = %d, want 38", got)
	}
}

func TestGetExistingCommitHashes(t *testing.T) {
	// Test with non-existent file
	hashes := getExistingCommitHashes("/nonexistent/path")
	if len(hashes) != 0 {
		t.Errorf("expected empty map for non-existent file, got %d entries", len(hashes))
	}
}
