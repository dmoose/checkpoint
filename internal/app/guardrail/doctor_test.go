package guardrail

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dmoose/checkpoint/pkg/config"
)

func TestCheckCheckpointDir_Exists(t *testing.T) {
	tmp := t.TempDir()
	checkpointPath := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointPath, 0755); err != nil {
		t.Fatal(err)
	}

	result := checkCheckpointDir(tmp)
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", result.Status)
	}
	if result.Name != "Checkpoint Directory" {
		t.Errorf("expected name 'Checkpoint Directory', got %q", result.Name)
	}
}

func TestCheckCheckpointDir_Missing(t *testing.T) {
	tmp := t.TempDir()

	result := checkCheckpointDir(tmp)
	if result.Status != "missing" {
		t.Errorf("expected status 'missing', got %q", result.Status)
	}
	if result.Fix == "" {
		t.Error("expected a fix suggestion for missing checkpoint dir")
	}
}

func TestCheckProjectYml_Valid(t *testing.T) {
	tmp := t.TempDir()
	checkpointPath := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointPath, 0755); err != nil {
		t.Fatal(err)
	}

	projectYaml := `schema_version: "1"
name: my-project
purpose: A useful tool
languages:
  primary: go
`
	if err := os.WriteFile(filepath.Join(checkpointPath, config.ExplainProjectYaml), []byte(projectYaml), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkProjectYml(tmp)
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckProjectYml_Missing(t *testing.T) {
	tmp := t.TempDir()

	result := checkProjectYml(tmp)
	if result.Status != "missing" {
		t.Errorf("expected status 'missing', got %q", result.Status)
	}
}

func TestCheckProjectYml_Placeholder(t *testing.T) {
	tmp := t.TempDir()
	checkpointPath := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointPath, 0755); err != nil {
		t.Fatal(err)
	}

	projectYaml := `schema_version: "1"
name: project-name
purpose: "TODO: fill in"
languages:
  primary: ""
`
	if err := os.WriteFile(filepath.Join(checkpointPath, config.ExplainProjectYaml), []byte(projectYaml), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkProjectYml(tmp)
	if result.Status != "warning" {
		t.Errorf("expected status 'warning', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckToolsYml_Valid(t *testing.T) {
	tmp := t.TempDir()
	checkpointPath := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointPath, 0755); err != nil {
		t.Fatal(err)
	}

	toolsYaml := `schema_version: "1"
build:
  binary:
    command: "go build ./..."
test:
  unit:
    command: "go test ./..."
`
	if err := os.WriteFile(filepath.Join(checkpointPath, config.ExplainToolsYaml), []byte(toolsYaml), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkToolsYml(tmp)
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckToolsYml_Missing(t *testing.T) {
	tmp := t.TempDir()

	result := checkToolsYml(tmp)
	if result.Status != "missing" {
		t.Errorf("expected status 'missing', got %q", result.Status)
	}
}

func TestCheckToolsYml_Empty(t *testing.T) {
	tmp := t.TempDir()
	checkpointPath := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointPath, 0755); err != nil {
		t.Fatal(err)
	}

	toolsYaml := `schema_version: "1"
`
	if err := os.WriteFile(filepath.Join(checkpointPath, config.ExplainToolsYaml), []byte(toolsYaml), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkToolsYml(tmp)
	if result.Status != "warning" {
		t.Errorf("expected status 'warning', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckGuidelinesYml_Valid(t *testing.T) {
	tmp := t.TempDir()
	checkpointPath := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointPath, 0755); err != nil {
		t.Fatal(err)
	}

	guidelinesYaml := `schema_version: "1"
rules:
  - "Always validate input"
  - "Use structured logging"
`
	if err := os.WriteFile(filepath.Join(checkpointPath, config.ExplainGuidelinesYaml), []byte(guidelinesYaml), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkGuidelinesYml(tmp)
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckGuidelinesYml_Missing(t *testing.T) {
	tmp := t.TempDir()

	result := checkGuidelinesYml(tmp)
	if result.Status != "missing" {
		t.Errorf("expected status 'missing', got %q", result.Status)
	}
}

func TestCheckGuidelinesYml_Empty(t *testing.T) {
	tmp := t.TempDir()
	checkpointPath := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointPath, 0755); err != nil {
		t.Fatal(err)
	}

	guidelinesYaml := `schema_version: "1"
`
	if err := os.WriteFile(filepath.Join(checkpointPath, config.ExplainGuidelinesYaml), []byte(guidelinesYaml), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkGuidelinesYml(tmp)
	if result.Status != "warning" {
		t.Errorf("expected status 'warning', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckSkills_WithSkills(t *testing.T) {
	tmp := t.TempDir()
	checkpointPath := filepath.Join(tmp, config.CheckpointDir)

	// Create skills.yaml
	if err := os.MkdirAll(checkpointPath, 0755); err != nil {
		t.Fatal(err)
	}
	skillsYaml := `schema_version: "1"
local:
  - my-skill
`
	if err := os.WriteFile(filepath.Join(checkpointPath, config.ExplainSkillsYaml), []byte(skillsYaml), 0644); err != nil {
		t.Fatal(err)
	}

	// Create skills directory with a skill subdirectory
	skillDir := filepath.Join(checkpointPath, config.SkillsDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	result := checkSkills(tmp)
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckSkills_MissingSkillsYaml(t *testing.T) {
	tmp := t.TempDir()

	result := checkSkills(tmp)
	if result.Status != "warning" {
		t.Errorf("expected status 'warning', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckPrompts_WithFiles(t *testing.T) {
	tmp := t.TempDir()
	promptsDir := filepath.Join(tmp, config.CheckpointDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(promptsDir, "review.md"), []byte("review prompt"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkPrompts(tmp)
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckPrompts_MissingDir(t *testing.T) {
	tmp := t.TempDir()

	result := checkPrompts(tmp)
	if result.Status != "missing" {
		t.Errorf("expected status 'missing', got %q", result.Status)
	}
}

func TestCheckExamples_WithFiles(t *testing.T) {
	tmp := t.TempDir()
	examplesDir := filepath.Join(tmp, config.CheckpointDir, "examples")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(examplesDir, "example1.yaml"), []byte("example"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkExamples(tmp)
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %q: %s", result.Status, result.Message)
	}
}

func TestCheckExamples_MissingDir(t *testing.T) {
	tmp := t.TempDir()

	result := checkExamples(tmp)
	if result.Status != "missing" {
		t.Errorf("expected status 'missing', got %q", result.Status)
	}
}
