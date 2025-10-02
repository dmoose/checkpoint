package schema

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Change struct {
	Summary    string `yaml:"summary"`
	Details    string `yaml:"details,omitempty"`
	ChangeType string `yaml:"change_type"`
	Scope      string `yaml:"scope,omitempty"`
}

type CheckpointEntry struct {
	SchemaVersion string     `yaml:"schema_version"`
	Timestamp     string     `yaml:"timestamp"`
	CommitHash    string     `yaml:"commit_hash,omitempty"`
	GitStatus     string     `yaml:"git_status,omitempty"`
	DiffFile      string     `yaml:"diff_file,omitempty"`
	Changes       []Change   `yaml:"changes"`
	NextSteps     []NextStep `yaml:"next_steps,omitempty"`
}

const (
	SchemaVersion = "1"
	LLMPrompt     = `# INSTRUCTIONS FOR LLM:
# This is a checkpoint input. Fill the changes array with all changes in this checkpoint.
# Each change has: summary (required), details (optional), change_type (required), scope (optional).
# Allowed change_type values: feature, fix, refactor, docs, perf, other.
# Keep summaries concise (<80 chars), present tense; use consistent scope names.
# Derive distinct changes from git_status/diff context where possible.
# Optionally propose next_steps (planned work) below; do not alter schema_version/timestamp; leave commit_hash empty.
`
)

type NextStep struct {
	Summary  string `yaml:"summary"`
	Details  string `yaml:"details,omitempty"`
	Priority string `yaml:"priority,omitempty"` // low|med|high
	Scope    string `yaml:"scope,omitempty"`
	Owner    string `yaml:"owner,omitempty"`
}

func GenerateInputTemplate(gitStatus, diffFileName string) string {
	ts := time.Now().Format(time.RFC3339)
	return fmt.Sprintf(`%s
schema_version: "%s"
timestamp: "%s"
commit_hash: ""

# Git status output (informational):
git_status: |
%s

# Reference to diff context (path to git diff output):
diff_file: "%s"

# List all changes made in this checkpoint
changes:
  - summary: "[FILL IN: what changed]"
    details: "[OPTIONAL: longer description]"
    change_type: "[FILL IN: feature|fix|refactor|docs|perf|other]"
    scope: "[FILL IN: affected component]"
#  - summary: "[FILL IN: another change]"
#    change_type: "[FILL IN]"
#    scope: "[FILL IN]"

# Planned next steps (optional)
next_steps:
#  - summary: "[FILL IN: next action]"
#    details: "[OPTIONAL: context]"
#    priority: "[OPTIONAL: low|med|high]"
#    scope: "[OPTIONAL: affected component]"
#    owner: "[OPTIONAL: who]"
`, LLMPrompt, SchemaVersion, ts, indent(gitStatus), diffFileName)
}

func ParseInputFile(content string) (*CheckpointEntry, error) {
	trimmed := stripPrompt(content)
	var e CheckpointEntry
	if err := yaml.Unmarshal([]byte(trimmed), &e); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &e, nil
}

func ValidateEntry(e *CheckpointEntry) error {
	var missing []string
	if e.SchemaVersion == "" { missing = append(missing, "schema_version") }
	if e.Timestamp == "" { missing = append(missing, "timestamp") }
	if len(e.Changes) == 0 { missing = append(missing, "changes[>=1]") }
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	valid := map[string]struct{}{ "feature":{}, "fix":{}, "refactor":{}, "docs":{}, "perf":{}, "other":{} }
	for i, c := range e.Changes {
		if strings.TrimSpace(c.Summary) == "" {
			return fmt.Errorf("change[%d]: summary required", i)
		}
		if _, ok := valid[c.ChangeType]; !ok {
			return fmt.Errorf("change[%d]: invalid change_type '%s'", i, c.ChangeType)
		}
	}
	return nil
}

// RenderChangelogDocument renders only the persisted fields (omits git_status/diff_file)
func RenderChangelogDocument(e *CheckpointEntry) (string, error) {
	out := struct {
		SchemaVersion string     `yaml:"schema_version"`
		Timestamp     string     `yaml:"timestamp"`
		CommitHash    string     `yaml:"commit_hash,omitempty"`
		Changes       []Change   `yaml:"changes"`
		NextSteps     []NextStep `yaml:"next_steps,omitempty"`
	}{
		SchemaVersion: e.SchemaVersion,
		Timestamp:     e.Timestamp,
		CommitHash:    e.CommitHash,
		Changes:       e.Changes,
		NextSteps:     e.NextSteps,
	}
	b, err := yaml.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("marshal yaml: %w", err)
	}
	return "---\n" + string(b), nil
}

func indent(s string) string {
	if s == "" { return s }
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if ln != "" {
			lines[i] = "  " + ln
		}
	}
	return strings.Join(lines, "\n")
}

func stripPrompt(content string) string {
	lines := strings.Split(content, "\n")
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "schema_version:") {
			return strings.Join(lines[i:], "\n")
		}
	}
	return content
}