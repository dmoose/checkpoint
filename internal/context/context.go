package context

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ContextEntry represents a single checkpoint's context
type ContextEntry struct {
	SchemaVersion string            `yaml:"schema_version"`
	Timestamp     string            `yaml:"timestamp"`
	CommitHash    string            `yaml:"commit_hash,omitempty"`
	Context       CheckpointContext `yaml:"context"`
}

// CheckpointContext represents the context captured at a checkpoint
type CheckpointContext struct {
	ProblemStatement    string             `yaml:"problem_statement"`
	KeyInsights         []Insight          `yaml:"key_insights,omitempty"`
	DecisionsMade       []Decision         `yaml:"decisions_made,omitempty"`
	FailedApproaches    []FailedApproach   `yaml:"failed_approaches,omitempty"`
	EstablishedPatterns []Pattern          `yaml:"established_patterns,omitempty"`
	ConversationContext []ConversationItem `yaml:"conversation_context,omitempty"`
}

type Insight struct {
	Insight string `yaml:"insight"`
	Impact  string `yaml:"impact,omitempty"`
	Scope   string `yaml:"scope,omitempty"` // checkpoint|project
}

type Decision struct {
	Decision                  string   `yaml:"decision"`
	Rationale                 string   `yaml:"rationale"`
	AlternativesConsidered    []string `yaml:"alternatives_considered,omitempty"`
	ConstraintsThatInfluenced string   `yaml:"constraints_that_influenced,omitempty"`
	Scope                     string   `yaml:"scope,omitempty"` // checkpoint|project
}

type FailedApproach struct {
	Approach       string `yaml:"approach"`
	WhyFailed      string `yaml:"why_failed,omitempty"`
	LessonsLearned string `yaml:"lessons_learned,omitempty"`
	Scope          string `yaml:"scope,omitempty"` // checkpoint|project
}

type Pattern struct {
	Pattern   string `yaml:"pattern"`
	Rationale string `yaml:"rationale,omitempty"`
	Examples  string `yaml:"examples,omitempty"`
	Scope     string `yaml:"scope"` // checkpoint|project - required if present
}

type ConversationItem struct {
	Exchange string `yaml:"exchange"`
	Outcome  string `yaml:"outcome,omitempty"`
}

// AppendContextEntry appends a context entry to the context file
func AppendContextEntry(contextPath string, entry *ContextEntry) error {
	// Render as YAML
	yamlData, err := yaml.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal context entry: %w", err)
	}

	// Open file for appending (create if doesn't exist)
	f, err := os.OpenFile(contextPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open context file: %w", err)
	}
	defer f.Close()

	// Append with document separator
	content := "---\n" + string(yamlData)
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("write context entry: %w", err)
	}

	return nil
}

// GetRecentContextEntries reads the last N context entries from the file
func GetRecentContextEntries(contextPath string, count int) ([]ContextEntry, error) {
	f, err := os.Open(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ContextEntry{}, nil // No context file yet
		}
		return nil, fmt.Errorf("open context file: %w", err)
	}
	defer f.Close()

	// Decode all documents
	decoder := yaml.NewDecoder(f)
	var entries []ContextEntry

	for {
		var entry ContextEntry
		if err := decoder.Decode(&entry); err != nil {
			break // End of documents
		}
		entries = append(entries, entry)
	}

	// Return last N entries
	if len(entries) <= count {
		return entries, nil
	}
	return entries[len(entries)-count:], nil
}

// RenderRecentContextForLLM returns recent context entries as formatted string
func RenderRecentContextForLLM(contextPath string, count int) (string, error) {
	entries, err := GetRecentContextEntries(contextPath, count)
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "", nil
	}

	// Render as YAML for LLM consumption
	var result string
	for _, entry := range entries {
		yamlData, err := yaml.Marshal(entry)
		if err != nil {
			continue
		}
		result += "---\n" + string(yamlData)
	}

	return result, nil
}

// GenerateContextTemplate generates the context input template section
func GenerateContextTemplate() string {
	return `
# CONTEXT SECTION (REQUIRED):
# Capture the reasoning and decision-making process for this checkpoint.
# This helps maintain continuity across development sessions with LLM agents.
#
# For each field, consider whether it's specific to this checkpoint (scope: checkpoint)
# or represents a project-wide pattern/principle (scope: project).
# Project-scoped items will be suggested for inclusion in .checkpoint-project.yml.

context:
  problem_statement: "[REQUIRED: What problem is this checkpoint solving?]"

  key_insights:
    - insight: "[REQUIRED: What did you learn during implementation?]"
      impact: "[OPTIONAL: How does this affect future development?]"
      scope: "[OPTIONAL: checkpoint|project - default is checkpoint]"

  decisions_made:
    - decision: "[REQUIRED: Significant architectural/implementation choice]"
      rationale: "[REQUIRED: Why this approach over alternatives?]"
      alternatives_considered:
        - "[OPTIONAL: Other approaches evaluated]"
      constraints_that_influenced: "[OPTIONAL: Limitations that drove this choice]"
      scope: "[OPTIONAL: checkpoint|project - default is checkpoint]"

  failed_approaches:
    - approach: "[OPTIONAL: What was tried but didn't work?]"
      why_failed: "[OPTIONAL: Specific reason for failure]"
      lessons_learned: "[OPTIONAL: What to avoid in future]"
      scope: "[OPTIONAL: checkpoint|project - default is checkpoint]"

  established_patterns:
    - pattern: "[OPTIONAL: New convention established]"
      rationale: "[OPTIONAL: Why this pattern works for this codebase]"
      examples: "[OPTIONAL: Where this pattern should be applied]"
      scope: "[REQUIRED if present: checkpoint|project]"

  conversation_context:
    - exchange: "[OPTIONAL: Key discussion points that influenced decisions]"
      outcome: "[OPTIONAL: How this shaped the implementation]"
`
}

// ParseContextFromInput extracts context from checkpoint input content
func ParseContextFromInput(inputContent string) (*CheckpointContext, error) {
	// This would parse the input YAML and extract the context section
	// For now, we'll rely on the full input parsing in schema package
	// This is a placeholder for future context-specific parsing
	return nil, fmt.Errorf("not implemented - use schema.ParseInputFile")
}

// CreateContextEntry creates a context entry from parsed input
func CreateContextEntry(timestamp string, ctx CheckpointContext) *ContextEntry {
	return &ContextEntry{
		SchemaVersion: "1",
		Timestamp:     timestamp,
		Context:       ctx,
	}
}

// UpdateContextEntryHash updates the commit hash in the last context entry
func UpdateContextEntryHash(contextPath, commitHash string) error {
	content, err := os.ReadFile(contextPath)
	if err != nil {
		return fmt.Errorf("read context file: %w", err)
	}

	contentStr := string(content)

	// Find the last occurrence of "commit_hash: " (empty or with value)
	lastHashIdx := -1
	searchStr := "commit_hash:"
	for i := len(contentStr) - len(searchStr); i >= 0; i-- {
		if contentStr[i:i+len(searchStr)] == searchStr {
			lastHashIdx = i
			break
		}
	}

	if lastHashIdx == -1 {
		return fmt.Errorf("could not find commit_hash field in last entry")
	}

	// Find the end of this line
	lineEnd := lastHashIdx
	for lineEnd < len(contentStr) && contentStr[lineEnd] != '\n' {
		lineEnd++
	}

	// Replace the line
	newLine := fmt.Sprintf("commit_hash: %s", commitHash)
	newContent := contentStr[:lastHashIdx] + newLine + contentStr[lineEnd:]

	// Write back
	return os.WriteFile(contextPath, []byte(newContent), 0644)
}
