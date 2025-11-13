package explain

// ProjectConfig represents .checkpoint/project.yml
type ProjectConfig struct {
	SchemaVersion string              `yaml:"schema_version"`
	Name          string              `yaml:"name"`
	Type          string              `yaml:"type"`
	Purpose       string              `yaml:"purpose"`
	Repository    string              `yaml:"repository,omitempty"`
	Architecture  ArchitectureConfig  `yaml:"architecture,omitempty"`
	Languages     LanguagesConfig     `yaml:"languages,omitempty"`
	Dependencies  DependenciesConfig  `yaml:"dependencies,omitempty"`
	Integrations  []IntegrationConfig `yaml:"integrations,omitempty"`
}

type ArchitectureConfig struct {
	Overview string            `yaml:"overview,omitempty"`
	KeyPaths map[string]string `yaml:"key_paths,omitempty"`
	DataFlow string            `yaml:"data_flow,omitempty"`
	KeyFiles []KeyFileConfig   `yaml:"key_files,omitempty"`
}

type KeyFileConfig struct {
	Path    string `yaml:"path"`
	Purpose string `yaml:"purpose"`
	Tracked bool   `yaml:"tracked"`
}

type LanguagesConfig struct {
	Primary string `yaml:"primary"`
	Version string `yaml:"version,omitempty"`
}

type DependenciesConfig struct {
	External []ExternalDepConfig `yaml:"external,omitempty"`
}

type ExternalDepConfig struct {
	Name    string `yaml:"name"`
	Purpose string `yaml:"purpose"`
}

type IntegrationConfig struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Interaction string `yaml:"interaction,omitempty"`
}

// ToolsConfig represents .checkpoint/tools.yml
type ToolsConfig struct {
	SchemaVersion string                 `yaml:"schema_version"`
	Build         map[string]ToolCommand `yaml:"build,omitempty"`
	Test          map[string]ToolCommand `yaml:"test,omitempty"`
	Lint          map[string]ToolCommand `yaml:"lint,omitempty"`
	Check         map[string]ToolCommand `yaml:"check,omitempty"`
	Run           map[string]ToolCommand `yaml:"run,omitempty"`
	Checkpoint    map[string]ToolCommand `yaml:"checkpoint,omitempty"`
	Maintenance   map[string]ToolCommand `yaml:"maintenance,omitempty"`
}

type ToolCommand struct {
	Command string `yaml:"command"`
	Output  string `yaml:"output,omitempty"`
	Notes   string `yaml:"notes,omitempty"`
	Example string `yaml:"example,omitempty"`
}

// GuidelinesConfig represents .checkpoint/guidelines.yml
// Uses interface{} for flexible nested structures
type GuidelinesConfig struct {
	SchemaVersion string                 `yaml:"schema_version"`
	Naming        map[string]interface{} `yaml:"naming,omitempty"`
	Structure     map[string]string      `yaml:"structure,omitempty"`
	Errors        map[string]interface{} `yaml:"errors,omitempty"`
	Testing       map[string]interface{} `yaml:"testing,omitempty"`
	Commits       map[string]string      `yaml:"commits,omitempty"`
	Rules         []string               `yaml:"rules,omitempty"`
	Avoid         []string               `yaml:"avoid,omitempty"`
	Principles    []string               `yaml:"principles,omitempty"`
}

// SkillsConfig represents .checkpoint/skills.yml
type SkillsConfig struct {
	SchemaVersion string                 `yaml:"schema_version"`
	Global        []string               `yaml:"global,omitempty"`
	Local         []string               `yaml:"local,omitempty"`
	Config        map[string]interface{} `yaml:"config,omitempty"`
	AutoDetect    AutoDetectConfig       `yaml:"auto_detect,omitempty"`
}

type AutoDetectConfig struct {
	IncludeInExplain []string `yaml:"include_in_explain,omitempty"`
}

// Skill represents a skill definition from skill.md
type Skill struct {
	Name    string
	Path    string
	Content string
	IsLocal bool
}

// Learning represents a captured insight from learnings.yml
type Learning struct {
	Timestamp string `yaml:"timestamp"`
	Learning  string `yaml:"learning"`
}
