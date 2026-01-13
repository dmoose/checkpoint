package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRecentContextEntries(t *testing.T) {
	// Create temp file with multiple YAML documents
	tmpDir := t.TempDir()
	contextPath := filepath.Join(tmpDir, "context.yml")

	content := `---
schema_version: "1"
timestamp: "2025-01-01T00:00:00Z"
context:
  problem_statement: "First entry"
---
schema_version: "1"
timestamp: "2025-01-02T00:00:00Z"
context:
  problem_statement: "Second entry"
---
schema_version: "1"
timestamp: "2025-01-03T00:00:00Z"
context:
  problem_statement: "Third entry"
---
schema_version: "1"
timestamp: "2025-01-04T00:00:00Z"
context:
  problem_statement: "Fourth entry"
`
	if err := os.WriteFile(contextPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name     string
		count    int
		expected int
		first    string // problem_statement of first returned entry
	}{
		{
			name:     "get last 2",
			count:    2,
			expected: 2,
			first:    "Third entry",
		},
		{
			name:     "get last 1",
			count:    1,
			expected: 1,
			first:    "Fourth entry",
		},
		{
			name:     "get more than exist",
			count:    10,
			expected: 4,
			first:    "First entry",
		},
		{
			name:     "get all",
			count:    4,
			expected: 4,
			first:    "First entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := GetRecentContextEntries(contextPath, tt.count)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(entries) != tt.expected {
				t.Errorf("expected %d entries, got %d", tt.expected, len(entries))
			}
			if len(entries) > 0 && entries[0].Context.ProblemStatement != tt.first {
				t.Errorf("expected first entry problem_statement %q, got %q",
					tt.first, entries[0].Context.ProblemStatement)
			}
		})
	}
}

func TestGetRecentContextEntriesNonexistent(t *testing.T) {
	entries, err := GetRecentContextEntries("/nonexistent/path/context.yml", 5)
	if err != nil {
		t.Fatalf("expected no error for nonexistent file, got: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty slice for nonexistent file, got %d entries", len(entries))
	}
}

func TestCreateContextEntry(t *testing.T) {
	ctx := CheckpointContext{
		ProblemStatement: "Test problem",
		KeyInsights: []Insight{
			{Insight: "Test insight", Scope: "project"},
		},
	}

	entry := CreateContextEntry("2025-01-01T00:00:00Z", ctx)

	if entry.SchemaVersion != "1" {
		t.Errorf("expected schema_version '1', got %q", entry.SchemaVersion)
	}
	if entry.Timestamp != "2025-01-01T00:00:00Z" {
		t.Errorf("expected timestamp '2025-01-01T00:00:00Z', got %q", entry.Timestamp)
	}
	if entry.Context.ProblemStatement != "Test problem" {
		t.Errorf("expected problem_statement 'Test problem', got %q", entry.Context.ProblemStatement)
	}
}
