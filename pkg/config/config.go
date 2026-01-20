package config

// Configuration constants for checkpoint tool
const (
	ChangelogFileName    = ".checkpoint-changelog.yaml"
	ContextFileName      = ".checkpoint-context.yaml"
	InputFileName        = "checkpoint-input"
	DiffFileName         = ".checkpoint-diff"
	StatusFileName       = ".checkpoint-status.yaml"
	LockFileName         = ".checkpoint-lock"
	CheckpointMdFileName = "CHECKPOINT.md"

	// Checkpoint directory and schema files
	CheckpointDir           = ".checkpoint"
	ExplainProjectYaml      = "project.yaml"
	ExplainToolsYaml        = "tools.yaml"
	ExplainGuidelinesYaml   = "guidelines.yaml"
	ExplainSkillsYaml       = "skills.yaml"
	SkillsDir               = "skills"

	// Global config directory
	GlobalConfigDir    = ".config/checkpoint"
	GlobalSkillsDir    = "skills"
	GlobalTemplatesDir = "templates"
)
