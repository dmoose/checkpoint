package project

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Insight struct {
	Insight             string `yaml:"insight"`
	Rationale           string `yaml:"rationale,omitempty"`
	EstablishedInCommit string `yaml:"established_in_commit,omitempty"`
}

type Pattern struct {
	Pattern             string `yaml:"pattern"`
	Rationale           string `yaml:"rationale,omitempty"`
	Examples            string `yaml:"examples,omitempty"`
	EstablishedInCommit string `yaml:"established_in_commit,omitempty"`
}

type FailedApproach struct {
	Approach           string `yaml:"approach"`
	WhyFailed          string `yaml:"why_failed,omitempty"`
	LessonsLearned     string `yaml:"lessons_learned,omitempty"`
	DiscoveredInCommit string `yaml:"discovered_in_commit,omitempty"`
}

type Principle struct {
	Principle string `yaml:"principle"`
	Rationale string `yaml:"rationale,omitempty"`
	AppliesTo string `yaml:"applies_to,omitempty"`
}

// RecommendationsDocument represents LLM-generated suggestions
type RecommendationsDocument struct {
	SchemaVersion        string            `yaml:"schema_version"`
	DocumentType         string            `yaml:"document_type"` // "recommendations"
	Timestamp            string            `yaml:"timestamp"`
	RecommendedAdditions ProjectAdditions  `yaml:"recommended_additions,omitempty"`
	RecommendedDeletions []ProjectDeletion `yaml:"recommended_deletions,omitempty"`
}

type ProjectAdditions struct {
	KeyInsights      []Insight        `yaml:"key_insights,omitempty"`
	FailedApproaches []FailedApproach `yaml:"failed_approaches,omitempty"`
	DesignPrinciples []Principle      `yaml:"design_principles,omitempty"`
}

type ProjectDeletion struct {
	Section string `yaml:"section"`
	Item    string `yaml:"item"`
	Reason  string `yaml:"reason,omitempty"`
}

// AppendRecommendations adds a recommendations document to the project file
func AppendRecommendations(projectPath, timestamp string, additions ProjectAdditions, deletions []ProjectDeletion) error {
	// Create recommendations document
	rec := RecommendationsDocument{
		SchemaVersion:        "1",
		DocumentType:         "recommendations",
		Timestamp:            timestamp,
		RecommendedAdditions: additions,
		RecommendedDeletions: deletions,
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal recommendations: %w", err)
	}

	// Append to project file with document separator
	f, err := os.OpenFile(projectPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open project file: %w", err)
	}
	defer func() { _ = f.Close() }()

	content := "---\n" + string(yamlData)
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("append recommendations: %w", err)
	}

	return nil
}
