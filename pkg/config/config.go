package config

// Configuration constants for checkpoint tool
const (
	ChangelogFileName    = ".checkpoint-changelog.yaml"
	ContextFileName      = ".checkpoint-context.yml"
	ProjectFileName      = ".checkpoint-project.yml"
	InputFileName        = ".checkpoint-input"
	DiffFileName         = ".checkpoint-diff"
	StatusFileName       = ".checkpoint-status.yaml"
	LockFileName         = ".checkpoint-lock"
	CheckpointMdFileName = "CHECKPOINT.md"

	// Checkpoint directory and new schema files
	CheckpointDir        = ".checkpoint"
	ExplainProjectYml    = "project.yml"
	ExplainToolsYml      = "tools.yml"
	ExplainGuidelinesYml = "guidelines.yml"
	ExplainSkillsYml     = "skills.yml"
	SkillsDir            = "skills"

	// Global config directory
	GlobalConfigDir    = ".config/checkpoint"
	GlobalSkillsDir    = "skills"
	GlobalTemplatesDir = "templates"
)
