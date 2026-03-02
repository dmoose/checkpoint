package explain

// ProjectConfig represents .checkpoint/project.yaml
type ProjectConfig struct {
	SchemaVersion  string               `yaml:"schema_version"`
	Name           string               `yaml:"name"`
	Type           string               `yaml:"type"`
	Purpose        string               `yaml:"purpose"`
	Repository     string               `yaml:"repository,omitempty"`
	Architecture   ArchitectureConfig   `yaml:"architecture,omitempty"`
	Languages      LanguagesConfig      `yaml:"languages,omitempty"`
	Dependencies   DependenciesConfig   `yaml:"dependencies,omitempty"`
	Integrations   []IntegrationConfig  `yaml:"integrations,omitempty"`
	AIAuthority    AIAuthorityConfig    `yaml:"ai_authority,omitempty"`
	LessonsLearned []LessonLearnedEntry `yaml:"lessons_learned,omitempty"`
	Roadmap        RoadmapConfig        `yaml:"roadmap,omitempty"`
}

// AIAuthorityConfig defines what AI can do autonomously vs requiring approval
type AIAuthorityConfig struct {
	Autonomous       []string `yaml:"autonomous,omitempty"`
	RequiresApproval []string `yaml:"requires_approval,omitempty"`
	Notes            string   `yaml:"notes,omitempty"`
}

// LessonLearnedEntry captures project-wide lessons from failed approaches
type LessonLearnedEntry struct {
	Topic          string `yaml:"topic"`
	FailedApproach string `yaml:"failed_approach"`
	WhyFailed      string `yaml:"why_failed"`
	Lesson         string `yaml:"lesson"`
	Date           string `yaml:"date,omitempty"`
}

// RoadmapConfig captures project-level future considerations
type RoadmapConfig struct {
	Planned  []RoadmapItem  `yaml:"planned,omitempty"`
	Deferred []DeferredItem `yaml:"deferred,omitempty"`
}

// RoadmapItem represents a planned future task
type RoadmapItem struct {
	Summary  string   `yaml:"summary"`
	Priority string   `yaml:"priority,omitempty"`
	Blockers []string `yaml:"blockers,omitempty"`
}

// DeferredItem represents a task that was considered but deferred
type DeferredItem struct {
	Summary        string `yaml:"summary"`
	Reason         string `yaml:"reason"`
	ReconsiderWhen string `yaml:"reconsider_when,omitempty"`
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

// ToolsConfig represents .checkpoint/tools.yaml
type ToolsConfig struct {
	SchemaVersion string                 `yaml:"schema_version"`
	Build         map[string]ToolCommand `yaml:"build,omitempty"`
	Test          map[string]ToolCommand `yaml:"test,omitempty"`
	Lint          map[string]ToolCommand `yaml:"lint,omitempty"`
	Check         map[string]ToolCommand `yaml:"check,omitempty"`
	Run           map[string]ToolCommand `yaml:"run,omitempty"`
	Checkpoint    map[string]ToolCommand `yaml:"checkpoint,omitempty"`
	Maintenance   map[string]ToolCommand `yaml:"maintenance,omitempty"`
	Verify        VerifyConfig           `yaml:"verify,omitempty"`
}

// VerifyConfig defines pre-commit verification commands
type VerifyConfig struct {
	PreCommit []VerifyCommand `yaml:"pre_commit,omitempty"`
}

// VerifyCommand represents a verification command to run before commit
type VerifyCommand struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description,omitempty"`
	Required    bool   `yaml:"required"`
}

type ToolCommand struct {
	Command string `yaml:"command"`
	Output  string `yaml:"output,omitempty"`
	Notes   string `yaml:"notes,omitempty"`
	Example string `yaml:"example,omitempty"`
}

// GuidelinesConfig represents .checkpoint/guidelines.yaml
// Uses any for flexible nested structures
type GuidelinesConfig struct {
	SchemaVersion string              `yaml:"schema_version"`
	Naming        map[string]any      `yaml:"naming,omitempty"`
	Structure     map[string]string   `yaml:"structure,omitempty"`
	Errors        map[string]any      `yaml:"errors,omitempty"`
	Testing       map[string]any      `yaml:"testing,omitempty"`
	Commits       map[string]string   `yaml:"commits,omitempty"`
	Rules         []string            `yaml:"rules,omitempty"`
	Avoid         []string            `yaml:"avoid,omitempty"`
	Principles    []string            `yaml:"principles,omitempty"`
	Collaboration CollaborationConfig `yaml:"collaboration,omitempty"`
}

// CollaborationConfig defines the human-AI collaboration protocol
type CollaborationConfig struct {
	Protocol   string   `yaml:"protocol,omitempty"`
	Principles []string `yaml:"principles,omitempty"`
}

// SkillsConfig represents .checkpoint/skills.yaml
type SkillsConfig struct {
	SchemaVersion string           `yaml:"schema_version"`
	Global        []string         `yaml:"global,omitempty"`
	Local         []string         `yaml:"local,omitempty"`
	Config        map[string]any   `yaml:"config,omitempty"`
	AutoDetect    AutoDetectConfig `yaml:"auto_detect,omitempty"`
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

// Learning represents a captured insight from learnings.yaml
type Learning struct {
	Timestamp string `yaml:"timestamp"`
	Learning  string `yaml:"learning"`
}
