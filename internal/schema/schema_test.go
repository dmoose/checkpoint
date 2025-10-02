package schema

import (
	"strings"
	"testing"
)

func TestGenerateInputTemplateContainsFields(t *testing.T) {
	status := "M main.go\n?? newfile.go"
	out := GenerateInputTemplate(status, ".checkpoint-diff")
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