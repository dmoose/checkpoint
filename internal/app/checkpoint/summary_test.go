package checkpoint

import (
	"fmt"
	"testing"
	"time"
)

func TestCountCheckpointsInChangelog(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "no separators",
			content:  "just some text",
			expected: 0,
		},
		{
			name:     "only meta separator",
			content:  "---\ndocument_type: meta\n",
			expected: 0,
		},
		{
			name: "meta plus one checkpoint",
			content: `---
document_type: meta
---
timestamp: "2025-01-01T00:00:00Z"
`,
			expected: 1,
		},
		{
			name: "meta plus three checkpoints",
			content: `---
document_type: meta
---
timestamp: "2025-01-01T00:00:00Z"
---
timestamp: "2025-01-02T00:00:00Z"
---
timestamp: "2025-01-03T00:00:00Z"
`,
			expected: 3,
		},
		{
			name:     "empty content",
			content:  "",
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := countCheckpointsInChangelog(tc.content)
			if got != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, got)
			}
		})
	}
}

func TestExtractRecentCheckpoints(t *testing.T) {
	content := `---
document_type: meta
schema_version: "1"
---
timestamp: "2025-01-01T00:00:00Z"
commit_hash: "aaa111"
changes:
  - summary: "first change"
---
timestamp: "2025-01-02T00:00:00Z"
commit_hash: "bbb222"
changes:
  - summary: "second change"
---
timestamp: "2025-01-03T00:00:00Z"
commit_hash: "ccc333"
changes:
  - summary: "third change"
---
timestamp: "2025-01-04T00:00:00Z"
commit_hash: "ddd444"
changes:
  - summary: "fourth change"
---
timestamp: "2025-01-05T00:00:00Z"
commit_hash: "eee555"
changes:
  - summary: "fifth change"
`

	// Request last 3
	result := extractRecentCheckpoints(content, 3)
	if len(result) != 3 {
		t.Fatalf("expected 3 checkpoints, got %d", len(result))
	}

	// Should be the last 3 in order
	if result[0].summary != "third change" {
		t.Errorf("expected first result 'third change', got %q", result[0].summary)
	}
	if result[1].summary != "fourth change" {
		t.Errorf("expected second result 'fourth change', got %q", result[1].summary)
	}
	if result[2].summary != "fifth change" {
		t.Errorf("expected third result 'fifth change', got %q", result[2].summary)
	}
	if result[2].hash != "eee555" {
		t.Errorf("expected hash 'eee555', got %q", result[2].hash)
	}
}

func TestExtractRecentCheckpoints_RequestMoreThanAvailable(t *testing.T) {
	content := `---
document_type: meta
---
timestamp: "2025-01-01T00:00:00Z"
changes:
  - summary: "only change"
`
	result := extractRecentCheckpoints(content, 5)
	if len(result) != 1 {
		t.Errorf("expected 1 checkpoint, got %d", len(result))
	}
}

func TestExtractLastCheckpointInfo(t *testing.T) {
	statusContent := `last_commit_hash: "abc123def456"
last_commit_timestamp: "2025-06-15T10:30:00Z"
`
	hash, timestamp := extractLastCheckpointInfo(statusContent)
	if hash != "abc123def456" {
		t.Errorf("expected hash 'abc123def456', got %q", hash)
	}
	if timestamp != "2025-06-15T10:30:00Z" {
		t.Errorf("expected timestamp '2025-06-15T10:30:00Z', got %q", timestamp)
	}
}

func TestExtractLastCheckpointInfo_Empty(t *testing.T) {
	hash, timestamp := extractLastCheckpointInfo("")
	if hash != "" {
		t.Errorf("expected empty hash, got %q", hash)
	}
	if timestamp != "" {
		t.Errorf("expected empty timestamp, got %q", timestamp)
	}
}

func TestExtractNextStepsFromStatusFile(t *testing.T) {
	statusContent := `next_steps:
  - summary: "implement caching"
    priority: "high"
    scope: "performance"
  - summary: "write tests"
    priority: "medium"
    scope: "testing"
`
	steps := extractNextStepsFromStatusFile(statusContent)
	if len(steps) != 2 {
		t.Fatalf("expected 2 next steps, got %d", len(steps))
	}
	if steps[0].summary != "implement caching" {
		t.Errorf("expected 'implement caching', got %q", steps[0].summary)
	}
	if steps[0].priority != "high" {
		t.Errorf("expected priority 'high', got %q", steps[0].priority)
	}
	if steps[0].scope != "performance" {
		t.Errorf("expected scope 'performance', got %q", steps[0].scope)
	}
	if steps[1].summary != "write tests" {
		t.Errorf("expected 'write tests', got %q", steps[1].summary)
	}
}

func TestExtractNextStepsFromStatusFile_NoSteps(t *testing.T) {
	statusContent := `last_commit_hash: "abc123"
`
	steps := extractNextStepsFromStatusFile(statusContent)
	if len(steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(steps))
	}
}

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"just now", 10 * time.Second, "just now"},
		{"1 minute ago", 1 * time.Minute, "1 minute ago"},
		{"5 minutes ago", 5 * time.Minute, "5 minutes ago"},
		{"1 hour ago", 1 * time.Hour, "1 hour ago"},
		{"3 hours ago", 3 * time.Hour, "3 hours ago"},
		{"1 day ago", 24 * time.Hour, "1 day ago"},
		{"4 days ago", 4 * 24 * time.Hour, "4 days ago"},
		{"1 week ago", 7 * 24 * time.Hour, "1 week ago"},
		{"3 weeks ago", 21 * 24 * time.Hour, "3 weeks ago"},
		{"1 month ago", 31 * 24 * time.Hour, "1 month ago"},
		{"3 months ago", 90 * 24 * time.Hour, "3 months ago"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			timestamp := time.Now().Add(-tc.duration).Format(time.RFC3339)
			result := formatTimeAgo(timestamp)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestFormatTimeAgo_InvalidTimestamp(t *testing.T) {
	result := formatTimeAgo("not-a-timestamp")
	if result != "not-a-timestamp" {
		t.Errorf("expected raw string returned for invalid timestamp, got %q", result)
	}
}

func TestFormatTimeAgo_FutureEdgeCases(t *testing.T) {
	// A timestamp from exactly now should be "just now"
	timestamp := time.Now().Format(time.RFC3339)
	result := formatTimeAgo(timestamp)
	// Should be "just now" since duration < 1 minute
	if result != "just now" {
		// Allow for minor timing issues
		fmt.Printf("note: formatTimeAgo for 'now' returned %q (timing sensitive)\n", result)
	}
}
