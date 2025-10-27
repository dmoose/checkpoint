package schema

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"go-llm/internal/context"
	"go-llm/internal/language"
)

type FileChange struct {
	Path      string `yaml:"path"`
	Additions int    `yaml:"additions"`
	Deletions int    `yaml:"deletions"`
}

type Change struct {
	Summary    string `yaml:"summary"`
	Details    string `yaml:"details,omitempty"`
	ChangeType string `yaml:"change_type"`
	Scope      string `yaml:"scope,omitempty"`
}

type CheckpointEntry struct {
	SchemaVersion string                    `yaml:"schema_version"`
	Timestamp     string                    `yaml:"timestamp"`
	CommitHash    string                    `yaml:"commit_hash,omitempty"`
	GitStatus     string                    `yaml:"git_status,omitempty"`
	DiffFile      string                    `yaml:"diff_file,omitempty"`
	FilesChanged  []FileChange              `yaml:"files_changed,omitempty"`
	Context       context.CheckpointContext `yaml:"context,omitempty"`
	Changes       []Change                  `yaml:"changes"`
	NextSteps     []NextStep                `yaml:"next_steps,omitempty"`
}

const (
	SchemaVersion    = "1"
	MaxSummaryLength = 80
	LLMPrompt        = `# INSTRUCTIONS FOR LLM:
# 1. Fill the changes array with all changes in this checkpoint
# 2. Run 'checkpoint lint' to check your work
# 3. Human will review and edit before running 'checkpoint commit'
#
# Each change has: summary (required), details (optional), change_type (required), scope (optional).
# Allowed change_type values: feature, fix, refactor, docs, perf, other.
# Keep summaries concise (<80 chars), present tense; use consistent scope names.
# Derive distinct changes from git_status/diff context - group related file changes into logical units.
# If previous next_steps are present, remove completed items, keep unfinished ones, add new items as needed.
# Do not alter schema_version/timestamp; leave commit_hash empty.

# EXAMPLES OF GOOD CHANGES:
# - summary: "Add user authentication endpoint"
#   details: "Implemented JWT-based auth with login/logout routes and middleware"
#   change_type: "feature"
#   scope: "api"
#
# - summary: "Fix memory leak in connection pool"
#   details: "Connections were not being properly closed after timeout"
#   change_type: "fix"
#   scope: "database"
#
# - summary: "Refactor validation logic into separate module"
#   details: "Moved input validation from handlers to dedicated validator package"
#   change_type: "refactor"
#   scope: "internal/validator"
#
# - summary: "Update API documentation for v2 endpoints"
#   change_type: "docs"
#   scope: "api"

# GUIDELINES:
# - One change = one logical unit (feature/fix/refactor)
# - Group related file changes into single entries
# - Use present tense, active voice ("Add", "Fix", "Update")
# - Details should explain WHY, not just WHAT
# - Be specific in summaries - avoid vague words like "improve" or "update"
# - After filling, run 'checkpoint lint' to catch obvious mistakes
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
	return GenerateInputTemplateWithMetadata(gitStatus, diffFileName, prevNextSteps, nil, nil, "", "")
}

func GenerateInputTemplateWithMetadata(gitStatus, diffFileName string, prevNextSteps []NextStep, filesChanged []FileChange, languages []language.Language, projectContext, recentContext string) string {
	ts := time.Now().Format(time.RFC3339)
	prev := renderNextStepsYAML(prevNextSteps)

	// Build file changes section
	filesSection := ""
	if len(filesChanged) > 0 {
		filesSection = "\n# File changes (informational):\nfiles_changed:\n"
		for _, file := range filesChanged {
			filesSection += fmt.Sprintf("  - path: \"%s\"\n    additions: %d\n    deletions: %d\n",
				file.Path, file.Additions, file.Deletions)
		}
	}

	// Note: Language detection, project context, and recent context are no longer embedded
	// to keep the input file manageable. Reference files directly if needed:
	// - Project patterns: .checkpoint-project.yml
	// - Recent decisions: .checkpoint-context.yml
	// - Run 'checkpoint start' to see next steps and project summary

	// Get context template
	contextTemplate := context.GenerateContextTemplate()

	return fmt.Sprintf(`%s
schema_version: "%s"
timestamp: "%s"
commit_hash: ""

# Git status output (informational):
git_status: |
%s

# Reference to diff context (path to git diff output):
diff_file: "%s"%s

# REFERENCE FILES (if needed):
# - Project patterns and conventions: .checkpoint-project.yml
# - Recent checkpoint decisions: .checkpoint-context.yml
# - Run 'checkpoint start' to see project summary and next steps

# List all changes made in this checkpoint
changes:
  - summary: "[FILL IN: what changed]"
    details: "[OPTIONAL: longer description]"
    change_type: "[FILL IN: feature|fix|refactor|docs|perf|other]"
    scope: "[FILL IN: affected component]"
#  - summary: "[FILL IN: another change]"
#    change_type: "[FILL IN]"
#    scope: "[FILL IN]"
%s
# Planned next steps (optional)
# If previous next steps are present below, update by removing completed items and keeping unfinished ones.
next_steps:
%s
`, LLMPrompt, SchemaVersion, ts, indent(gitStatus), diffFileName, filesSection, contextTemplate, prev)
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

// LintEntry performs simple checks to catch obvious mistakes
func LintEntry(e *CheckpointEntry) []string {
	var issues []string

	// Check for placeholder text
	placeholderPatterns := []string{
		"[fill in", "[FILL IN", "[optional", "[OPTIONAL",
	}

	for i, c := range e.Changes {
		summary := strings.ToLower(c.Summary)
		details := strings.ToLower(c.Details)
		changeType := strings.ToLower(c.ChangeType)
		scope := strings.ToLower(c.Scope)

		// Check for placeholders
		for _, pattern := range placeholderPatterns {
			if strings.Contains(summary, pattern) {
				issues = append(issues, fmt.Sprintf("change[%d]: summary contains placeholder text", i))
				break
			}
		}
		for _, pattern := range placeholderPatterns {
			if strings.Contains(details, pattern) {
				issues = append(issues, fmt.Sprintf("change[%d]: details contains placeholder text", i))
				break
			}
		}
		for _, pattern := range placeholderPatterns {
			if strings.Contains(changeType, pattern) {
				issues = append(issues, fmt.Sprintf("change[%d]: change_type contains placeholder text", i))
				break
			}
		}
		for _, pattern := range placeholderPatterns {
			if strings.Contains(scope, pattern) {
				issues = append(issues, fmt.Sprintf("change[%d]: scope contains placeholder text", i))
				break
			}
		}

		// Check for vague summaries
		vagueWords := []string{"improve", "update", "enhance", "optimize", "various", "misc", "stuff"}
		for _, vague := range vagueWords {
			if strings.Contains(summary, vague) && len(strings.Fields(c.Summary)) < 5 {
				issues = append(issues, fmt.Sprintf("change[%d]: summary may be too vague (contains '%s')", i, vague))
				break
			}
		}

		// Check for overly long entries that might need splitting
		if strings.Count(c.Summary, " and ") > 1 {
			issues = append(issues, fmt.Sprintf("change[%d]: summary contains multiple 'and' - consider splitting into separate changes", i))
		}
	}

	// Check next_steps for placeholders
	for i, n := range e.NextSteps {
		summary := strings.ToLower(n.Summary)
		for _, pattern := range placeholderPatterns {
			if strings.Contains(summary, pattern) {
				issues = append(issues, fmt.Sprintf("next_steps[%d]: summary contains placeholder text", i))
				break
			}
		}
	}

	return issues
}

// RenderChangelogDocument renders only the persisted fields (omits git_status/diff_file)
func RenderChangelogDocument(e *CheckpointEntry) (string, error) {
	out := struct {
		SchemaVersion string       `yaml:"schema_version"`
		Timestamp     string       `yaml:"timestamp"`
		CommitHash    string       `yaml:"commit_hash"`
		FilesChanged  []FileChange `yaml:"files_changed,omitempty"`
		Changes       []Change     `yaml:"changes"`
		NextSteps     []NextStep   `yaml:"next_steps"`
	}{
		SchemaVersion: e.SchemaVersion,
		Timestamp:     e.Timestamp,
		CommitHash:    e.CommitHash,
		FilesChanged:  e.FilesChanged,
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

// ParseNumStat parses git diff --numstat output into FileChange structs
func ParseNumStat(numstat string) []FileChange {
	var files []FileChange
	lines := strings.Split(strings.TrimSpace(numstat), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		additions := 0
		deletions := 0

		// Handle binary files (marked with "-")
		if parts[0] != "-" {
			if add, err := strconv.Atoi(parts[0]); err == nil {
				additions = add
			}
		}
		if parts[1] != "-" {
			if del, err := strconv.Atoi(parts[1]); err == nil {
				deletions = del
			}
		}

		// Join remaining parts as filename (handles spaces in filenames)
		filename := strings.Join(parts[2:], " ")

		files = append(files, FileChange{
			Path:      filename,
			Additions: additions,
			Deletions: deletions,
		})
	}

	return files
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
