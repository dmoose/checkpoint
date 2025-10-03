package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-llm/internal/schema"
)

func TestUpdateLastDocumentBackfillsHash(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".checkpoint-changelog.yaml")
	// two docs
	doc1 := `---
 schema_version: "1"
 timestamp: "2025-10-22T00:00:00Z"
 changes:
   - summary: "first"
     change_type: "feature"
`
	doc2 := `---
 schema_version: "1"
 timestamp: "2025-10-22T01:00:00Z"
 changes:
   - summary: "second"
     change_type: "fix"
`
	if err := os.WriteFile(p, []byte(doc1+doc2), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// backfill
	if err := UpdateLastDocument(p, func(e *schema.CheckpointEntry) *schema.CheckpointEntry {
		e.CommitHash = "deadbeef"
		return e
	}); err != nil {
		t.Fatalf("UpdateLastDocument: %v", err)
	}
	b, err := os.ReadFile(p)
	if err != nil { t.Fatalf("read: %v", err) }
	if got := string(b); !containsAll(got, []string{"deadbeef", "first", "second"}) {
		t.Fatalf("unexpected content: %s", got)
	}
}

func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) { return false }
	}
	return true
}