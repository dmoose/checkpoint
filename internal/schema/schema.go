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
	SchemaVersion    = "1"
	MaxSummaryLength = 80
	LLMPrompt        = `# INSTRUCTIONS FOR LLM:
# This is a checkpoint input. Fill the changes array with all changes in this checkpoint.
# Each change has: summary (required), details (optional), change_type (required), scope (optional).
# Allowed change_type values: feature, fix, refactor, docs, perf, other.
# Keep summaries concise (<80 chars), present tense; use consistent scope names.
# Derive distinct changes from git_status/diff context where possible.
# If previous next_steps are present, remove items completed in this checkpoint, keep unfinished ones, and add new items as needed.
# Do not alter schema_version/timestamp; leave commit_hash empty.
`
	ValidChangeTypes = "feature, fix, refactor, docs, perf, other"
)

type NextStep struct {
	Summary  string `yaml:"summary"`
	Details  string `yaml:"details,omitempty"`
	Priority string `yaml:"priority,omitempty"` // low|med|high
	Scope    string `yaml:"scope,omitempty"`
}

func GenerateInputTemplate(gitStatus, diffFileName string, prevNextSteps []NextStep) string {
	ts := time.Now().Format(time.RFC3339)
	prev := renderNextStepsYAML(prevNextSteps)
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
# If previous next steps are present below, update by removing completed items and keeping unfinished ones.
next_steps:
%s
`, LLMPrompt, SchemaVersion, ts, indent(gitStatus), diffFileName, prev)
}

// ExtractNextStepsFromStatus parses a status YAML and returns next_steps if present
func ExtractNextStepsFromStatus(statusYAML string) []NextStep {
	var aux struct {
		NextSteps []NextStep `yaml:"next_steps"`
	}
	_ = yaml.Unmarshal([]byte(statusYAML), &aux)
	return aux.NextSteps
}

func renderNextStepsYAML(steps []NextStep) string {
	if len(steps) == 0 {
		return "#  - summary: \"[FILL IN: next action]\"\n#    details: \"[OPTIONAL: context]\"\n#    priority: \"[OPTIONAL: low|med|high]\"\n#    scope: \"[OPTIONAL: affected component]\"\n"
	}
	var b strings.Builder
	for _, s := range steps {
		b.WriteString("  - summary: \"")
		b.WriteString(strings.ReplaceAll(s.Summary, "\"", "'"))
		b.WriteString("\"\n")
		if s.Details != "" {
			b.WriteString("    details: \"" + strings.ReplaceAll(s.Details, "\"", "'") + "\"\n")
		}
		if s.Priority != "" {
			b.WriteString("    priority: \"" + s.Priority + "\"\n")
		}
		if s.Scope != "" {
			b.WriteString("    scope: \"" + s.Scope + "\"\n")
		}
	}
	return b.String()
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
	if e.SchemaVersion == "" {
		missing = append(missing, "schema_version")
	}
	if e.Timestamp == "" {
		missing = append(missing, "timestamp")
	}
	if len(e.Changes) == 0 {
		missing = append(missing, "changes[>=1]")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	valid := map[string]struct{}{"feature": {}, "fix": {}, "refactor": {}, "docs": {}, "perf": {}, "other": {}}
	for i, c := range e.Changes {
		summary := strings.TrimSpace(c.Summary)
		if summary == "" {
			return fmt.Errorf("change[%d]: summary required", i)
		}
		if isPlaceholder(summary) {
			return fmt.Errorf("change[%d]: summary contains placeholder text", i)
		}
		if _, ok := valid[c.ChangeType]; !ok {
			return fmt.Errorf("change[%d]: invalid change_type '%s' (valid: %s)", i, c.ChangeType, ValidChangeTypes)
		}
		if isPlaceholder(c.ChangeType) {
			return fmt.Errorf("change[%d]: change_type contains placeholder text", i)
		}
		if len([]rune(summary)) > MaxSummaryLength {
			return fmt.Errorf("change[%d]: summary too long (%d > %d chars)", i, len([]rune(summary)), MaxSummaryLength)
		}
	}

	// Validate next_steps
	for i, n := range e.NextSteps {
		summary := strings.TrimSpace(n.Summary)
		if summary == "" {
			return fmt.Errorf("next_steps[%d]: summary required", i)
		}
		if isPlaceholder(summary) {
			return fmt.Errorf("next_steps[%d]: summary contains placeholder text", i)
		}
		if n.Priority != "" {
			p := strings.ToLower(n.Priority)
			if p != "low" && p != "med" && p != "high" {
				return fmt.Errorf("next_steps[%d]: priority must be low|med|high (got: %s)", i, n.Priority)
			}
		}
	}

	return nil
}

// isPlaceholder detects placeholder text in input fields
func isPlaceholder(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return strings.HasPrefix(s, "[fill in") || strings.Contains(s, "[fill in") || strings.HasPrefix(s, "[optional")
}

// PreCommitValidate performs additional scope and consistency checks
func PreCommitValidate(e *CheckpointEntry) []string {
	var errs []string

	// Additional scope consistency checks could go here
	// For now, this is mainly for future extensibility

	return errs
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
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if ln == "" {
			continue
		}
		// Normalize leading whitespace to avoid YAML block issues
		trimmed := strings.TrimLeft(ln, " \t")
		lines[i] = "  " + trimmed
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
