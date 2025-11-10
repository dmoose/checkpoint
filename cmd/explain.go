package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"go-llm/internal/explain"
)

// ExplainOptions holds flags for the explain command
type ExplainOptions struct {
	Topic     string // project, tools, guidelines, skills, skill, history, or empty for summary
	SkillName string // specific skill name when topic is "skill"
	Full      bool   // --full flag
	Markdown  bool   // --md flag
	JSON      bool   // --json flag
}

// Explain displays project context for LLMs and developers
func Explain(projectPath string, opts ExplainOptions) {
	ctx, err := explain.LoadExplainContext(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading context: %v\n", err)
		os.Exit(1)
	}

	var output string

	switch opts.Topic {
	case "":
		if opts.Full {
			output = ctx.RenderFull()
		} else {
			output = ctx.RenderSummary()
		}
	case "project":
		output = ctx.RenderProject()
	case "tools":
		output = ctx.RenderTools()
	case "guidelines":
		output = ctx.RenderGuidelines()
	case "skills":
		output = ctx.RenderSkills()
	case "skill":
		if opts.SkillName == "" {
			fmt.Fprintf(os.Stderr, "error: skill name required\n")
			fmt.Fprintf(os.Stderr, "usage: checkpoint explain skill <name>\n")
			os.Exit(1)
		}
		output = ctx.RenderSkill(opts.SkillName)
	case "history":
		output = renderHistory(projectPath)
	default:
		// Check if it's a skill name directly
		skillOutput := ctx.RenderSkill(opts.Topic)
		if skillOutput != "" && !isSkillNotFound(skillOutput) {
			output = skillOutput
		} else {
			fmt.Fprintf(os.Stderr, "unknown topic: %s\n", opts.Topic)
			fmt.Fprintf(os.Stderr, "available: project, tools, guidelines, skills, skill <name>, history\n")
			os.Exit(1)
		}
	}

	// Handle output format
	if opts.JSON {
		outputJSON(ctx, opts.Topic)
		return
	}

	fmt.Print(output)
}

func isSkillNotFound(output string) bool {
	return len(output) > 0 && output[0:5] == "Skill"
}

func renderHistory(projectPath string) string {
	// TODO: Implement history rendering from changelog
	return `# Recent History

(History view coming soon - will show recent checkpoints, patterns, and decisions)

For now, see:
- .checkpoint-changelog.yaml for full history
- .checkpoint-context.yml for decisions
- checkpoint search <term> to query history
`
}

func outputJSON(ctx *explain.ExplainOutput, topic string) {
	var data interface{}

	switch topic {
	case "project":
		data = ctx.Project
	case "tools":
		data = ctx.Tools
	case "guidelines":
		data = ctx.Guidelines
	case "skills":
		data = struct {
			Config *explain.SkillsConfig `json:"config"`
			Skills []explain.Skill       `json:"skills"`
		}{ctx.Skills, ctx.SkillDefs}
	default:
		data = struct {
			Project    *explain.ProjectConfig    `json:"project"`
			Tools      *explain.ToolsConfig      `json:"tools"`
			Guidelines *explain.GuidelinesConfig `json:"guidelines"`
			Skills     *explain.SkillsConfig     `json:"skills_config"`
			SkillDefs  []explain.Skill           `json:"skills"`
		}{ctx.Project, ctx.Tools, ctx.Guidelines, ctx.Skills, ctx.SkillDefs}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
