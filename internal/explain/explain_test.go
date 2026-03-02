package explain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmoose/checkpoint/pkg/config"
)

// --- LoadExplainContext tests ---

func TestLoadExplainContext_AllFiles(t *testing.T) {
	dir := t.TempDir()
	cpDir := filepath.Join(dir, config.CheckpointDir)
	if err := os.MkdirAll(cpDir, 0o755); err != nil {
		t.Fatal(err)
	}

	projectYAML := `schema_version: "1"
name: test-project
purpose: testing
type: cli
`
	toolsYAML := `schema_version: "1"
build:
  default:
    command: make build
test:
  default:
    command: go test ./...
`
	guidelinesYAML := `schema_version: "1"
rules:
  - "test rule"
avoid:
  - "bad pattern"
principles:
  - "keep it simple"
`
	skillsYAML := `schema_version: "1"
local: []
global: []
`

	files := map[string]string{
		config.ExplainProjectYaml:    projectYAML,
		config.ExplainToolsYaml:      toolsYAML,
		config.ExplainGuidelinesYaml: guidelinesYAML,
		config.ExplainSkillsYaml:     skillsYAML,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(cpDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	out, err := LoadExplainContext(dir)
	if err != nil {
		t.Fatalf("LoadExplainContext returned error: %v", err)
	}

	if out.Project == nil {
		t.Fatal("expected Project to be loaded")
	}
	if out.Project.Name != "test-project" {
		t.Errorf("Project.Name = %q, want %q", out.Project.Name, "test-project")
	}
	if out.Project.Purpose != "testing" {
		t.Errorf("Project.Purpose = %q, want %q", out.Project.Purpose, "testing")
	}
	if out.Project.Type != "cli" {
		t.Errorf("Project.Type = %q, want %q", out.Project.Type, "cli")
	}

	if out.Tools == nil {
		t.Fatal("expected Tools to be loaded")
	}
	if cmd, ok := out.Tools.Build["default"]; !ok || cmd.Command != "make build" {
		t.Errorf("Tools.Build[default] = %+v, want command 'make build'", out.Tools.Build["default"])
	}
	if cmd, ok := out.Tools.Test["default"]; !ok || cmd.Command != "go test ./..." {
		t.Errorf("Tools.Test[default] = %+v, want command 'go test ./...'", out.Tools.Test["default"])
	}

	if out.Guidelines == nil {
		t.Fatal("expected Guidelines to be loaded")
	}
	if len(out.Guidelines.Rules) != 1 || out.Guidelines.Rules[0] != "test rule" {
		t.Errorf("Guidelines.Rules = %v, want [test rule]", out.Guidelines.Rules)
	}
	if len(out.Guidelines.Avoid) != 1 || out.Guidelines.Avoid[0] != "bad pattern" {
		t.Errorf("Guidelines.Avoid = %v, want [bad pattern]", out.Guidelines.Avoid)
	}

	if out.Skills == nil {
		t.Fatal("expected Skills to be loaded")
	}
	if out.Skills.SchemaVersion != "1" {
		t.Errorf("Skills.SchemaVersion = %q, want %q", out.Skills.SchemaVersion, "1")
	}
}

func TestLoadExplainContext_MissingFiles(t *testing.T) {
	dir := t.TempDir()
	// No .checkpoint directory at all

	out, err := LoadExplainContext(dir)
	if err != nil {
		t.Fatalf("LoadExplainContext returned error: %v", err)
	}

	if out.Project != nil {
		t.Error("expected Project to be nil when file missing")
	}
	if out.Tools != nil {
		t.Error("expected Tools to be nil when file missing")
	}
	if out.Guidelines != nil {
		t.Error("expected Guidelines to be nil when file missing")
	}
	if out.Skills != nil {
		t.Error("expected Skills to be nil when file missing")
	}
	if out.ProjectPath != dir {
		t.Errorf("ProjectPath = %q, want %q", out.ProjectPath, dir)
	}
}

func TestLoadExplainContext_PartialFiles(t *testing.T) {
	dir := t.TempDir()
	cpDir := filepath.Join(dir, config.CheckpointDir)
	if err := os.MkdirAll(cpDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Only write project.yaml
	projectYAML := `schema_version: "1"
name: partial-project
purpose: partial test
`
	if err := os.WriteFile(filepath.Join(cpDir, config.ExplainProjectYaml), []byte(projectYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := LoadExplainContext(dir)
	if err != nil {
		t.Fatalf("LoadExplainContext returned error: %v", err)
	}

	if out.Project == nil {
		t.Fatal("expected Project to be loaded")
	}
	if out.Project.Name != "partial-project" {
		t.Errorf("Project.Name = %q, want %q", out.Project.Name, "partial-project")
	}
	if out.Tools != nil {
		t.Error("expected Tools to be nil")
	}
	if out.Guidelines != nil {
		t.Error("expected Guidelines to be nil")
	}
	if out.Skills != nil {
		t.Error("expected Skills to be nil")
	}
}

func TestLoadExplainContext_Learnings(t *testing.T) {
	dir := t.TempDir()
	cpDir := filepath.Join(dir, config.CheckpointDir)
	if err := os.MkdirAll(cpDir, 0o755); err != nil {
		t.Fatal(err)
	}

	learningsYAML := `timestamp: "2026-01-15T10:00:00Z"
learning: "first insight"
---
timestamp: "2026-01-16T10:00:00Z"
learning: "second insight"
`
	if err := os.WriteFile(filepath.Join(cpDir, "learnings.yaml"), []byte(learningsYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := LoadExplainContext(dir)
	if err != nil {
		t.Fatalf("LoadExplainContext returned error: %v", err)
	}

	if len(out.Learnings) != 2 {
		t.Fatalf("expected 2 learnings, got %d", len(out.Learnings))
	}
	if out.Learnings[0].Learning != "first insight" {
		t.Errorf("Learnings[0].Learning = %q, want %q", out.Learnings[0].Learning, "first insight")
	}
	if out.Learnings[1].Learning != "second insight" {
		t.Errorf("Learnings[1].Learning = %q, want %q", out.Learnings[1].Learning, "second insight")
	}
}

func TestLoadExplainContext_LocalSkills(t *testing.T) {
	dir := t.TempDir()
	cpDir := filepath.Join(dir, config.CheckpointDir)
	skillDir := filepath.Join(cpDir, config.SkillsDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	skillsYAML := `schema_version: "1"
local:
  - my-skill
`
	if err := os.WriteFile(filepath.Join(cpDir, config.ExplainSkillsYaml), []byte(skillsYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.md"), []byte("# My Skill\nDoes things."), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := LoadExplainContext(dir)
	if err != nil {
		t.Fatalf("LoadExplainContext returned error: %v", err)
	}

	if len(out.SkillDefs) != 1 {
		t.Fatalf("expected 1 skill def, got %d", len(out.SkillDefs))
	}
	if out.SkillDefs[0].Name != "my-skill" {
		t.Errorf("SkillDefs[0].Name = %q, want %q", out.SkillDefs[0].Name, "my-skill")
	}
	if !out.SkillDefs[0].IsLocal {
		t.Error("expected skill to be local")
	}
	if out.SkillDefs[0].Content != "# My Skill\nDoes things." {
		t.Errorf("SkillDefs[0].Content = %q, want %q", out.SkillDefs[0].Content, "# My Skill\nDoes things.")
	}
}

// --- RenderSummary tests ---

func TestRenderSummary_WithAllData(t *testing.T) {
	e := &ExplainOutput{
		Project: &ProjectConfig{
			Name:    "my-app",
			Type:    "cli",
			Purpose: "A test application",
			Architecture: ArchitectureConfig{
				KeyPaths: map[string]string{
					"src": "cmd/",
				},
			},
		},
		Tools: &ToolsConfig{
			Build: map[string]ToolCommand{
				"default": {Command: "make build"},
			},
			Test: map[string]ToolCommand{
				"default": {Command: "go test ./..."},
			},
			Check: map[string]ToolCommand{
				"default": {Command: "make check"},
			},
		},
		Guidelines: &GuidelinesConfig{
			Rules: []string{"rule one", "rule two"},
		},
		SkillDefs: []Skill{
			{Name: "deploy", IsLocal: true},
		},
		Learnings: []Learning{
			{Timestamp: "2026-01-15T10:00:00Z", Learning: "learned something"},
		},
	}

	result := e.RenderSummary()

	checks := []string{
		"PROJECT: my-app (cli)",
		"PURPOSE: A test application",
		"build: make build",
		"test:  go test ./...",
		"check: make check",
		"src: cmd/",
		"deploy",
		"rule one",
		"rule two",
		"LEARNINGS: 1 captured insights",
		"learned something",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("RenderSummary missing %q", check)
		}
	}
}

func TestRenderSummary_NoProject(t *testing.T) {
	e := &ExplainOutput{}
	result := e.RenderSummary()

	if !strings.Contains(result, "PROJECT: (not configured)") {
		t.Error("expected '(not configured)' hint when no project")
	}
	if !strings.Contains(result, "checkpoint init") {
		t.Error("expected init hint when no project")
	}
}

func TestRenderSummary_NoTools(t *testing.T) {
	e := &ExplainOutput{
		Project: &ProjectConfig{Name: "x", Type: "lib"},
	}
	result := e.RenderSummary()

	if !strings.Contains(result, "no tools configured") {
		t.Error("expected 'no tools configured' when tools is nil")
	}
}

func TestRenderSummary_LintFallback(t *testing.T) {
	e := &ExplainOutput{
		Project: &ProjectConfig{Name: "x", Type: "lib"},
		Tools: &ToolsConfig{
			Lint: map[string]ToolCommand{
				"default": {Command: "golangci-lint run"},
			},
		},
	}
	result := e.RenderSummary()

	if !strings.Contains(result, "lint:  golangci-lint run") {
		t.Error("expected lint fallback when no check command")
	}
}

func TestRenderSummary_ManyRules(t *testing.T) {
	e := &ExplainOutput{
		Guidelines: &GuidelinesConfig{
			Rules: []string{"r1", "r2", "r3", "r4", "r5"},
		},
	}
	result := e.RenderSummary()

	if !strings.Contains(result, "... and 2 more") {
		t.Error("expected truncation message for >3 rules")
	}
}

func TestRenderSummary_PurposeTruncation(t *testing.T) {
	longPurpose := strings.Repeat("x", 120)
	e := &ExplainOutput{
		Project: &ProjectConfig{
			Name:    "trunctest",
			Type:    "cli",
			Purpose: longPurpose,
		},
	}
	result := e.RenderSummary()

	if !strings.Contains(result, "...") {
		t.Error("expected purpose to be truncated with '...'")
	}
}

// --- RenderProject tests ---

func TestRenderProject_Full(t *testing.T) {
	e := &ExplainOutput{
		Project: &ProjectConfig{
			Name:       "my-project",
			Type:       "cli",
			Purpose:    "Build great things",
			Repository: "https://github.com/test/repo",
			Architecture: ArchitectureConfig{
				Overview: "Layered architecture",
				KeyPaths: map[string]string{
					"commands": "cmd/",
				},
				DataFlow: "Input -> Process -> Output",
				KeyFiles: []KeyFileConfig{
					{Path: "main.go", Purpose: "entry point", Tracked: true},
					{Path: ".env", Purpose: "secrets", Tracked: false},
				},
			},
			Languages: LanguagesConfig{
				Primary: "go",
				Version: "1.22",
			},
			Dependencies: DependenciesConfig{
				External: []ExternalDepConfig{
					{Name: "cobra", Purpose: "CLI framework"},
				},
			},
			Integrations: []IntegrationConfig{
				{Name: "GitHub", Type: "vcs", Interaction: "Push and PR"},
			},
		},
	}

	result := e.RenderProject()

	checks := []string{
		"# my-project",
		"Type: cli",
		"Repository: https://github.com/test/repo",
		"## Purpose",
		"Build great things",
		"## Architecture",
		"Layered architecture",
		"### Key Paths",
		"commands",
		"cmd/",
		"### Data Flow",
		"Input -> Process -> Output",
		"### Key Files",
		"`main.go`",
		"entry point",
		"`.env` (untracked)",
		"## Languages",
		"Primary: go 1.22",
		"## Dependencies",
		"**cobra**: CLI framework",
		"## Integrations",
		"### GitHub (vcs)",
		"Push and PR",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("RenderProject missing %q", check)
		}
	}
}

func TestRenderProject_Nil(t *testing.T) {
	e := &ExplainOutput{}
	result := e.RenderProject()

	if !strings.Contains(result, "No project configuration found") {
		t.Error("expected 'No project configuration found' when nil")
	}
	if !strings.Contains(result, "hint:") {
		t.Error("expected hint when project is nil")
	}
}

// --- RenderTools tests ---

func TestRenderTools_Full(t *testing.T) {
	e := &ExplainOutput{
		Tools: &ToolsConfig{
			Build: map[string]ToolCommand{
				"default": {Command: "make build", Notes: "Produces binary"},
			},
			Test: map[string]ToolCommand{
				"default": {Command: "go test ./...", Output: "test results"},
				"single":  {Command: "go test -run TestName", Example: "go test -run TestFoo ./pkg/..."},
			},
			Lint: map[string]ToolCommand{
				"default": {Command: "golangci-lint run"},
			},
		},
	}

	result := e.RenderTools()

	checks := []string{
		"# Tools",
		"## Build",
		"make build",
		"Notes: Produces binary",
		"## Test",
		"go test ./...",
		"Output: test results",
		"go test -run TestName",
		"Example: `go test -run TestFoo ./pkg/...`",
		"## Lint",
		"golangci-lint run",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("RenderTools missing %q", check)
		}
	}
}

func TestRenderTools_Nil(t *testing.T) {
	e := &ExplainOutput{}
	result := e.RenderTools()

	if !strings.Contains(result, "No tools configuration found") {
		t.Error("expected 'No tools configuration found' when nil")
	}
}

// --- RenderGuidelines tests ---

func TestRenderGuidelines_Full(t *testing.T) {
	e := &ExplainOutput{
		Guidelines: &GuidelinesConfig{
			Rules:      []string{"always test", "no globals"},
			Avoid:      []string{"magic numbers", "deep nesting"},
			Principles: []string{"KISS", "YAGNI"},
			Structure: map[string]string{
				"packages": "one per domain concept",
			},
			Commits: map[string]string{
				"format": "conventional commits",
			},
		},
	}

	result := e.RenderGuidelines()

	checks := []string{
		"# Guidelines",
		"## Rules",
		"- always test",
		"- no globals",
		"## Avoid",
		"- magic numbers",
		"- deep nesting",
		"## Design Principles",
		"- KISS",
		"- YAGNI",
		"## Code Structure",
		"packages",
		"one per domain concept",
		"## Commits",
		"**format**: conventional commits",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("RenderGuidelines missing %q", check)
		}
	}
}

func TestRenderGuidelines_Nil(t *testing.T) {
	e := &ExplainOutput{}
	result := e.RenderGuidelines()

	if !strings.Contains(result, "No guidelines configuration found") {
		t.Error("expected 'No guidelines configuration found' when nil")
	}
}

func TestRenderGuidelines_NamingAndErrors(t *testing.T) {
	e := &ExplainOutput{
		Guidelines: &GuidelinesConfig{
			Naming: map[string]any{
				"functions": "camelCase",
			},
			Errors: map[string]any{
				"pattern": "wrap with context",
			},
			Testing: map[string]any{
				"style": "table-driven",
			},
		},
	}

	result := e.RenderGuidelines()

	checks := []string{
		"## Naming Conventions",
		"functions",
		"camelCase",
		"## Error Handling",
		"pattern",
		"wrap with context",
		"## Testing",
		"style",
		"table-driven",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("RenderGuidelines missing %q", check)
		}
	}
}

// --- RenderSkills tests ---

func TestRenderSkills_WithSkills(t *testing.T) {
	e := &ExplainOutput{
		SkillDefs: []Skill{
			{Name: "deploy", IsLocal: true},
			{Name: "review", IsLocal: false},
		},
	}

	result := e.RenderSkills()

	checks := []string{
		"# Available Skills",
		"## Local Skills",
		"**deploy**",
		"## Global Skills",
		"**review**",
		"checkpoint explain skill deploy",
		"checkpoint explain skill review",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("RenderSkills missing %q", check)
		}
	}
}

func TestRenderSkills_Empty(t *testing.T) {
	e := &ExplainOutput{}
	result := e.RenderSkills()

	if !strings.Contains(result, "No skills configured") {
		t.Error("expected 'No skills configured' when empty")
	}
}

func TestRenderSkills_OnlyLocal(t *testing.T) {
	e := &ExplainOutput{
		SkillDefs: []Skill{
			{Name: "local-skill", IsLocal: true},
		},
	}
	result := e.RenderSkills()

	if !strings.Contains(result, "**local-skill**") {
		t.Error("expected local skill listed")
	}
	if !strings.Contains(result, "(none loaded") {
		t.Error("expected '(none loaded' for global section")
	}
}

// --- RenderSkill tests ---

func TestRenderSkill_Found(t *testing.T) {
	e := &ExplainOutput{
		SkillDefs: []Skill{
			{Name: "deploy", Content: "# Deploy Skill\nDeploy to production."},
			{Name: "review", Content: "# Review Skill\nCode review process."},
		},
	}

	result := e.RenderSkill("deploy")
	if result != "# Deploy Skill\nDeploy to production." {
		t.Errorf("RenderSkill(deploy) = %q, want skill content", result)
	}

	result = e.RenderSkill("review")
	if result != "# Review Skill\nCode review process." {
		t.Errorf("RenderSkill(review) = %q, want skill content", result)
	}
}

func TestRenderSkill_NotFound(t *testing.T) {
	e := &ExplainOutput{
		SkillDefs: []Skill{
			{Name: "deploy", Content: "content"},
		},
	}

	result := e.RenderSkill("nonexistent")
	if !strings.Contains(result, "Skill 'nonexistent' not found") {
		t.Errorf("expected not-found message, got %q", result)
	}
}

func TestRenderSkill_EmptySkills(t *testing.T) {
	e := &ExplainOutput{}
	result := e.RenderSkill("anything")
	if !strings.Contains(result, "not found") {
		t.Errorf("expected not-found message, got %q", result)
	}
}

// --- RenderLearnings tests ---

func TestRenderLearnings_WithLearnings(t *testing.T) {
	e := &ExplainOutput{
		Learnings: []Learning{
			{Timestamp: "2026-01-15T10:00:00Z", Learning: "first insight"},
			{Timestamp: "2026-01-16T12:30:00Z", Learning: "second insight"},
		},
	}

	result := e.RenderLearnings()

	checks := []string{
		"# Captured Learnings",
		"Total: 2 learnings",
		"[2026-01-15] first insight",
		"[2026-01-16] second insight",
		"checkpoint learn",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("RenderLearnings missing %q", check)
		}
	}

	// Verify reverse order: second insight should appear before first
	idxSecond := strings.Index(result, "second insight")
	idxFirst := strings.Index(result, "first insight")
	if idxSecond > idxFirst {
		t.Error("expected learnings in reverse chronological order")
	}
}

func TestRenderLearnings_Empty(t *testing.T) {
	e := &ExplainOutput{}
	result := e.RenderLearnings()

	if !strings.Contains(result, "No learnings captured yet") {
		t.Error("expected 'No learnings captured yet' when empty")
	}
	if !strings.Contains(result, "checkpoint learn") {
		t.Error("expected hint about checkpoint learn command")
	}
}

func TestRenderLearnings_ShortTimestamp(t *testing.T) {
	e := &ExplainOutput{
		Learnings: []Learning{
			{Timestamp: "2026-01", Learning: "short ts"},
		},
	}
	result := e.RenderLearnings()
	// Timestamp shorter than 10 chars should be used as-is
	if !strings.Contains(result, "[2026-01]") {
		t.Errorf("expected short timestamp to be used as-is, got %q", result)
	}
}

// --- RenderFull tests ---

func TestRenderFull_ContainsAllSections(t *testing.T) {
	e := &ExplainOutput{
		Project: &ProjectConfig{
			Name:    "full-test",
			Type:    "service",
			Purpose: "Full render test",
		},
		Tools: &ToolsConfig{
			Build: map[string]ToolCommand{
				"default": {Command: "make build"},
			},
		},
		Guidelines: &GuidelinesConfig{
			Rules: []string{"test rule"},
		},
		SkillDefs: []Skill{
			{Name: "deploy", Content: "Deploy content here", IsLocal: true},
		},
		Learnings: []Learning{
			{Timestamp: "2026-03-01T00:00:00Z", Learning: "a learning"},
		},
	}

	result := e.RenderFull()

	checks := []string{
		// From RenderProject
		"# full-test",
		"Full render test",
		// From RenderTools
		"# Tools",
		"make build",
		// From RenderGuidelines
		"# Guidelines",
		"test rule",
		// From RenderSkills
		"# Available Skills",
		"deploy",
		// Skill content included
		"## Skill: deploy",
		"Deploy content here",
		// Learnings
		"# Captured Learnings",
		"a learning",
		// Section separators
		"---",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("RenderFull missing %q", check)
		}
	}
}

func TestRenderFull_NoLearnings(t *testing.T) {
	e := &ExplainOutput{
		Project: &ProjectConfig{
			Name: "no-learn",
			Type: "lib",
		},
	}

	result := e.RenderFull()

	// Should NOT contain learnings section when empty
	if strings.Contains(result, "# Captured Learnings") {
		t.Error("RenderFull should not include learnings section when empty")
	}
}

func TestRenderFull_NoSkillContent(t *testing.T) {
	e := &ExplainOutput{
		Project: &ProjectConfig{
			Name: "no-skills",
			Type: "lib",
		},
	}

	result := e.RenderFull()

	if strings.Contains(result, "## Skill:") {
		t.Error("RenderFull should not include skill content when no skills")
	}
	// But should still include the skills listing section
	if !strings.Contains(result, "No skills configured") {
		t.Error("expected 'No skills configured' in full render")
	}
}

// --- loadLearnings tests ---

func TestLoadLearnings_MultiDoc(t *testing.T) {
	data := []byte(`timestamp: "2026-01-01"
learning: "insight one"
---
timestamp: "2026-01-02"
learning: "insight two"
---
timestamp: "2026-01-03"
learning: "insight three"
`)
	learnings := loadLearnings(data)
	if len(learnings) != 3 {
		t.Fatalf("expected 3 learnings, got %d", len(learnings))
	}
	if learnings[0].Learning != "insight one" {
		t.Errorf("learnings[0] = %q, want 'insight one'", learnings[0].Learning)
	}
	if learnings[2].Learning != "insight three" {
		t.Errorf("learnings[2] = %q, want 'insight three'", learnings[2].Learning)
	}
}

func TestLoadLearnings_EmptyLearning(t *testing.T) {
	data := []byte(`timestamp: "2026-01-01"
learning: ""
---
timestamp: "2026-01-02"
learning: "valid"
`)
	learnings := loadLearnings(data)
	if len(learnings) != 1 {
		t.Fatalf("expected 1 learning (empty ones skipped), got %d", len(learnings))
	}
	if learnings[0].Learning != "valid" {
		t.Errorf("learnings[0] = %q, want 'valid'", learnings[0].Learning)
	}
}

func TestLoadLearnings_Empty(t *testing.T) {
	learnings := loadLearnings([]byte(""))
	if len(learnings) != 0 {
		t.Errorf("expected 0 learnings for empty input, got %d", len(learnings))
	}
}
