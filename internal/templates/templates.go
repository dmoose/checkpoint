package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Template represents a project template
type Template struct {
	Name        string
	Description string
	Path        string // Empty for embedded templates
	IsEmbedded  bool
	Files       map[string]string // filename -> content for embedded
}

// TemplateInfo is a summary for listing
type TemplateInfo struct {
	Name        string
	Description string
	Source      string // "embedded" or path
}

// ListTemplates returns all available templates (embedded + global)
func ListTemplates() ([]TemplateInfo, error) {
	var templates []TemplateInfo

	// Add embedded templates
	for name, tmpl := range embeddedTemplates {
		templates = append(templates, TemplateInfo{
			Name:        name,
			Description: tmpl.Description,
			Source:      "embedded",
		})
	}

	// Add global templates from ~/.config/checkpoint/templates/
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalDir := filepath.Join(homeDir, ".config", "checkpoint", "templates")
		if entries, err := os.ReadDir(globalDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					name := entry.Name()
					// Skip if already have embedded with same name
					if _, exists := embeddedTemplates[name]; exists {
						continue
					}
					desc := loadTemplateDescription(filepath.Join(globalDir, name))
					templates = append(templates, TemplateInfo{
						Name:        name,
						Description: desc,
						Source:      globalDir,
					})
				}
			}
		}
	}

	// Sort by name
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}

// loadTemplateDescription reads description from template.yaml or first line of project.yml
func loadTemplateDescription(templateDir string) string {
	// Try template.yaml first
	templateYaml := filepath.Join(templateDir, "template.yaml")
	if data, err := os.ReadFile(templateYaml); err == nil {
		// Simple extraction - look for description: line
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "description:") {
				return strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}
	}

	// Fall back to project.yml purpose field
	projectYml := filepath.Join(templateDir, "project.yml")
	if data, err := os.ReadFile(projectYml); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "purpose:") {
				purpose := strings.TrimSpace(strings.TrimPrefix(line, "purpose:"))
				if len(purpose) > 60 {
					purpose = purpose[:57] + "..."
				}
				return purpose
			}
		}
	}

	return "(no description)"
}

// GetTemplate returns a template by name
func GetTemplate(name string) (*Template, error) {
	// Check embedded first
	if tmpl, exists := embeddedTemplates[name]; exists {
		return &tmpl, nil
	}

	// Check global templates
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get home directory: %w", err)
	}

	templateDir := filepath.Join(homeDir, ".config", "checkpoint", "templates", name)
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	// Load template files
	files := make(map[string]string)
	templateFiles := []string{"project.yml", "tools.yml", "guidelines.yml", "skills.yml"}
	for _, filename := range templateFiles {
		path := filepath.Join(templateDir, filename)
		if data, err := os.ReadFile(path); err == nil {
			files[filename] = string(data)
		}
	}

	return &Template{
		Name:        name,
		Description: loadTemplateDescription(templateDir),
		Path:        templateDir,
		IsEmbedded:  false,
		Files:       files,
	}, nil
}

// ApplyTemplate writes template files to the project's .checkpoint directory
func ApplyTemplate(tmpl *Template, projectPath, projectName string) error {
	checkpointDir := filepath.Join(projectPath, ".checkpoint")

	// Ensure directory exists
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		return fmt.Errorf("create .checkpoint directory: %w", err)
	}

	// Write each template file, substituting variables
	for filename, content := range tmpl.Files {
		// Substitute template variables
		content = substituteVariables(content, projectName, projectPath)

		targetPath := filepath.Join(checkpointDir, filename)

		// Don't overwrite existing files
		if _, err := os.Stat(targetPath); err == nil {
			continue
		}

		if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", filename, err)
		}
	}

	return nil
}

// substituteVariables replaces template placeholders
func substituteVariables(content, projectName, projectPath string) string {
	content = strings.ReplaceAll(content, "{{project_name}}", projectName)
	content = strings.ReplaceAll(content, "{{project_path}}", projectPath)
	return content
}

// embeddedTemplates contains built-in templates
var embeddedTemplates = map[string]Template{
	"go-cli": {
		Name:        "go-cli",
		Description: "Go command-line application",
		IsEmbedded:  true,
		Files: map[string]string{
			"project.yml":    goCliProjectYml,
			"tools.yml":      goCliToolsYml,
			"guidelines.yml": goCliGuidelinesYml,
			"skills.yml":     goCliSkillsYml,
		},
	},
	"go-lib": {
		Name:        "go-lib",
		Description: "Go library package",
		IsEmbedded:  true,
		Files: map[string]string{
			"project.yml":    goLibProjectYml,
			"tools.yml":      goLibToolsYml,
			"guidelines.yml": goCliGuidelinesYml, // Reuse Go guidelines
			"skills.yml":     goCliSkillsYml,
		},
	},
	"node-api": {
		Name:        "node-api",
		Description: "Node.js API server",
		IsEmbedded:  true,
		Files: map[string]string{
			"project.yml":    nodeApiProjectYml,
			"tools.yml":      nodeApiToolsYml,
			"guidelines.yml": nodeApiGuidelinesYml,
			"skills.yml":     nodeApiSkillsYml,
		},
	},
	"python-cli": {
		Name:        "python-cli",
		Description: "Python command-line application",
		IsEmbedded:  true,
		Files: map[string]string{
			"project.yml":    pythonCliProjectYml,
			"tools.yml":      pythonCliToolsYml,
			"guidelines.yml": pythonCliGuidelinesYml,
			"skills.yml":     pythonCliSkillsYml,
		},
	},
	"generic": {
		Name:        "generic",
		Description: "Generic project template",
		IsEmbedded:  true,
		Files: map[string]string{
			"project.yml":    genericProjectYml,
			"tools.yml":      genericToolsYml,
			"guidelines.yml": genericGuidelinesYml,
			"skills.yml":     genericSkillsYml,
		},
	},
}
