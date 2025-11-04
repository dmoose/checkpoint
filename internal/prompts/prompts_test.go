package prompts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		expected string
	}{
		{
			name:     "simple substitution",
			template: "Hello {{name}}!",
			vars:     map[string]string{"name": "World"},
			expected: "Hello World!",
		},
		{
			name:     "multiple variables",
			template: "{{greeting}} {{name}}!",
			vars:     map[string]string{"greeting": "Hello", "name": "World"},
			expected: "Hello World!",
		},
		{
			name:     "unknown variable becomes empty",
			template: "Hello {{name}}!",
			vars:     map[string]string{},
			expected: "Hello !",
		},
		{
			name:     "mixed known and unknown",
			template: "{{greeting}} {{name}} {{unknown}}!",
			vars:     map[string]string{"greeting": "Hi", "name": "Alice"},
			expected: "Hi Alice !",
		},
		{
			name:     "underscores in variable names",
			template: "{{project_name}} uses {{primary_language}}",
			vars:     map[string]string{"project_name": "checkpoint", "primary_language": "Go"},
			expected: "checkpoint uses Go",
		},
		{
			name:     "numbers in variable names",
			template: "{{var1}} and {{var2}}",
			vars:     map[string]string{"var1": "first", "var2": "second"},
			expected: "first and second",
		},
		{
			name:     "no variables",
			template: "Plain text without variables",
			vars:     map[string]string{},
			expected: "Plain text without variables",
		},
		{
			name:     "variable at start and end",
			template: "{{start}} middle {{end}}",
			vars:     map[string]string{"start": "BEGIN", "end": "END"},
			expected: "BEGIN middle END",
		},
		{
			name:     "same variable multiple times",
			template: "{{name}} loves {{name}}",
			vars:     map[string]string{"name": "Go"},
			expected: "Go loves Go",
		},
		{
			name:     "empty variable value",
			template: "Hello {{name}}!",
			vars:     map[string]string{"name": ""},
			expected: "Hello !",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubstituteVariables(tt.template, tt.vars)
			if result != tt.expected {
				t.Errorf("SubstituteVariables() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLoadPromptsConfig(t *testing.T) {
	// Create a temporary directory with test prompts.yaml
	tmpDir := t.TempDir()

	validYaml := `schema_version: "1"
variables:
  project_name: "test-project"
  primary_language: "Go"
prompts:
  - id: test-prompt
    name: "Test Prompt"
    category: test
    description: "A test prompt"
    file: test.md
    variables:
      - test_var
`

	promptsYaml := filepath.Join(tmpDir, "prompts.yaml")
	if err := os.WriteFile(promptsYaml, []byte(validYaml), 0644); err != nil {
		t.Fatalf("Failed to create test prompts.yaml: %v", err)
	}

	// Test loading valid config
	config, err := LoadPromptsConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadPromptsConfig() error = %v", err)
	}

	if config.SchemaVersion != "1" {
		t.Errorf("SchemaVersion = %q, want %q", config.SchemaVersion, "1")
	}

	if len(config.Variables) != 2 {
		t.Errorf("len(Variables) = %d, want 2", len(config.Variables))
	}

	if config.Variables["project_name"] != "test-project" {
		t.Errorf("Variables[project_name] = %q, want %q", config.Variables["project_name"], "test-project")
	}

	if len(config.Prompts) != 1 {
		t.Fatalf("len(Prompts) = %d, want 1", len(config.Prompts))
	}

	prompt := config.Prompts[0]
	if prompt.ID != "test-prompt" {
		t.Errorf("Prompt.ID = %q, want %q", prompt.ID, "test-prompt")
	}
	if prompt.Name != "Test Prompt" {
		t.Errorf("Prompt.Name = %q, want %q", prompt.Name, "Test Prompt")
	}
	if prompt.Category != "test" {
		t.Errorf("Prompt.Category = %q, want %q", prompt.Category, "test")
	}
}

func TestLoadPromptsConfigMissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadPromptsConfig(tmpDir)
	if err == nil {
		t.Error("LoadPromptsConfig() expected error for missing file, got nil")
	}
}

func TestListPrompts(t *testing.T) {
	config := &PromptsConfig{
		SchemaVersion: "1",
		Prompts: []PromptDefinition{
			{ID: "prompt1", Name: "First", Category: "cat1", Description: "First prompt"},
			{ID: "prompt2", Name: "Second", Category: "cat2", Description: "Second prompt"},
			{ID: "prompt3", Name: "Third", Category: "cat1", Description: "Third prompt"},
		},
	}

	prompts := ListPrompts(config)

	if len(prompts) != 3 {
		t.Fatalf("len(prompts) = %d, want 3", len(prompts))
	}

	// Check that prompts are sorted by category then name
	// Expected order: cat1/First, cat1/Third, cat2/Second
	if prompts[0].ID != "prompt1" {
		t.Errorf("prompts[0].ID = %q, want %q", prompts[0].ID, "prompt1")
	}
	if prompts[1].ID != "prompt3" {
		t.Errorf("prompts[1].ID = %q, want %q", prompts[1].ID, "prompt3")
	}
	if prompts[2].ID != "prompt2" {
		t.Errorf("prompts[2].ID = %q, want %q", prompts[2].ID, "prompt2")
	}
}

func TestGetPrompt(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config
	config := &PromptsConfig{
		SchemaVersion: "1",
		Prompts: []PromptDefinition{
			{ID: "test-prompt", Name: "Test", Category: "test", Description: "Test", File: "test.md"},
		},
	}

	// Create template file
	templateContent := "# Test Prompt\n\nHello {{name}}!"
	templatePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Test getting existing prompt
	prompt, err := GetPrompt(config, tmpDir, "test-prompt")
	if err != nil {
		t.Fatalf("GetPrompt() error = %v", err)
	}

	if prompt.Definition.ID != "test-prompt" {
		t.Errorf("Prompt.Definition.ID = %q, want %q", prompt.Definition.ID, "test-prompt")
	}

	if prompt.Template != templateContent {
		t.Errorf("Prompt.Template = %q, want %q", prompt.Template, templateContent)
	}
}

func TestGetPromptNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	config := &PromptsConfig{
		SchemaVersion: "1",
		Prompts:       []PromptDefinition{},
	}

	_, err := GetPrompt(config, tmpDir, "nonexistent")
	if err == nil {
		t.Error("GetPrompt() expected error for nonexistent prompt, got nil")
	}
}

func TestLoadPromptTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	content := "Test template content"
	templatePath := filepath.Join(tmpDir, "template.md")
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	result, err := LoadPromptTemplate(tmpDir, "template.md")
	if err != nil {
		t.Fatalf("LoadPromptTemplate() error = %v", err)
	}

	if result != content {
		t.Errorf("LoadPromptTemplate() = %q, want %q", result, content)
	}
}

func TestLoadPromptTemplateMissing(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadPromptTemplate(tmpDir, "nonexistent.md")
	if err == nil {
		t.Error("LoadPromptTemplate() expected error for missing file, got nil")
	}
}
