package guardrail

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmoose/checkpoint/internal/explain"
	"github.com/dmoose/checkpoint/pkg/config"
	"gopkg.in/yaml.v3"
)

func TestAddGuideline(t *testing.T) {
	tmp := t.TempDir()
	checkpointDir := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write empty guidelines file
	guidelinesPath := filepath.Join(checkpointDir, config.ExplainGuidelinesYaml)
	if err := os.WriteFile(guidelinesPath, []byte("schema_version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Add a guideline
	err := addGuideline(checkpointDir, "Always validate input")
	if err != nil {
		t.Fatalf("addGuideline failed: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(guidelinesPath)
	if err != nil {
		t.Fatal(err)
	}

	var guidelines explain.GuidelinesConfig
	if err := yaml.Unmarshal(data, &guidelines); err != nil {
		t.Fatalf("failed to unmarshal guidelines: %v", err)
	}

	if len(guidelines.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(guidelines.Rules))
	}
	if guidelines.Rules[0] != "Always validate input" {
		t.Errorf("expected rule 'Always validate input', got %q", guidelines.Rules[0])
	}
}

func TestAddGuideline_Duplicate(t *testing.T) {
	tmp := t.TempDir()
	checkpointDir := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		t.Fatal(err)
	}

	guidelinesPath := filepath.Join(checkpointDir, config.ExplainGuidelinesYaml)
	if err := os.WriteFile(guidelinesPath, []byte("schema_version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Add same guideline twice
	if err := addGuideline(checkpointDir, "No globals"); err != nil {
		t.Fatal(err)
	}
	if err := addGuideline(checkpointDir, "No globals"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(guidelinesPath)
	if err != nil {
		t.Fatal(err)
	}

	var guidelines explain.GuidelinesConfig
	if err := yaml.Unmarshal(data, &guidelines); err != nil {
		t.Fatal(err)
	}

	if len(guidelines.Rules) != 1 {
		t.Errorf("expected 1 rule after duplicate add, got %d", len(guidelines.Rules))
	}
}

func TestAddAvoid(t *testing.T) {
	tmp := t.TempDir()
	checkpointDir := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		t.Fatal(err)
	}

	guidelinesPath := filepath.Join(checkpointDir, config.ExplainGuidelinesYaml)
	if err := os.WriteFile(guidelinesPath, []byte("schema_version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := addAvoid(checkpointDir, "Don't use global mutable state")
	if err != nil {
		t.Fatalf("addAvoid failed: %v", err)
	}

	data, err := os.ReadFile(guidelinesPath)
	if err != nil {
		t.Fatal(err)
	}

	var guidelines explain.GuidelinesConfig
	if err := yaml.Unmarshal(data, &guidelines); err != nil {
		t.Fatal(err)
	}

	if len(guidelines.Avoid) != 1 {
		t.Fatalf("expected 1 avoid entry, got %d", len(guidelines.Avoid))
	}
	if guidelines.Avoid[0] != "Don't use global mutable state" {
		t.Errorf("expected avoid entry 'Don't use global mutable state', got %q", guidelines.Avoid[0])
	}
}

func TestAddAvoid_Duplicate(t *testing.T) {
	tmp := t.TempDir()
	checkpointDir := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		t.Fatal(err)
	}

	guidelinesPath := filepath.Join(checkpointDir, config.ExplainGuidelinesYaml)
	if err := os.WriteFile(guidelinesPath, []byte("schema_version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := addAvoid(checkpointDir, "No panics"); err != nil {
		t.Fatal(err)
	}
	if err := addAvoid(checkpointDir, "No panics"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(guidelinesPath)
	if err != nil {
		t.Fatal(err)
	}

	var guidelines explain.GuidelinesConfig
	if err := yaml.Unmarshal(data, &guidelines); err != nil {
		t.Fatal(err)
	}

	if len(guidelines.Avoid) != 1 {
		t.Errorf("expected 1 avoid entry after duplicate add, got %d", len(guidelines.Avoid))
	}
}

func TestAddLearning(t *testing.T) {
	tmp := t.TempDir()
	checkpointDir := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := addLearning(checkpointDir, "YAML multi-doc requires --- separator")
	if err != nil {
		t.Fatalf("addLearning failed: %v", err)
	}

	learningsPath := filepath.Join(checkpointDir, "learnings.yaml")
	data, err := os.ReadFile(learningsPath)
	if err != nil {
		t.Fatalf("failed to read learnings file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "YAML multi-doc requires --- separator") {
		t.Error("learnings file does not contain the added learning")
	}
	if !strings.Contains(content, "timestamp:") {
		t.Error("learnings file does not contain a timestamp")
	}
	if !strings.Contains(content, "---") {
		t.Error("learnings file does not contain document separator")
	}

	// Add a second learning and verify append-only behavior
	err = addLearning(checkpointDir, "Second learning entry")
	if err != nil {
		t.Fatalf("second addLearning failed: %v", err)
	}

	data, err = os.ReadFile(learningsPath)
	if err != nil {
		t.Fatal(err)
	}

	content = string(data)
	if !strings.Contains(content, "YAML multi-doc requires --- separator") {
		t.Error("first learning missing after append")
	}
	if !strings.Contains(content, "Second learning entry") {
		t.Error("second learning not appended")
	}
}

func TestAddTool(t *testing.T) {
	tmp := t.TempDir()
	checkpointDir := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		t.Fatal(err)
	}

	toolsPath := filepath.Join(checkpointDir, config.ExplainToolsYaml)
	if err := os.WriteFile(toolsPath, []byte("schema_version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := addTool(checkpointDir, "race", "make test-race")
	if err != nil {
		t.Fatalf("addTool failed: %v", err)
	}

	data, err := os.ReadFile(toolsPath)
	if err != nil {
		t.Fatal(err)
	}

	var tools explain.ToolsConfig
	if err := yaml.Unmarshal(data, &tools); err != nil {
		t.Fatalf("failed to unmarshal tools: %v", err)
	}

	if tools.Maintenance == nil {
		t.Fatal("expected maintenance section to exist")
	}

	tool, ok := tools.Maintenance["race"]
	if !ok {
		t.Fatal("expected tool 'race' to exist in maintenance")
	}
	if tool.Command != "make test-race" {
		t.Errorf("expected command 'make test-race', got %q", tool.Command)
	}
}

func TestAddTool_AutoName(t *testing.T) {
	tmp := t.TempDir()
	checkpointDir := filepath.Join(tmp, config.CheckpointDir)
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		t.Fatal(err)
	}

	toolsPath := filepath.Join(checkpointDir, config.ExplainToolsYaml)
	if err := os.WriteFile(toolsPath, []byte("schema_version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// When name is empty, addTool should extract name from command
	err := addTool(checkpointDir, "", "go vet ./...")
	if err != nil {
		t.Fatalf("addTool with auto-name failed: %v", err)
	}

	data, err := os.ReadFile(toolsPath)
	if err != nil {
		t.Fatal(err)
	}

	var tools explain.ToolsConfig
	if err := yaml.Unmarshal(data, &tools); err != nil {
		t.Fatal(err)
	}

	// Last word of "go vet ./..." is "./..."
	if tools.Maintenance == nil {
		t.Fatal("expected maintenance section to exist")
	}
	if len(tools.Maintenance) != 1 {
		t.Errorf("expected 1 maintenance tool, got %d", len(tools.Maintenance))
	}
}
