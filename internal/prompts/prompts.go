package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"
)

// PromptsConfig represents the prompts.yaml configuration file
type PromptsConfig struct {
	SchemaVersion string             `yaml:"schema_version"`
	Variables     map[string]string  `yaml:"variables,omitempty"`
	Prompts       []PromptDefinition `yaml:"prompts"`
}

// PromptDefinition describes a single prompt in the library
type PromptDefinition struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Category    string   `yaml:"category"`
	Description string   `yaml:"description"`
	File        string   `yaml:"file"`
	Variables   []string `yaml:"variables,omitempty"`
}

// PromptInfo contains basic information about a prompt for listing
type PromptInfo struct {
	ID          string
	Name        string
	Category    string
	Description string
}

// Prompt represents a loaded prompt with its definition and template content
type Prompt struct {
	Definition PromptDefinition
	Template   string
}

// LoadPromptsConfig loads and parses the prompts.yaml configuration file
func LoadPromptsConfig(promptsDir string) (*PromptsConfig, error) {
	configPath := filepath.Join(promptsDir, "prompts.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompts.yaml: %w", err)
	}

	var config PromptsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse prompts.yaml: %w", err)
	}

	return &config, nil
}

// ListPrompts returns a list of all prompts from the config
func ListPrompts(config *PromptsConfig) []PromptInfo {
	prompts := make([]PromptInfo, 0, len(config.Prompts))

	for _, p := range config.Prompts {
		prompts = append(prompts, PromptInfo{
			ID:          p.ID,
			Name:        p.Name,
			Category:    p.Category,
			Description: p.Description,
		})
	}

	// Sort by category, then by name
	sort.Slice(prompts, func(i, j int) bool {
		if prompts[i].Category != prompts[j].Category {
			return prompts[i].Category < prompts[j].Category
		}
		return prompts[i].Name < prompts[j].Name
	})

	return prompts
}

// GetPrompt retrieves a specific prompt by ID
func GetPrompt(config *PromptsConfig, promptsDir string, id string) (*Prompt, error) {
	// Find the prompt definition
	var def *PromptDefinition
	for _, p := range config.Prompts {
		if p.ID == id {
			def = &p
			break
		}
	}

	if def == nil {
		return nil, fmt.Errorf("prompt '%s' not found", id)
	}

	// Load the template file
	template, err := LoadPromptTemplate(promptsDir, def.File)
	if err != nil {
		return nil, err
	}

	return &Prompt{
		Definition: *def,
		Template:   template,
	}, nil
}

// LoadPromptTemplate loads a prompt template file from the prompts directory
func LoadPromptTemplate(promptsDir string, filename string) (string, error) {
	templatePath := filepath.Join(promptsDir, filename)

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", filename, err)
	}

	return string(data), nil
}

// SubstituteVariables performs simple variable substitution in the template
// Variables are in the format {{variable_name}} and are replaced with values from the vars map
// Unknown variables are replaced with empty strings
func SubstituteVariables(template string, vars map[string]string) string {
	// Match {{variable_name}} where variable_name is [a-z_][a-z0-9_]*
	re := regexp.MustCompile(`\{\{([a-z_][a-z0-9_]*)\}\}`)

	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := match[2 : len(match)-2]

		// Return value if found, otherwise empty string
		if value, ok := vars[varName]; ok {
			return value
		}
		return ""
	})

	return result
}
