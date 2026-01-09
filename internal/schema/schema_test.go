package schema

import (
	"strings"
	"testing"
)

func TestGenerateInputTemplateContainsFields(t *testing.T) {
	status := "M main.go\n?? newfile.go"
	out := GenerateInputTemplate(status, ".checkpoint-diff", nil)
	checks := []string{
		"schema_version: \"1\"",
		"git_status: |",
		"diff_file: \".checkpoint-diff\"",
		"changes:",
		"- summary:",
		"change_type:",
	}
	for _, s := range checks {
		if !strings.Contains(out, s) {
			t.Fatalf("template missing expected substring: %q", s)
		}
	}
}

func TestParseAndValidateEntry(t *testing.T) {
	content := strings.TrimSpace(`
Schema_version: "1"
Timestamp: "2025-10-22T16:00:00Z"
Commit_hash: ""
Changes:
  - summary: "Added auth"
    change_type: "feature"
    scope: "auth"
`)
	// Use lowercase keys as the parser expects specific yaml names
	content = strings.ReplaceAll(content, "Schema_version", "schema_version")
	content = strings.ReplaceAll(content, "Timestamp", "timestamp")
	content = strings.ReplaceAll(content, "Commit_hash", "commit_hash")
	content = strings.ReplaceAll(content, "Changes", "changes")
	entry, err := ParseInputFile(content)
	if err != nil {
		t.Fatalf("ParseInputFile error: %v", err)
	}
	if err := ValidateEntry(entry); err != nil {
		t.Fatalf("ValidateEntry unexpected error: %v", err)
	}
}

func TestValidateEntryErrors(t *testing.T) {
	// Missing required fields
	bad := &CheckpointEntry{}
	if err := ValidateEntry(bad); err == nil {
		t.Fatalf("expected error for missing fields")
	}
	// Invalid change_type
	bad2 := &CheckpointEntry{SchemaVersion: "1", Timestamp: "2025-10-22T00:00:00Z", Changes: []Change{{Summary: "x", ChangeType: "weird"}}}
	if err := ValidateEntry(bad2); err == nil {
		t.Fatalf("expected error for invalid change_type")
	}
}

func TestLintEntry(t *testing.T) {
	tests := []struct {
		name       string
		entry      *CheckpointEntry
		wantIssues int
		wantMatch  string // substring to find in issues
	}{
		{
			name: "clean entry",
			entry: &CheckpointEntry{
				SchemaVersion: "1",
				Timestamp:     "2025-01-01T00:00:00Z",
				Changes: []Change{
					{Summary: "Add user authentication endpoint", ChangeType: "feature", Scope: "api"},
				},
			},
			wantIssues: 0,
		},
		{
			name: "placeholder in summary",
			entry: &CheckpointEntry{
				SchemaVersion: "1",
				Timestamp:     "2025-01-01T00:00:00Z",
				Changes: []Change{
					{Summary: "[FILL IN: what changed]", ChangeType: "feature"},
				},
			},
			wantIssues: 1,
			wantMatch:  "placeholder",
		},
		{
			name: "vague summary",
			entry: &CheckpointEntry{
				SchemaVersion: "1",
				Timestamp:     "2025-01-01T00:00:00Z",
				Changes: []Change{
					{Summary: "improve code", ChangeType: "refactor"},
				},
			},
			wantIssues: 1,
			wantMatch:  "vague",
		},
		{
			name: "multiple ands suggests split",
			entry: &CheckpointEntry{
				SchemaVersion: "1",
				Timestamp:     "2025-01-01T00:00:00Z",
				Changes: []Change{
					{Summary: "Add auth and update config and fix tests", ChangeType: "feature"},
				},
			},
			wantIssues: 1,
			wantMatch:  "splitting",
		},
		{
			name: "placeholder in next_steps",
			entry: &CheckpointEntry{
				SchemaVersion: "1",
				Timestamp:     "2025-01-01T00:00:00Z",
				Changes: []Change{
					{Summary: "Add feature", ChangeType: "feature"},
				},
				NextSteps: []NextStep{
					{Summary: "[FILL IN: next action]"},
				},
			},
			wantIssues: 1,
			wantMatch:  "next_steps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := LintEntry(tt.entry)
			if len(issues) != tt.wantIssues {
				t.Errorf("expected %d issues, got %d: %v", tt.wantIssues, len(issues), issues)
			}
			if tt.wantMatch != "" && len(issues) > 0 {
				found := false
				for _, issue := range issues {
					if strings.Contains(strings.ToLower(issue), tt.wantMatch) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected issue containing %q, got: %v", tt.wantMatch, issues)
				}
			}
		})
	}
}

func TestRenderChangelogDocument(t *testing.T) {
	entry := &CheckpointEntry{
		SchemaVersion: "1",
		Timestamp:     "2025-01-01T00:00:00Z",
		CommitHash:    "abc123",
		Changes: []Change{
			{Summary: "Add feature", ChangeType: "feature", Scope: "api"},
		},
		NextSteps: []NextStep{
			{Summary: "Write tests", Priority: "high"},
		},
	}

	output, err := RenderChangelogDocument(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should start with document separator
	if !strings.HasPrefix(output, "---\n") {
		t.Error("output should start with YAML document separator")
	}

	// Should contain key fields
	checks := []string{
		"schema_version:",
		"timestamp:",
		"commit_hash: abc123",
		"changes:",
		"summary: Add feature",
		"next_steps:",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing %q", check)
		}
	}

	// Should NOT contain ephemeral fields
	if strings.Contains(output, "git_status") {
		t.Error("output should not contain git_status")
	}
	if strings.Contains(output, "diff_file") {
		t.Error("output should not contain diff_file")
	}
}

func TestIsPlaceholder(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"[FILL IN: what changed]", true},
		{"[fill in: description]", true},
		{"[OPTIONAL: details]", true},
		{"Add user authentication", false},
		{"Fix bug in handler", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isPlaceholder(tt.input)
			if result != tt.expected {
				t.Errorf("isPlaceholder(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateEntrySummaryLength(t *testing.T) {
	// Summary exactly at limit should pass
	entry := &CheckpointEntry{
		SchemaVersion: "1",
		Timestamp:     "2025-01-01T00:00:00Z",
		Changes: []Change{
			{Summary: strings.Repeat("a", MaxSummaryLength), ChangeType: "feature"},
		},
	}
	if err := ValidateEntry(entry); err != nil {
		t.Errorf("summary at max length should pass: %v", err)
	}

	// Summary over limit should fail
	entry.Changes[0].Summary = strings.Repeat("a", MaxSummaryLength+1)
	if err := ValidateEntry(entry); err == nil {
		t.Error("summary over max length should fail")
	}
}

func TestParseNumStat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []FileChange
	}{
		{
			name:     "empty input",
			input:    "",
			expected: []FileChange{},
		},
		{
			name:  "single file change",
			input: "5\t2\tsrc/main.go",
			expected: []FileChange{
				{Path: "src/main.go", Additions: 5, Deletions: 2},
			},
		},
		{
			name: "multiple files",
			input: `10	3	README.md
2	1	src/handler.go
0	5	old/deprecated.go`,
			expected: []FileChange{
				{Path: "README.md", Additions: 10, Deletions: 3},
				{Path: "src/handler.go", Additions: 2, Deletions: 1},
				{Path: "old/deprecated.go", Additions: 0, Deletions: 5},
			},
		},
		{
			name:  "binary file",
			input: "-\t-\timage.png",
			expected: []FileChange{
				{Path: "image.png", Additions: 0, Deletions: 0},
			},
		},
		{
			name:  "filename with spaces",
			input: "3\t1\tmy file with spaces.txt",
			expected: []FileChange{
				{Path: "my file with spaces.txt", Additions: 3, Deletions: 1},
			},
		},
		{
			name: "mixed binary and text files",
			input: `15	8	src/main.go
-	-	assets/logo.png
4	0	docs/README.md`,
			expected: []FileChange{
				{Path: "src/main.go", Additions: 15, Deletions: 8},
				{Path: "assets/logo.png", Additions: 0, Deletions: 0},
				{Path: "docs/README.md", Additions: 4, Deletions: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseNumStat(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d files, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("missing file at index %d", i)
					continue
				}
				actual := result[i]
				if actual.Path != expected.Path {
					t.Errorf("file %d: expected path %q, got %q", i, expected.Path, actual.Path)
				}
				if actual.Additions != expected.Additions {
					t.Errorf("file %d: expected %d additions, got %d", i, expected.Additions, actual.Additions)
				}
				if actual.Deletions != expected.Deletions {
					t.Errorf("file %d: expected %d deletions, got %d", i, expected.Deletions, actual.Deletions)
				}
			}
		})
	}
}
