package config

// Configuration constants for checkpoint tool
const (
	ChangelogFileName    = ".checkpoint-changelog.yaml"
	ContextFileName      = ".checkpoint-context.yaml"
	ProjectFileName      = ".checkpoint-project.yaml"
	InputFileName        = ".checkpoint-input"
	DiffFileName         = ".checkpoint-diff"
	StatusFileName       = ".checkpoint-status.yaml"
	LockFileName         = ".checkpoint-lock"
	CheckpointMdFileName = "CHECKPOINT.md"

	// Legacy file names (for backward compatibility)
	ContextFileNameLegacy = ".checkpoint-context.yml"
	ProjectFileNameLegacy = ".checkpoint-project.yml"

	// Checkpoint directory and schema files
	CheckpointDir           = ".checkpoint"
	ExplainProjectYaml      = "project.yaml"
	ExplainToolsYaml        = "tools.yaml"
	ExplainGuidelinesYaml   = "guidelines.yaml"
	ExplainSkillsYaml       = "skills.yaml"
	SkillsDir               = "skills"

	// Legacy names (for backward compatibility)
	ExplainProjectYmlLegacy    = "project.yml"
	ExplainToolsYmlLegacy      = "tools.yml"
	ExplainGuidelinesYmlLegacy = "guidelines.yml"
	ExplainSkillsYmlLegacy     = "skills.yml"

	// Global config directory
	GlobalConfigDir    = ".config/checkpoint"
	GlobalSkillsDir    = "skills"
	GlobalTemplatesDir = "templates"
)
