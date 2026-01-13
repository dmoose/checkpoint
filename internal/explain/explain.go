package explain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/checkpoint/pkg/config"

	"gopkg.in/yaml.v3"
)

// ExplainOutput holds all the context for explain command
type ExplainOutput struct {
	Project     *ProjectConfig
	Tools       *ToolsConfig
	Guidelines  *GuidelinesConfig
	Skills      *SkillsConfig
	SkillDefs   []Skill
	Learnings   []Learning
	ProjectPath string
}

// LoadExplainContext loads all explain-related files from a project
func LoadExplainContext(projectPath string) (*ExplainOutput, error) {
	checkpointDir := filepath.Join(projectPath, config.CheckpointDir)

	output := &ExplainOutput{
		ProjectPath: projectPath,
	}

	// Load project.yaml (with .yml fallback)
	projectYaml := findYamlFile(checkpointDir, config.ExplainProjectYaml, config.ExplainProjectYmlLegacy)
	if data, err := os.ReadFile(projectYaml); err == nil {
		var proj ProjectConfig
		if err := yaml.Unmarshal(data, &proj); err == nil {
			output.Project = &proj
		}
	}

	// Load tools.yaml (with .yml fallback)
	toolsYaml := findYamlFile(checkpointDir, config.ExplainToolsYaml, config.ExplainToolsYmlLegacy)
	if data, err := os.ReadFile(toolsYaml); err == nil {
		var tools ToolsConfig
		if err := yaml.Unmarshal(data, &tools); err == nil {
			output.Tools = &tools
		}
	}

	// Load guidelines.yaml (with .yml fallback)
	guidelinesYaml := findYamlFile(checkpointDir, config.ExplainGuidelinesYaml, config.ExplainGuidelinesYmlLegacy)
	if data, err := os.ReadFile(guidelinesYaml); err == nil {
		var guidelines GuidelinesConfig
		if err := yaml.Unmarshal(data, &guidelines); err == nil {
			output.Guidelines = &guidelines
		}
	}

	// Load skills.yaml (with .yml fallback)
	skillsYaml := findYamlFile(checkpointDir, config.ExplainSkillsYaml, config.ExplainSkillsYmlLegacy)
	if data, err := os.ReadFile(skillsYaml); err == nil {
		var skills SkillsConfig
		if err := yaml.Unmarshal(data, &skills); err == nil {
			output.Skills = &skills
		}
	}

	// Load skill definitions
	output.SkillDefs = loadSkills(projectPath, output.Skills)

	// Load learnings.yaml (with .yml fallback)
	learningsYaml := findYamlFile(checkpointDir, "learnings.yaml", "learnings.yml")
	if data, err := os.ReadFile(learningsYaml); err == nil {
		output.Learnings = loadLearnings(data)
	}

	return output, nil
}

// findYamlFile returns the path to a yaml file, checking primary first then legacy
func findYamlFile(dir, primary, legacy string) string {
	primaryPath := filepath.Join(dir, primary)
	if _, err := os.Stat(primaryPath); err == nil {
		return primaryPath
	}
	legacyPath := filepath.Join(dir, legacy)
	if _, err := os.Stat(legacyPath); err == nil {
		return legacyPath
	}
	return primaryPath // Return primary for creation
}

// loadLearnings parses multi-document YAML learnings file
func loadLearnings(data []byte) []Learning {
	var learnings []Learning
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var l Learning
		if err := decoder.Decode(&l); err != nil {
			break
		}
		if l.Learning != "" {
			learnings = append(learnings, l)
		}
	}
	return learnings
}

// loadSkills loads skill.md files from local and global skills directories
func loadSkills(projectPath string, skillsConfig *SkillsConfig) []Skill {
	var skills []Skill

	if skillsConfig == nil {
		return skills
	}

	// Load local skills
	localSkillsDir := filepath.Join(projectPath, config.CheckpointDir, config.SkillsDir)
	for _, skillName := range skillsConfig.Local {
		skillPath := filepath.Join(localSkillsDir, skillName, "skill.md")
		if content, err := os.ReadFile(skillPath); err == nil {
			skills = append(skills, Skill{
				Name:    skillName,
				Path:    skillPath,
				Content: string(content),
				IsLocal: true,
			})
		}
	}

	// Load global skills
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalSkillsDir := filepath.Join(homeDir, config.GlobalConfigDir, config.GlobalSkillsDir)
		for _, skillName := range skillsConfig.Global {
			skillPath := filepath.Join(globalSkillsDir, skillName, "skill.md")
			if content, err := os.ReadFile(skillPath); err == nil {
				skills = append(skills, Skill{
					Name:    skillName,
					Path:    skillPath,
					Content: string(content),
					IsLocal: false,
				})
			}
		}
	}

	return skills
}

// RenderSummary returns the executive summary with option index
func (e *ExplainOutput) RenderSummary() string {
	var sb strings.Builder

	// Header
	if e.Project != nil {
		sb.WriteString(fmt.Sprintf("PROJECT: %s (%s)\n", e.Project.Name, e.Project.Type))
		if e.Project.Purpose != "" {
			purpose := strings.TrimSpace(e.Project.Purpose)
			// Truncate to first line or 100 chars for summary
			if idx := strings.Index(purpose, "\n"); idx > 0 {
				purpose = purpose[:idx]
			}
			if len(purpose) > 100 {
				purpose = purpose[:97] + "..."
			}
			sb.WriteString(fmt.Sprintf("PURPOSE: %s\n", purpose))
		}
	} else {
		sb.WriteString("PROJECT: (not configured)\n")
		sb.WriteString("hint: Run 'checkpoint init' or create .checkpoint/project.yml\n")
	}
	sb.WriteString("\n")

	// Quick start commands
	sb.WriteString("QUICK START:\n")
	if e.Tools != nil {
		if cmd, ok := e.Tools.Build["default"]; ok {
			sb.WriteString(fmt.Sprintf("  build: %s\n", cmd.Command))
		}
		if cmd, ok := e.Tools.Test["default"]; ok {
			sb.WriteString(fmt.Sprintf("  test:  %s\n", cmd.Command))
		}
		if cmd, ok := e.Tools.Check["default"]; ok {
			sb.WriteString(fmt.Sprintf("  check: %s\n", cmd.Command))
		} else if cmd, ok := e.Tools.Lint["default"]; ok {
			sb.WriteString(fmt.Sprintf("  lint:  %s\n", cmd.Command))
		}
	} else {
		sb.WriteString("  (no tools configured - create .checkpoint/tools.yml)\n")
	}
	sb.WriteString("\n")

	// Key paths
	if e.Project != nil && len(e.Project.Architecture.KeyPaths) > 0 {
		sb.WriteString("KEY PATHS:\n")
		for name, path := range e.Project.Architecture.KeyPaths {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", name, path))
		}
		sb.WriteString("\n")
	}

	// Skills available
	if len(e.SkillDefs) > 0 {
		sb.WriteString("SKILLS AVAILABLE:\n")
		var skillNames []string
		for _, s := range e.SkillDefs {
			skillNames = append(skillNames, s.Name)
		}
		sb.WriteString(fmt.Sprintf("  %s\n", strings.Join(skillNames, ", ")))
		sb.WriteString("\n")
	}

	// Guidelines summary
	if e.Guidelines != nil && len(e.Guidelines.Rules) > 0 {
		sb.WriteString("KEY RULES:\n")
		for i, rule := range e.Guidelines.Rules {
			if i >= 3 {
				sb.WriteString(fmt.Sprintf("  ... and %d more (see: checkpoint explain guidelines)\n", len(e.Guidelines.Rules)-3))
				break
			}
			sb.WriteString(fmt.Sprintf("  - %s\n", rule))
		}
		sb.WriteString("\n")
	}

	// Learnings summary if any
	if len(e.Learnings) > 0 {
		sb.WriteString(fmt.Sprintf("LEARNINGS: %d captured insights\n", len(e.Learnings)))
		// Show most recent
		if len(e.Learnings) > 0 {
			recent := e.Learnings[len(e.Learnings)-1]
			learning := recent.Learning
			if len(learning) > 60 {
				learning = learning[:57] + "..."
			}
			sb.WriteString(fmt.Sprintf("  latest: %s\n", learning))
		}
		sb.WriteString("\n")
	}

	// Option index
	sb.WriteString("FOR MORE DETAIL:\n")
	sb.WriteString("  checkpoint explain project      - Architecture, data flow, key files\n")
	sb.WriteString("  checkpoint explain tools        - All build, test, lint commands\n")
	sb.WriteString("  checkpoint explain guidelines   - Conventions, rules, anti-patterns\n")
	sb.WriteString("  checkpoint explain skills       - Available skills and how to use them\n")
	sb.WriteString("  checkpoint explain skill <name> - Specific skill details\n")
	sb.WriteString("  checkpoint explain learnings    - Captured insights and lessons\n")
	sb.WriteString("  checkpoint explain history      - Recent checkpoints and patterns\n")
	sb.WriteString("  checkpoint explain --full       - Complete context dump\n")
	sb.WriteString("\n")
	sb.WriteString("FLAGS:\n")
	sb.WriteString("  --md    Output as markdown\n")
	sb.WriteString("  --json  Output as JSON (machine-readable)\n")

	return sb.String()
}

// RenderProject returns detailed project information
func (e *ExplainOutput) RenderProject() string {
	if e.Project == nil {
		return "No project configuration found.\nhint: Create .checkpoint/project.yml\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", e.Project.Name))
	sb.WriteString(fmt.Sprintf("Type: %s\n", e.Project.Type))
	if e.Project.Repository != "" {
		sb.WriteString(fmt.Sprintf("Repository: %s\n", e.Project.Repository))
	}
	sb.WriteString("\n")

	sb.WriteString("## Purpose\n\n")
	sb.WriteString(e.Project.Purpose)
	sb.WriteString("\n\n")

	if e.Project.Architecture.Overview != "" {
		sb.WriteString("## Architecture\n\n")
		sb.WriteString(e.Project.Architecture.Overview)
		sb.WriteString("\n")
	}

	if len(e.Project.Architecture.KeyPaths) > 0 {
		sb.WriteString("\n### Key Paths\n\n")
		for name, path := range e.Project.Architecture.KeyPaths {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", name, path))
		}
	}

	if e.Project.Architecture.DataFlow != "" {
		sb.WriteString("\n### Data Flow\n\n")
		sb.WriteString(e.Project.Architecture.DataFlow)
		sb.WriteString("\n")
	}

	if len(e.Project.Architecture.KeyFiles) > 0 {
		sb.WriteString("\n### Key Files\n\n")
		for _, f := range e.Project.Architecture.KeyFiles {
			tracked := ""
			if !f.Tracked {
				tracked = " (untracked)"
			}
			sb.WriteString(fmt.Sprintf("- `%s`%s - %s\n", f.Path, tracked, f.Purpose))
		}
	}

	if e.Project.Languages.Primary != "" {
		sb.WriteString("\n## Languages\n\n")
		sb.WriteString(fmt.Sprintf("Primary: %s", e.Project.Languages.Primary))
		if e.Project.Languages.Version != "" {
			sb.WriteString(fmt.Sprintf(" %s", e.Project.Languages.Version))
		}
		sb.WriteString("\n")
	}

	if len(e.Project.Dependencies.External) > 0 {
		sb.WriteString("\n## Dependencies\n\n")
		for _, dep := range e.Project.Dependencies.External {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", dep.Name, dep.Purpose))
		}
	}

	if len(e.Project.Integrations) > 0 {
		sb.WriteString("\n## Integrations\n\n")
		for _, integ := range e.Project.Integrations {
			sb.WriteString(fmt.Sprintf("### %s (%s)\n\n", integ.Name, integ.Type))
			if integ.Interaction != "" {
				sb.WriteString(integ.Interaction)
				sb.WriteString("\n\n")
			}
		}
	}

	return sb.String()
}

// RenderTools returns detailed tools information
func (e *ExplainOutput) RenderTools() string {
	if e.Tools == nil {
		return "No tools configuration found.\nhint: Create .checkpoint/tools.yml\n"
	}

	var sb strings.Builder
	sb.WriteString("# Tools\n\n")

	renderSection := func(title string, cmds map[string]ToolCommand) {
		if len(cmds) == 0 {
			return
		}
		sb.WriteString(fmt.Sprintf("## %s\n\n", title))
		for name, cmd := range cmds {
			sb.WriteString(fmt.Sprintf("### %s\n", name))
			sb.WriteString(fmt.Sprintf("```\n%s\n```\n", cmd.Command))
			if cmd.Output != "" {
				sb.WriteString(fmt.Sprintf("Output: %s\n", cmd.Output))
			}
			if cmd.Notes != "" {
				sb.WriteString(fmt.Sprintf("Notes: %s\n", cmd.Notes))
			}
			if cmd.Example != "" {
				sb.WriteString(fmt.Sprintf("Example: `%s`\n", cmd.Example))
			}
			sb.WriteString("\n")
		}
	}

	renderSection("Build", e.Tools.Build)
	renderSection("Test", e.Tools.Test)
	renderSection("Lint", e.Tools.Lint)
	renderSection("Check", e.Tools.Check)
	renderSection("Run", e.Tools.Run)
	renderSection("Checkpoint", e.Tools.Checkpoint)
	renderSection("Maintenance", e.Tools.Maintenance)

	return sb.String()
}

// RenderGuidelines returns detailed guidelines information
func (e *ExplainOutput) RenderGuidelines() string {
	if e.Guidelines == nil {
		return "No guidelines configuration found.\nhint: Create .checkpoint/guidelines.yml\n"
	}

	var sb strings.Builder
	sb.WriteString("# Guidelines\n\n")

	if len(e.Guidelines.Naming) > 0 {
		sb.WriteString("## Naming Conventions\n\n")
		for name, rule := range e.Guidelines.Naming {
			sb.WriteString(fmt.Sprintf("### %s\n", name))
			renderFlexibleValue(&sb, rule, "")
			sb.WriteString("\n")
		}
	}

	if len(e.Guidelines.Structure) > 0 {
		sb.WriteString("## Code Structure\n\n")
		for name, desc := range e.Guidelines.Structure {
			sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", name, desc))
		}
	}

	if len(e.Guidelines.Errors) > 0 {
		sb.WriteString("## Error Handling\n\n")
		for name, val := range e.Guidelines.Errors {
			sb.WriteString(fmt.Sprintf("### %s\n", name))
			renderFlexibleValue(&sb, val, "")
			sb.WriteString("\n")
		}
	}

	if len(e.Guidelines.Testing) > 0 {
		sb.WriteString("## Testing\n\n")
		for name, val := range e.Guidelines.Testing {
			sb.WriteString(fmt.Sprintf("### %s\n", name))
			renderFlexibleValue(&sb, val, "")
			sb.WriteString("\n")
		}
	}

	if len(e.Guidelines.Commits) > 0 {
		sb.WriteString("## Commits\n\n")
		for name, desc := range e.Guidelines.Commits {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", name, desc))
		}
		sb.WriteString("\n")
	}

	if len(e.Guidelines.Rules) > 0 {
		sb.WriteString("## Rules\n\n")
		for _, rule := range e.Guidelines.Rules {
			sb.WriteString(fmt.Sprintf("- %s\n", rule))
		}
		sb.WriteString("\n")
	}

	if len(e.Guidelines.Avoid) > 0 {
		sb.WriteString("## Avoid\n\n")
		for _, item := range e.Guidelines.Avoid {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(e.Guidelines.Principles) > 0 {
		sb.WriteString("## Design Principles\n\n")
		for _, p := range e.Guidelines.Principles {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderFlexibleValue renders interface{} values in a readable way
func renderFlexibleValue(sb *strings.Builder, val interface{}, indent string) {
	switch v := val.(type) {
	case string:
		sb.WriteString(fmt.Sprintf("%s%s\n", indent, v))
	case []interface{}:
		for _, item := range v {
			sb.WriteString(fmt.Sprintf("%s- %v\n", indent, item))
		}
	case map[string]interface{}:
		for key, subval := range v {
			switch sv := subval.(type) {
			case string:
				sb.WriteString(fmt.Sprintf("%s**%s**: %s\n", indent, key, sv))
			case []interface{}:
				sb.WriteString(fmt.Sprintf("%s**%s**:\n", indent, key))
				for _, item := range sv {
					sb.WriteString(fmt.Sprintf("%s  - %v\n", indent, item))
				}
			default:
				sb.WriteString(fmt.Sprintf("%s**%s**: %v\n", indent, key, sv))
			}
		}
	default:
		sb.WriteString(fmt.Sprintf("%s%v\n", indent, v))
	}
}

// RenderSkills returns skills listing
func (e *ExplainOutput) RenderSkills() string {
	var sb strings.Builder
	sb.WriteString("# Available Skills\n\n")

	if len(e.SkillDefs) == 0 {
		sb.WriteString("No skills configured.\n")
		sb.WriteString("hint: Add skills to .checkpoint/skills.yml or create ~/.config/checkpoint/skills/\n")
		return sb.String()
	}

	sb.WriteString("## Local Skills\n\n")
	hasLocal := false
	for _, s := range e.SkillDefs {
		if s.IsLocal {
			hasLocal = true
			sb.WriteString(fmt.Sprintf("- **%s** - `checkpoint explain skill %s`\n", s.Name, s.Name))
		}
	}
	if !hasLocal {
		sb.WriteString("(none)\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## Global Skills\n\n")
	hasGlobal := false
	for _, s := range e.SkillDefs {
		if !s.IsLocal {
			hasGlobal = true
			sb.WriteString(fmt.Sprintf("- **%s** - `checkpoint explain skill %s`\n", s.Name, s.Name))
		}
	}
	if !hasGlobal {
		sb.WriteString("(none loaded - check ~/.config/checkpoint/skills/)\n")
	}
	sb.WriteString("\n")

	sb.WriteString("To view a skill: `checkpoint explain skill <name>`\n")

	return sb.String()
}

// RenderSkill returns a specific skill's content
func (e *ExplainOutput) RenderSkill(name string) string {
	for _, s := range e.SkillDefs {
		if s.Name == name {
			return s.Content
		}
	}
	return fmt.Sprintf("Skill '%s' not found.\nAvailable skills: checkpoint explain skills\n", name)
}

// RenderLearnings returns captured learnings/insights
func (e *ExplainOutput) RenderLearnings() string {
	var sb strings.Builder
	sb.WriteString("# Captured Learnings\n\n")

	if len(e.Learnings) == 0 {
		sb.WriteString("No learnings captured yet.\n")
		sb.WriteString("hint: Use 'checkpoint learn \"your insight\"' to capture learnings\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Total: %d learnings\n\n", len(e.Learnings)))

	// Show most recent first (reverse order)
	for i := len(e.Learnings) - 1; i >= 0; i-- {
		l := e.Learnings[i]
		// Format timestamp nicely
		ts := l.Timestamp
		if len(ts) > 10 {
			ts = ts[:10] // Just the date
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", ts, l.Learning))
	}

	sb.WriteString("\nTo add: checkpoint learn \"your insight\"\n")
	sb.WriteString("To categorize: checkpoint learn \"insight\" --guideline|--avoid|--principle\n")

	return sb.String()
}

// RenderFull returns complete context dump
func (e *ExplainOutput) RenderFull() string {
	var sb strings.Builder

	sb.WriteString(e.RenderProject())
	sb.WriteString("\n---\n\n")
	sb.WriteString(e.RenderTools())
	sb.WriteString("\n---\n\n")
	sb.WriteString(e.RenderGuidelines())
	sb.WriteString("\n---\n\n")
	sb.WriteString(e.RenderSkills())

	// Include skill contents
	for _, s := range e.SkillDefs {
		sb.WriteString(fmt.Sprintf("\n---\n\n## Skill: %s\n\n", s.Name))
		sb.WriteString(s.Content)
	}

	// Include learnings if any
	if len(e.Learnings) > 0 {
		sb.WriteString("\n---\n\n")
		sb.WriteString(e.RenderLearnings())
	}

	return sb.String()
}
