package explain

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dmoose/checkpoint/pkg/config"

	"gopkg.in/yaml.v3"
)

// CheckpointEntry represents a checkpoint from the changelog
type CheckpointEntry struct {
	SchemaVersion string                 `yaml:"schema_version"`
	Timestamp     string                 `yaml:"timestamp"`
	CommitHash    string                 `yaml:"commit_hash"`
	Changes       []ChangeEntry          `yaml:"changes"`
	NextSteps     []NextStepEntry        `yaml:"next_steps"`
	Context       map[string]interface{} `yaml:"context"`
}

// ChangeEntry represents a change in a checkpoint
type ChangeEntry struct {
	Summary    string `yaml:"summary"`
	Details    string `yaml:"details"`
	ChangeType string `yaml:"change_type"`
	Scope      string `yaml:"scope"`
}

// NextStepEntry represents a planned next step
type NextStepEntry struct {
	Summary  string `yaml:"summary"`
	Details  string `yaml:"details"`
	Priority string `yaml:"priority"`
	Scope    string `yaml:"scope"`
}

// ContextEntry represents a context document
type ContextEntry struct {
	SchemaVersion       string        `yaml:"schema_version"`
	Timestamp           string        `yaml:"timestamp"`
	CommitHash          string        `yaml:"commit_hash"`
	ProblemStatement    string        `yaml:"problem_statement"`
	KeyInsights         []interface{} `yaml:"key_insights"`
	DecisionsMade       []interface{} `yaml:"decisions_made"`
	EstablishedPatterns []interface{} `yaml:"established_patterns"`
	FailedApproaches    []interface{} `yaml:"failed_approaches"`
}

// HistoryData holds aggregated history data
type HistoryData struct {
	RecentCheckpoints []CheckpointEntry
	AllNextSteps      []NextStepWithSource
	RecentPatterns    []PatternWithSource
	RecentDecisions   []DecisionWithSource
	RecentFailed      []FailedWithSource
}

// NextStepWithSource includes the source checkpoint info
type NextStepWithSource struct {
	NextStepEntry
	FromTimestamp string
	FromCommit    string
}

// PatternWithSource includes source info
type PatternWithSource struct {
	Content       string
	Rationale     string
	FromTimestamp string
}

// DecisionWithSource includes source info
type DecisionWithSource struct {
	Content       string
	Rationale     string
	FromTimestamp string
}

// FailedWithSource includes source info
type FailedWithSource struct {
	Approach      string
	WhyFailed     string
	FromTimestamp string
}

// LoadHistory loads recent checkpoint history
func LoadHistory(projectPath string, limit int) (*HistoryData, error) {
	if limit <= 0 {
		limit = 10
	}

	data := &HistoryData{}

	// Load changelog
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if entries, err := loadChangelogEntries(changelogPath, limit); err == nil {
		data.RecentCheckpoints = entries

		// Extract next_steps from all checkpoints
		for _, entry := range entries {
			for _, step := range entry.NextSteps {
				data.AllNextSteps = append(data.AllNextSteps, NextStepWithSource{
					NextStepEntry: step,
					FromTimestamp: entry.Timestamp,
					FromCommit:    entry.CommitHash,
				})
			}
		}
	}

	// Load context for patterns, decisions, failed approaches
	contextPath := filepath.Join(projectPath, config.ContextFileName)
	if contexts, err := loadContextEntries(contextPath, limit); err == nil {
		for _, ctx := range contexts {
			// Extract patterns
			for _, p := range ctx.EstablishedPatterns {
				pattern := extractPatternContent(p)
				if pattern.Content != "" {
					pattern.FromTimestamp = ctx.Timestamp
					data.RecentPatterns = append(data.RecentPatterns, pattern)
				}
			}
			// Extract decisions
			for _, d := range ctx.DecisionsMade {
				decision := extractDecisionContent(d)
				if decision.Content != "" {
					decision.FromTimestamp = ctx.Timestamp
					data.RecentDecisions = append(data.RecentDecisions, decision)
				}
			}
			// Extract failed approaches
			for _, f := range ctx.FailedApproaches {
				failed := extractFailedContent(f)
				if failed.Approach != "" {
					failed.FromTimestamp = ctx.Timestamp
					data.RecentFailed = append(data.RecentFailed, failed)
				}
			}
		}
	}

	return data, nil
}

func loadChangelogEntries(path string, limit int) ([]CheckpointEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	docs := splitYAMLDocs(string(data))
	var entries []CheckpointEntry

	// Skip meta document, process from newest
	for i := len(docs) - 1; i >= 1 && len(entries) < limit; i-- {
		var entry CheckpointEntry
		if err := yaml.Unmarshal([]byte(docs[i]), &entry); err == nil {
			if entry.Timestamp != "" { // Valid checkpoint entry
				entries = append(entries, entry)
			}
		}
	}

	return entries, nil
}

func loadContextEntries(path string, limit int) ([]ContextEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	docs := splitYAMLDocs(string(data))
	var entries []ContextEntry

	// Process from newest
	for i := len(docs) - 1; i >= 0 && len(entries) < limit; i-- {
		var entry ContextEntry
		if err := yaml.Unmarshal([]byte(docs[i]), &entry); err == nil {
			if entry.Timestamp != "" {
				entries = append(entries, entry)
			}
		}
	}

	return entries, nil
}

func splitYAMLDocs(content string) []string {
	var docs []string
	parts := strings.Split(content, "\n---")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && part != "---" {
			docs = append(docs, part)
		}
	}
	return docs
}

func extractPatternContent(item interface{}) PatternWithSource {
	switch v := item.(type) {
	case string:
		return PatternWithSource{Content: v}
	case map[string]interface{}:
		p := PatternWithSource{}
		if pattern, ok := v["pattern"].(string); ok {
			p.Content = pattern
		}
		if rationale, ok := v["rationale"].(string); ok {
			p.Rationale = rationale
		}
		return p
	}
	return PatternWithSource{}
}

func extractDecisionContent(item interface{}) DecisionWithSource {
	switch v := item.(type) {
	case string:
		return DecisionWithSource{Content: v}
	case map[string]interface{}:
		d := DecisionWithSource{}
		if decision, ok := v["decision"].(string); ok {
			d.Content = decision
		} else if desc, ok := v["description"].(string); ok {
			d.Content = desc
		}
		if rationale, ok := v["rationale"].(string); ok {
			d.Rationale = rationale
		}
		return d
	}
	return DecisionWithSource{}
}

func extractFailedContent(item interface{}) FailedWithSource {
	switch v := item.(type) {
	case string:
		return FailedWithSource{Approach: v}
	case map[string]interface{}:
		f := FailedWithSource{}
		if approach, ok := v["approach"].(string); ok {
			f.Approach = approach
		}
		if why, ok := v["why_failed"].(string); ok {
			f.WhyFailed = why
		}
		return f
	}
	return FailedWithSource{}
}

// RenderHistory returns formatted history output
func RenderHistory(projectPath string, limit int) string {
	history, err := LoadHistory(projectPath, limit)
	if err != nil {
		return fmt.Sprintf("Error loading history: %v\n", err)
	}

	var sb strings.Builder
	sb.WriteString("# Recent History\n\n")

	// Recent checkpoints
	sb.WriteString("## Recent Checkpoints\n\n")
	if len(history.RecentCheckpoints) == 0 {
		sb.WriteString("No checkpoints yet.\n\n")
	} else {
		for _, cp := range history.RecentCheckpoints {
			commit := ""
			if cp.CommitHash != "" {
				commit = fmt.Sprintf(" [%s]", cp.CommitHash[:minInt(8, len(cp.CommitHash))])
			}
			sb.WriteString(fmt.Sprintf("### %s%s\n\n", cp.Timestamp, commit))

			for _, change := range cp.Changes {
				typeStr := ""
				if change.ChangeType != "" {
					typeStr = fmt.Sprintf(" (%s)", change.ChangeType)
				}
				sb.WriteString(fmt.Sprintf("- %s%s\n", change.Summary, typeStr))
			}
			sb.WriteString("\n")
		}
	}

	// Outstanding next steps
	sb.WriteString("## Outstanding Next Steps\n\n")
	if len(history.AllNextSteps) == 0 {
		sb.WriteString("No outstanding next steps.\n\n")
	} else {
		// Sort by priority
		sorted := make([]NextStepWithSource, len(history.AllNextSteps))
		copy(sorted, history.AllNextSteps)
		sort.Slice(sorted, func(i, j int) bool {
			return priorityRank(sorted[i].Priority) > priorityRank(sorted[j].Priority)
		})

		for _, step := range sorted {
			priority := ""
			if step.Priority != "" {
				priority = fmt.Sprintf(" [%s]", step.Priority)
			}
			scope := ""
			if step.Scope != "" {
				scope = fmt.Sprintf(" (%s)", step.Scope)
			}
			sb.WriteString(fmt.Sprintf("- %s%s%s\n", step.Summary, priority, scope))
		}
		sb.WriteString("\n")
	}

	// Recent patterns
	if len(history.RecentPatterns) > 0 {
		sb.WriteString("## Recently Established Patterns\n\n")
		seen := make(map[string]bool)
		for _, p := range history.RecentPatterns {
			if seen[p.Content] {
				continue
			}
			seen[p.Content] = true
			sb.WriteString(fmt.Sprintf("- %s\n", p.Content))
			if p.Rationale != "" {
				sb.WriteString(fmt.Sprintf("  *%s*\n", p.Rationale))
			}
		}
		sb.WriteString("\n")
	}

	// Recent decisions
	if len(history.RecentDecisions) > 0 {
		sb.WriteString("## Recent Decisions\n\n")
		seen := make(map[string]bool)
		count := 0
		for _, d := range history.RecentDecisions {
			if seen[d.Content] || count >= 5 {
				continue
			}
			seen[d.Content] = true
			count++
			sb.WriteString(fmt.Sprintf("- %s\n", d.Content))
			if d.Rationale != "" {
				sb.WriteString(fmt.Sprintf("  *%s*\n", d.Rationale))
			}
		}
		sb.WriteString("\n")
	}

	// Failed approaches
	if len(history.RecentFailed) > 0 {
		sb.WriteString("## Failed Approaches (Don't Repeat)\n\n")
		seen := make(map[string]bool)
		for _, f := range history.RecentFailed {
			if seen[f.Approach] {
				continue
			}
			seen[f.Approach] = true
			sb.WriteString(fmt.Sprintf("- %s\n", f.Approach))
			if f.WhyFailed != "" {
				sb.WriteString(fmt.Sprintf("  *Why: %s*\n", f.WhyFailed))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderNext returns all outstanding next steps
func RenderNext(projectPath string) string {
	history, err := LoadHistory(projectPath, 50) // Look at more checkpoints for next steps
	if err != nil {
		return fmt.Sprintf("Error loading history: %v\n", err)
	}

	var sb strings.Builder
	sb.WriteString("# Outstanding Next Steps\n\n")

	if len(history.AllNextSteps) == 0 {
		sb.WriteString("No outstanding next steps.\n")
		sb.WriteString("\nhint: Next steps are captured during checkpoint commit.\n")
		return sb.String()
	}

	// Group by priority
	high := []NextStepWithSource{}
	med := []NextStepWithSource{}
	low := []NextStepWithSource{}
	other := []NextStepWithSource{}

	for _, step := range history.AllNextSteps {
		switch strings.ToLower(step.Priority) {
		case "high":
			high = append(high, step)
		case "med", "medium":
			med = append(med, step)
		case "low":
			low = append(low, step)
		default:
			other = append(other, step)
		}
	}

	if len(high) > 0 {
		sb.WriteString("## High Priority\n\n")
		for _, step := range high {
			renderNextStep(&sb, step)
		}
		sb.WriteString("\n")
	}

	if len(med) > 0 {
		sb.WriteString("## Medium Priority\n\n")
		for _, step := range med {
			renderNextStep(&sb, step)
		}
		sb.WriteString("\n")
	}

	if len(low) > 0 {
		sb.WriteString("## Low Priority\n\n")
		for _, step := range low {
			renderNextStep(&sb, step)
		}
		sb.WriteString("\n")
	}

	if len(other) > 0 {
		sb.WriteString("## Other\n\n")
		for _, step := range other {
			renderNextStep(&sb, step)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderNextStep(sb *strings.Builder, step NextStepWithSource) {
	scope := ""
	if step.Scope != "" {
		scope = fmt.Sprintf(" [%s]", step.Scope)
	}
	fmt.Fprintf(sb, "- %s%s\n", step.Summary, scope)
	if step.Details != "" {
		fmt.Fprintf(sb, "  %s\n", step.Details)
	}
	if step.FromCommit != "" {
		fmt.Fprintf(sb, "  *(from %s)*\n", step.FromCommit[:minInt(8, len(step.FromCommit))])
	}
}

func priorityRank(p string) int {
	switch strings.ToLower(p) {
	case "high":
		return 3
	case "med", "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
