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
	Context       CheckpointContext `yaml:"context"`
}

// CheckpointContext represents the context captured at a checkpoint
type CheckpointContext struct {
	ProblemStatement string           `yaml:"problem_statement"`
	KeyInsights      []Insight        `yaml:"key_insights,omitempty"`
	DecisionsMade    []Decision       `yaml:"decisions_made,omitempty"`
	FailedApproaches []FailedApproach `yaml:"failed_approaches,omitempty"`
	KeyExchanges     []KeyExchange    `yaml:"key_exchanges,omitempty"`
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

type KeyExchange struct {
	PromptSummary string `yaml:"prompt_summary"`
	ApproachTaken string `yaml:"approach_taken,omitempty"`
	HumanFeedback string `yaml:"human_feedback,omitempty"`
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
	defer func() { _ = f.Close() }()

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
	defer func() { _ = f.Close() }()

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
# SCOPE: checkpoint = this change only | project = becomes recommendation in .checkpoint/project.yaml

context:
  problem_statement: "[REQUIRED: What problem is this checkpoint solving?]"

  key_insights:
    - insight: "[What did you learn?]"
      impact: "[How does this affect future development?]"
      scope: "[checkpoint|project]"

  decisions_made:
    - decision: "[Significant choice made]"
      rationale: "[Why this approach?]"
      alternatives_considered:
        - "[Other approaches evaluated]"
      scope: "[checkpoint|project]"

  failed_approaches:
    - approach: "[What didn't work?]"
      why_failed: "[Why?]"
      lessons_learned: "[What to avoid]"
      scope: "[checkpoint|project]"

  key_exchanges:
    - prompt_summary: "[What the human asked the LLM to do]"
      approach_taken: "[What the LLM proposed/did]"
      human_feedback: "[How the human responded - approved, modified, rejected]"
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
