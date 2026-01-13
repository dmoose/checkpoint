package project

import (
	"fmt"
	"os"
	"time"

	"github.com/dmoose/checkpoint/internal/language"

	"gopkg.in/yaml.v3"
)

// ProjectDocument represents the main curated project context
type ProjectDocument struct {
	SchemaVersion             string                     `yaml:"schema_version"`
	ProjectName               string                     `yaml:"project_name"`
	LastUpdated               string                     `yaml:"last_updated"`
	KeyInsights               []Insight                  `yaml:"key_insights,omitempty"`
	EstablishedPatterns       []Pattern                  `yaml:"established_patterns,omitempty"`
	FailedApproaches          []FailedApproach           `yaml:"failed_approaches,omitempty"`
	DesignPrinciples          []Principle                `yaml:"design_principles,omitempty"`
	IntegrationPoints         []Integration              `yaml:"integration_points,omitempty"`
	CodePatterns              []CodePattern              `yaml:"code_patterns,omitempty"`
	Dependencies              []Dependency               `yaml:"dependencies,omitempty"`
	LanguageRequirements      []LanguageRequirement      `yaml:"language_requirements,omitempty"`
	DeploymentTargets         []DeploymentTarget         `yaml:"deployment_targets,omitempty"`
	TestingMethodologies      []TestingMethodology       `yaml:"testing_methodologies,omitempty"`
	DevelopmentRoles          []DevelopmentRole          `yaml:"development_roles,omitempty"`
	ErrorHandlingPatterns     []ErrorHandlingPattern     `yaml:"error_handling_patterns,omitempty"`
	CompatibilityStrategy     []CompatibilityStrategy    `yaml:"compatibility_strategy,omitempty"`
	FileManagement            []FileManagement           `yaml:"file_management,omitempty"`
	SecurityConsiderations    []SecurityConsideration    `yaml:"security_considerations,omitempty"`
	PerformanceConsiderations []PerformanceConsideration `yaml:"performance_considerations,omitempty"`
	CrossCuttingConcerns      []CrossCuttingConcern      `yaml:"cross_cutting_concerns,omitempty"`
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

type Dependency struct {
	Name                   string `yaml:"name"`
	Purpose                string `yaml:"purpose,omitempty"`
	VersionConstraint      string `yaml:"version_constraint,omitempty"`
	Rationale              string `yaml:"rationale,omitempty"`
	AlternativesConsidered string `yaml:"alternatives_considered,omitempty"`
}

type LanguageRequirement struct {
	Language       string `yaml:"language"`
	MinimumVersion string `yaml:"minimum_version"`
	Rationale      string `yaml:"rationale,omitempty"`
}

type DeploymentTarget struct {
	Platform       string   `yaml:"platform"`
	MinimumVersion string   `yaml:"minimum_version,omitempty"`
	Architectures  []string `yaml:"architectures,omitempty"`
	Rationale      string   `yaml:"rationale,omitempty"`
}

type TestingMethodology struct {
	Approach       string `yaml:"approach"`
	Rationale      string `yaml:"rationale,omitempty"`
	Examples       string `yaml:"examples,omitempty"`
	CoverageTarget string `yaml:"coverage_target,omitempty"`
}

type DevelopmentRole struct {
	Role             string `yaml:"role"`
	Responsibilities string `yaml:"responsibilities"`
	Workflow         string `yaml:"workflow,omitempty"`
}

type ErrorHandlingPattern struct {
	Pattern   string `yaml:"pattern"`
	Rationale string `yaml:"rationale,omitempty"`
	Examples  string `yaml:"examples,omitempty"`
}

type CompatibilityStrategy struct {
	Principle     string `yaml:"principle"`
	Rationale     string `yaml:"rationale,omitempty"`
	MigrationPath string `yaml:"migration_path,omitempty"`
}

type FileManagement struct {
	File         string `yaml:"file"`
	Lifecycle    string `yaml:"lifecycle"`
	Ownership    string `yaml:"ownership"`
	TrackedInGit bool   `yaml:"tracked_in_git"`
}

type SecurityConsideration struct {
	Concern    string `yaml:"concern"`
	Mitigation string `yaml:"mitigation"`
	Guidance   string `yaml:"guidance,omitempty"`
}

type PerformanceConsideration struct {
	Aspect     string `yaml:"aspect"`
	Impact     string `yaml:"impact"`
	Mitigation string `yaml:"mitigation"`
	Threshold  string `yaml:"threshold,omitempty"`
}

type CrossCuttingConcern struct {
	Concern   string `yaml:"concern"`
	Approach  string `yaml:"approach"`
	Rationale string `yaml:"rationale,omitempty"`
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
	KeyInsights               []Insight                  `yaml:"key_insights,omitempty"`
	EstablishedPatterns       []Pattern                  `yaml:"established_patterns,omitempty"`
	FailedApproaches          []FailedApproach           `yaml:"failed_approaches,omitempty"`
	DesignPrinciples          []Principle                `yaml:"design_principles,omitempty"`
	IntegrationPoints         []Integration              `yaml:"integration_points,omitempty"`
	CodePatterns              []CodePattern              `yaml:"code_patterns,omitempty"`
	Dependencies              []Dependency               `yaml:"dependencies,omitempty"`
	LanguageRequirements      []LanguageRequirement      `yaml:"language_requirements,omitempty"`
	DeploymentTargets         []DeploymentTarget         `yaml:"deployment_targets,omitempty"`
	TestingMethodologies      []TestingMethodology       `yaml:"testing_methodologies,omitempty"`
	DevelopmentRoles          []DevelopmentRole          `yaml:"development_roles,omitempty"`
	ErrorHandlingPatterns     []ErrorHandlingPattern     `yaml:"error_handling_patterns,omitempty"`
	CompatibilityStrategy     []CompatibilityStrategy    `yaml:"compatibility_strategy,omitempty"`
	FileManagement            []FileManagement           `yaml:"file_management,omitempty"`
	SecurityConsiderations    []SecurityConsideration    `yaml:"security_considerations,omitempty"`
	PerformanceConsiderations []PerformanceConsideration `yaml:"performance_considerations,omitempty"`
	CrossCuttingConcerns      []CrossCuttingConcern      `yaml:"cross_cutting_concerns,omitempty"`
}

type ProjectUpdates struct {
	KeyInsights               []InsightUpdate                  `yaml:"key_insights,omitempty"`
	EstablishedPatterns       []PatternUpdate                  `yaml:"established_patterns,omitempty"`
	FailedApproaches          []FailedApproachUpdate           `yaml:"failed_approaches,omitempty"`
	DesignPrinciples          []PrincipleUpdate                `yaml:"design_principles,omitempty"`
	IntegrationPoints         []IntegrationUpdate              `yaml:"integration_points,omitempty"`
	CodePatterns              []CodePatternUpdate              `yaml:"code_patterns,omitempty"`
	Dependencies              []DependencyUpdate               `yaml:"dependencies,omitempty"`
	LanguageRequirements      []LanguageRequirementUpdate      `yaml:"language_requirements,omitempty"`
	DeploymentTargets         []DeploymentTargetUpdate         `yaml:"deployment_targets,omitempty"`
	TestingMethodologies      []TestingMethodologyUpdate       `yaml:"testing_methodologies,omitempty"`
	DevelopmentRoles          []DevelopmentRoleUpdate          `yaml:"development_roles,omitempty"`
	ErrorHandlingPatterns     []ErrorHandlingPatternUpdate     `yaml:"error_handling_patterns,omitempty"`
	CompatibilityStrategy     []CompatibilityStrategyUpdate    `yaml:"compatibility_strategy,omitempty"`
	FileManagement            []FileManagementUpdate           `yaml:"file_management,omitempty"`
	SecurityConsiderations    []SecurityConsiderationUpdate    `yaml:"security_considerations,omitempty"`
	PerformanceConsiderations []PerformanceConsiderationUpdate `yaml:"performance_considerations,omitempty"`
	CrossCuttingConcerns      []CrossCuttingConcernUpdate      `yaml:"cross_cutting_concerns,omitempty"`
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

type DependencyUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type LanguageRequirementUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type DeploymentTargetUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type TestingMethodologyUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type DevelopmentRoleUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type ErrorHandlingPatternUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type CompatibilityStrategyUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type FileManagementUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type SecurityConsiderationUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type PerformanceConsiderationUpdate struct {
	Existing string `yaml:"existing"`
	Updated  string `yaml:"updated"`
}

type CrossCuttingConcernUpdate struct {
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
		Dependencies: []Dependency{
			{
				Name:                   "Add external dependencies here",
				Purpose:                "Example: YAML parsing for multi-document files",
				VersionConstraint:      "Example: v3.0.1 or compatible",
				Rationale:              "Example: Must support multi-document format",
				AlternativesConsidered: "Example: Standard library (lacks feature X)",
			},
		},
		LanguageRequirements: []LanguageRequirement{
			{
				Language:       "Add language requirements here",
				MinimumVersion: "Example: 1.21",
				Rationale:      "Example: Uses newer standard library features",
			},
		},
		DeploymentTargets: []DeploymentTarget{
			{
				Platform:       "Add deployment targets here",
				MinimumVersion: "Example: 12.0",
				Architectures:  []string{"Example: arm64", "amd64"},
				Rationale:      "Example: Requires modern platform features",
			},
		},
		TestingMethodologies: []TestingMethodology{
			{
				Approach:       "Add testing approaches here",
				Rationale:      "Example: Comprehensive coverage with minimal duplication",
				Examples:       "Example: See test files in package X",
				CoverageTarget: "Example: 80%+ for core packages",
			},
		},
		DevelopmentRoles: []DevelopmentRole{
			{
				Role:             "Add development roles here",
				Responsibilities: "Example: Run commands, maintain quality",
				Workflow:         "Example: Initiates process, reviews output",
			},
		},
		ErrorHandlingPatterns: []ErrorHandlingPattern{
			{
				Pattern:   "Add error handling patterns here",
				Rationale: "Example: User-facing tool needs graceful degradation",
				Examples:  "Example: File I/O errors, command failures",
			},
		},
		CompatibilityStrategy: []CompatibilityStrategy{
			{
				Principle:     "Add compatibility strategies here",
				Rationale:     "Example: Enables future migrations and multi-version support",
				MigrationPath: "Example: Read old versions, write current version",
			},
		},
		FileManagement: []FileManagement{
			{
				File:         "Add file management policies here",
				Lifecycle:    "Example: Created by init, appended by commit",
				Ownership:    "Example: Tool-managed, human readable",
				TrackedInGit: true,
			},
		},
		SecurityConsiderations: []SecurityConsideration{
			{
				Concern:    "Add security considerations here",
				Mitigation: "Example: Files are git-tracked; avoid sensitive data",
				Guidance:   "Example: Review before commit; use environment variables",
			},
		},
		PerformanceConsiderations: []PerformanceConsideration{
			{
				Aspect:     "Add performance considerations here",
				Impact:     "Example: Linear growth with item count",
				Mitigation: "Example: Recent-only loading keeps memory bounded",
				Threshold:  "Example: Acceptable up to ~1000 items",
			},
		},
		CrossCuttingConcerns: []CrossCuttingConcern{
			{
				Concern:   "Add cross-cutting concerns here",
				Approach:  "Example: Always use RFC3339 with timezone",
				Rationale: "Example: Enables correlation across contexts",
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
	defer func() { _ = f.Close() }()

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
	defer func() { _ = f.Close() }()

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
