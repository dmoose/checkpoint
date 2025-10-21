package project

import (
	"fmt"
	"os"
	"time"

	"go-llm/internal/language"

	"gopkg.in/yaml.v3"
)

// ProjectDocument represents the main curated project context
type ProjectDocument struct {
	SchemaVersion       string           `yaml:"schema_version"`
	ProjectName         string           `yaml:"project_name"`
	LastUpdated         string           `yaml:"last_updated"`
	KeyInsights         []Insight        `yaml:"key_insights,omitempty"`
	EstablishedPatterns []Pattern        `yaml:"established_patterns,omitempty"`
	FailedApproaches    []FailedApproach `yaml:"failed_approaches,omitempty"`
	DesignPrinciples    []Principle      `yaml:"design_principles,omitempty"`
	IntegrationPoints   []Integration    `yaml:"integration_points,omitempty"`
	CodePatterns        []CodePattern    `yaml:"code_patterns,omitempty"`
}

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

type Integration struct {
	System       string `yaml:"system"`
	Interaction  string `yaml:"interaction,omitempty"`
	Requirements string `yaml:"requirements,omitempty"`
	Constraints  string `yaml:"constraints,omitempty"`
}

type CodePattern struct {
	Preference string `yaml:"preference"`
	Rationale  string `yaml:"rationale,omitempty"`
	Examples   string `yaml:"examples,omitempty"`
}

// RecommendationsDocument represents LLM-generated suggestions
type RecommendationsDocument struct {
	SchemaVersion        string            `yaml:"schema_version"`
	DocumentType         string            `yaml:"document_type"` // "recommendations"
	Timestamp            string            `yaml:"timestamp"`
	RecommendedAdditions ProjectAdditions  `yaml:"recommended_additions,omitempty"`
	RecommendedUpdates   ProjectUpdates    `yaml:"recommended_updates,omitempty"`
	RecommendedDeletions []ProjectDeletion `yaml:"recommended_deletions,omitempty"`
}

type ProjectAdditions struct {
	KeyInsights         []Insight        `yaml:"key_insights,omitempty"`
	EstablishedPatterns []Pattern        `yaml:"established_patterns,omitempty"`
	FailedApproaches    []FailedApproach `yaml:"failed_approaches,omitempty"`
	DesignPrinciples    []Principle      `yaml:"design_principles,omitempty"`
	IntegrationPoints   []Integration    `yaml:"integration_points,omitempty"`
	CodePatterns        []CodePattern    `yaml:"code_patterns,omitempty"`
}

type ProjectUpdates struct {
	KeyInsights         []InsightUpdate        `yaml:"key_insights,omitempty"`
	EstablishedPatterns []PatternUpdate        `yaml:"established_patterns,omitempty"`
	FailedApproaches    []FailedApproachUpdate `yaml:"failed_approaches,omitempty"`
	DesignPrinciples    []PrincipleUpdate      `yaml:"design_principles,omitempty"`
	IntegrationPoints   []IntegrationUpdate    `yaml:"integration_points,omitempty"`
	CodePatterns        []CodePatternUpdate    `yaml:"code_patterns,omitempty"`
}

type InsightUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type PatternUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type FailedApproachUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type PrincipleUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type IntegrationUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type CodePatternUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type ProjectDeletion struct {
	Section string `yaml:"section"`
	Item    string `yaml:"item"`
	Reason  string `yaml:"reason,omitempty"`
}

// InitializeProjectFile creates the initial project document
func InitializeProjectFile(projectPath, projectName string, languages []language.Language) error {
	projectFile := projectPath

	// Check if file already exists
	if _, err := os.Stat(projectFile); err == nil {
		// File exists, don't overwrite
		return nil
	}

	// Create initial project document with all structured fields
	doc := ProjectDocument{
		SchemaVersion: "1",
		ProjectName:   projectName,
		LastUpdated:   time.Now().Format(time.RFC3339),
		KeyInsights: []Insight{
			{
				Insight:   "Add project-wide insights here that affect all development",
				Rationale: "Example: Minimal dependencies philosophy drives architectural decisions",
			},
		},
		EstablishedPatterns: []Pattern{
			{
				Pattern:   "Add project-wide patterns and conventions here",
				Rationale: "Example: Use table-driven tests for multiple scenario validation",
				Examples:  "Example: See language detection tests, numstat parsing tests",
			},
		},
		FailedApproaches: []FailedApproach{
			{
				Approach:       "Document anti-patterns and failed approaches here",
				WhyFailed:      "Example: Added complexity without providing clear value",
				LessonsLearned: "Example: Simpler solutions often outperform complex ones",
			},
		},
		DesignPrinciples: []Principle{
			{
				Principle: "Add core design principles here",
				Rationale: "Example: Append-only file structures prevent corruption and maintain history",
				AppliesTo: "Example: changelog, context files, any historical data",
			},
		},
		IntegrationPoints: []Integration{
			{
				System:       "Add external system integrations here",
				Interaction:  "Example: Reads changelog files for daily summaries",
				Requirements: "Example: YAML multi-document format, consistent schema versions",
				Constraints:  "Example: Files must be parseable independently",
			},
		},
		CodePatterns: []CodePattern{
			{
				Preference: "Add code style preferences here",
				Rationale:  "Example: Explicit error handling over silent failures",
				Examples:   "Example: git command failures, file I/O operations",
			},
		},
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal project document: %w", err)
	}

	// Write to file with document separator
	content := "---\n" + string(yamlData)
	if err := os.WriteFile(projectFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write project file: %w", err)
	}

	return nil
}

// AppendRecommendations adds a recommendations document to the project file
func AppendRecommendations(projectPath, timestamp string, additions ProjectAdditions, updates ProjectUpdates, deletions []ProjectDeletion) error {
	// Create recommendations document
	rec := RecommendationsDocument{
		SchemaVersion:        "1",
		DocumentType:         "recommendations",
		Timestamp:            timestamp,
		RecommendedAdditions: additions,
		RecommendedUpdates:   updates,
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
	defer f.Close()

	content := "---\n" + string(yamlData)
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("append recommendations: %w", err)
	}

	return nil
}

// ReadProjectDocument reads only the first (main) document from the project file
func ReadProjectDocument(projectPath string) (*ProjectDocument, error) {
	f, err := os.Open(projectPath)
	if err != nil {
		return nil, fmt.Errorf("open project file: %w", err)
	}
	defer f.Close()

	// Decode only the first document
	decoder := yaml.NewDecoder(f)
	var doc ProjectDocument
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode project document: %w", err)
	}

	return &doc, nil
}

// RenderProjectDocumentForLLM returns the first project document as a string for LLM context
func RenderProjectDocumentForLLM(projectPath string) (string, error) {
	content, err := os.ReadFile(projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No project file yet
		}
		return "", fmt.Errorf("read project file: %w", err)
	}

	// Find the first document (up to second --- or end of file)
	contentStr := string(content)

	// Skip first --- if present
	if len(contentStr) >= 4 && contentStr[:4] == "---\n" {
		contentStr = contentStr[4:]
	}

	// Find next ---
	nextSep := -1
	for i := 0; i < len(contentStr)-4; i++ {
		if contentStr[i:i+4] == "\n---" || contentStr[i:i+5] == "\n---\n" {
			nextSep = i
			break
		}
	}

	if nextSep != -1 {
		return contentStr[:nextSep], nil
	}

	return contentStr, nil
}
