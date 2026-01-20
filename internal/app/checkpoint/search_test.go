package checkpoint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dmoose/checkpoint/pkg/config"
)

func TestSplitYAMLDocuments(t *testing.T) {
	content := `---
document_type: meta
schema_version: "1"
---
timestamp: "2025-01-01T00:00:00Z"
changes:
  - summary: "first change"
---
timestamp: "2025-01-02T00:00:00Z"
changes:
  - summary: "second change"
`
	docs := splitYAMLDocuments(content)
	if len(docs) != 3 {
		t.Errorf("expected 3 documents, got %d", len(docs))
	}
}

func TestSplitYAMLDocuments_Empty(t *testing.T) {
	docs := splitYAMLDocuments("")
	if len(docs) != 0 {
		t.Errorf("expected 0 documents for empty input, got %d", len(docs))
	}
}

func TestSplitYAMLDocuments_SingleDoc(t *testing.T) {
	content := `timestamp: "2025-01-01T00:00:00Z"
changes:
  - summary: "only change"
`
	docs := splitYAMLDocuments(content)
	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}
}

func TestMatchesSearch_MatchingQuery(t *testing.T) {
	m := map[string]interface{}{
		"summary":     "add user authentication",
		"change_type": "feature",
		"scope":       "auth",
	}
	opts := SearchOptions{Query: "authentication"}
	if !matchesSearch(m, opts) {
		t.Error("expected match for query 'authentication'")
	}
}

func TestMatchesSearch_NonMatchingQuery(t *testing.T) {
	m := map[string]interface{}{
		"summary":     "add user authentication",
		"change_type": "feature",
		"scope":       "auth",
	}
	opts := SearchOptions{Query: "database"}
	if matchesSearch(m, opts) {
		t.Error("expected no match for query 'database'")
	}
}

func TestMatchesSearch_ScopeFilter(t *testing.T) {
	m := map[string]interface{}{
		"summary":     "add user authentication",
		"change_type": "feature",
		"scope":       "auth",
	}

	// Matching scope
	opts := SearchOptions{Scope: "auth"}
	if !matchesSearch(m, opts) {
		t.Error("expected match for scope 'auth'")
	}

	// Non-matching scope
	opts = SearchOptions{Scope: "database"}
	if matchesSearch(m, opts) {
		t.Error("expected no match for scope 'database'")
	}
}

func TestMatchesSearch_ScopeFilterMissing(t *testing.T) {
	m := map[string]interface{}{
		"summary": "some change without scope",
	}
	opts := SearchOptions{Scope: "auth"}
	if matchesSearch(m, opts) {
		t.Error("expected no match when scope field is missing")
	}
}

func TestMatchesSearch_NoQueryNoScope(t *testing.T) {
	m := map[string]interface{}{
		"summary": "anything",
	}
	opts := SearchOptions{}
	if !matchesSearch(m, opts) {
		t.Error("expected match when no query or scope filter")
	}
}

func TestMatchesQuery_StringValue(t *testing.T) {
	if !matchesQuery("hello world", "world") {
		t.Error("expected match for string containing query")
	}
	if matchesQuery("hello world", "missing") {
		t.Error("expected no match for string not containing query")
	}
}

func TestMatchesQuery_MapValue(t *testing.T) {
	m := map[string]interface{}{
		"description": "implement caching layer",
		"scope":       "performance",
	}
	if !matchesQuery(m, "caching") {
		t.Error("expected match for map value containing query")
	}
	if matchesQuery(m, "authentication") {
		t.Error("expected no match for map value not containing query")
	}
}

func TestMatchesQuery_EmptyQuery(t *testing.T) {
	if !matchesQuery("anything", "") {
		t.Error("expected match when query is empty")
	}
}

func TestMatchesQueryString_ExactMatch(t *testing.T) {
	if !matchesQueryString("hello world", "hello world") {
		t.Error("expected exact match")
	}
}

func TestMatchesQueryString_CaseInsensitive(t *testing.T) {
	if !matchesQueryString("Hello World", "hello world") {
		t.Error("expected case-insensitive match")
	}
	if !matchesQueryString("hello world", "HELLO") {
		t.Error("expected case-insensitive partial match")
	}
}

func TestMatchesQueryString_RegexPattern(t *testing.T) {
	if !matchesQueryString("implement caching layer", "cach.*layer") {
		t.Error("expected regex match")
	}
	if !matchesQueryString("error handling in auth module", "auth.*module") {
		t.Error("expected regex match for auth.*module")
	}
	if matchesQueryString("simple text", "complex.*pattern") {
		t.Error("expected no regex match")
	}
}

func TestMatchesQueryString_EmptyQuery(t *testing.T) {
	if matchesQueryString("anything", "") {
		t.Error("expected no match for empty query")
	}
}

func TestFormatChangeContent(t *testing.T) {
	m := map[string]interface{}{
		"summary":     "add login endpoint",
		"details":     "implemented JWT-based auth",
		"change_type": "feature",
		"scope":       "auth",
	}
	result := formatChangeContent(m)

	tests := []struct {
		expected string
	}{
		{"Summary: add login endpoint"},
		{"Details: implemented JWT-based auth"},
		{"Type: feature"},
		{"Scope: auth"},
	}
	for _, tc := range tests {
		if !contains(result, tc.expected) {
			t.Errorf("expected output to contain %q, got %q", tc.expected, result)
		}
	}
}

func TestFormatChangeContent_PartialFields(t *testing.T) {
	m := map[string]interface{}{
		"summary": "quick fix",
	}
	result := formatChangeContent(m)
	if !contains(result, "Summary: quick fix") {
		t.Errorf("expected summary in output, got %q", result)
	}
	if contains(result, "Details:") {
		t.Errorf("expected no details line, got %q", result)
	}
}

func TestFormatStepContent(t *testing.T) {
	m := map[string]interface{}{
		"summary":  "refactor database layer",
		"priority": "high",
		"scope":    "database",
	}
	result := formatStepContent(m)

	if !contains(result, "Summary: refactor database layer") {
		t.Errorf("expected summary, got %q", result)
	}
	if !contains(result, "Priority: high") {
		t.Errorf("expected priority, got %q", result)
	}
	if !contains(result, "Scope: database") {
		t.Errorf("expected scope, got %q", result)
	}
}

func TestFormatContextItem_StringItem(t *testing.T) {
	result := formatContextItem("insight", "use connection pooling for better performance")
	if result != "use connection pooling for better performance" {
		t.Errorf("expected string passthrough, got %q", result)
	}
}

func TestFormatContextItem_MapWithInsight(t *testing.T) {
	m := map[string]interface{}{
		"insight": "caching reduces latency by 50%",
		"scope":   "performance",
	}
	result := formatContextItem("insight", m)
	if !contains(result, "caching reduces latency by 50%") {
		t.Errorf("expected insight text, got %q", result)
	}
	if !contains(result, "Scope: performance") {
		t.Errorf("expected scope, got %q", result)
	}
}

func TestFormatContextItem_MapWithPattern(t *testing.T) {
	m := map[string]interface{}{
		"pattern":   "use table-driven tests",
		"rationale": "better test coverage",
	}
	result := formatContextItem("pattern", m)
	if !contains(result, "use table-driven tests") {
		t.Errorf("expected pattern text, got %q", result)
	}
	if !contains(result, "Rationale: better test coverage") {
		t.Errorf("expected rationale, got %q", result)
	}
}

func TestFormatContextItem_MapWithDecision(t *testing.T) {
	m := map[string]interface{}{
		"decision":  "use PostgreSQL over MySQL",
		"rationale": "better JSON support",
	}
	result := formatContextItem("decision", m)
	if !contains(result, "use PostgreSQL over MySQL") {
		t.Errorf("expected decision text, got %q", result)
	}
}

func TestFormatContextItem_MapWithFailedApproach(t *testing.T) {
	m := map[string]interface{}{
		"approach":   "tried using global state",
		"why_failed": "caused race conditions",
	}
	result := formatContextItem("failed_approach", m)
	if !contains(result, "tried using global state") {
		t.Errorf("expected approach text, got %q", result)
	}
	if !contains(result, "Why failed: caused race conditions") {
		t.Errorf("expected why_failed, got %q", result)
	}
}

func TestSearchChangelog(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, config.ChangelogFileName)

	content := `---
document_type: meta
schema_version: "1"
---
timestamp: "2025-01-01T00:00:00Z"
commit_hash: "abc123"
changes:
  - summary: "add user authentication"
    change_type: "feature"
    scope: "auth"
next_steps:
  - summary: "add password reset"
    priority: "high"
---
timestamp: "2025-01-02T00:00:00Z"
commit_hash: "def456"
changes:
  - summary: "fix database connection pooling"
    change_type: "fix"
    scope: "database"
`
	if err := os.WriteFile(changelogPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Search for "authentication"
	opts := SearchOptions{Query: "authentication"}
	results, err := searchChangelog(changelogPath, opts)
	if err != nil {
		t.Fatalf("searchChangelog failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result for 'authentication'")
	}
	if results[0].Source != "changelog" {
		t.Errorf("expected source 'changelog', got %q", results[0].Source)
	}
	if results[0].CommitHash != "abc123" {
		t.Errorf("expected commit_hash 'abc123', got %q", results[0].CommitHash)
	}

	// Search for "database" should match second doc
	opts = SearchOptions{Query: "database"}
	results, err = searchChangelog(changelogPath, opts)
	if err != nil {
		t.Fatalf("searchChangelog failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result for 'database'")
	}

	// Search for non-existent term
	opts = SearchOptions{Query: "nonexistent"}
	results, err = searchChangelog(changelogPath, opts)
	if err != nil {
		t.Fatalf("searchChangelog failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestSearchContext(t *testing.T) {
	tmpDir := t.TempDir()
	contextPath := filepath.Join(tmpDir, config.ContextFileName)

	content := `---
timestamp: "2025-01-01T00:00:00Z"
commit_hash: "abc123"
context:
  failed_approaches:
    - approach: "tried using global mutex"
      why_failed: "caused deadlocks under load"
  decisions_made:
    - decision: "use channel-based synchronization"
      rationale: "better composability and deadlock avoidance"
  key_insights:
    - insight: "goroutine leaks are hard to debug"
`
	if err := os.WriteFile(contextPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Search with --failed flag
	opts := SearchOptions{Failed: true}
	results, err := searchContext(contextPath, opts)
	if err != nil {
		t.Fatalf("searchContext failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result with --failed flag")
	}
	if results[0].Field != "failed_approaches" {
		t.Errorf("expected field 'failed_approaches', got %q", results[0].Field)
	}

	// Search with --decision flag
	opts = SearchOptions{Decision: true}
	results, err = searchContext(contextPath, opts)
	if err != nil {
		t.Fatalf("searchContext failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result with --decision flag")
	}
	found := false
	for _, r := range results {
		if r.Field == "decisions_made" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a result with field 'decisions_made'")
	}

	// Search with query for insights
	opts = SearchOptions{Query: "goroutine"}
	results, err = searchContext(contextPath, opts)
	if err != nil {
		t.Fatalf("searchContext failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result for query 'goroutine'")
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s, substr))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
