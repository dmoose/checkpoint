package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-llm/cmd"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		cmd.Help()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]
	projectPath := "."
	// crude flag parsing: collect known flags, last non-flag is path
	dryRun := false
	changelogOnly := false
	jsonOutput := false
	varFlags := make(map[string]string)
	var positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-n" || a == "--dry-run" {
			dryRun = true
			continue
		}
		if a == "--changelog-only" {
			changelogOnly = true
			continue
		}
		if a == "--json" {
			jsonOutput = true
			continue
		}
		if a == "--var" && i+1 < len(args) {
			// Parse key=value
			parts := strings.SplitN(args[i+1], "=", 2)
			if len(parts) == 2 {
				varFlags[parts[0]] = parts[1]
			}
			i++ // Skip next arg
			continue
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		positional = append(positional, a)
	}
	if len(positional) > 0 {
		projectPath = positional[len(positional)-1]
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
		os.Exit(1)
	}

	switch subcommand {
	case "explain":
		// Parse explain-specific flags
		explainOpts := cmd.ExplainOptions{}
		var explainPositional []string
		for i := 0; i < len(args); i++ {
			a := args[i]
			if a == "--full" {
				explainOpts.Full = true
				continue
			}
			if a == "--md" {
				explainOpts.Markdown = true
				continue
			}
			if a == "--json" {
				explainOpts.JSON = true
				continue
			}
			if strings.HasPrefix(a, "-") {
				continue
			}
			explainPositional = append(explainPositional, a)
		}
		// Parse positional: [topic] [skillname] [path]
		// topic can be: project, tools, guidelines, skills, skill, history
		// if topic is "skill", next positional is skill name
		explainPath := "."
		if len(explainPositional) > 0 {
			first := explainPositional[0]
			if first == "skill" && len(explainPositional) > 1 {
				explainOpts.Topic = "skill"
				explainOpts.SkillName = explainPositional[1]
				if len(explainPositional) > 2 {
					explainPath = explainPositional[2]
				}
			} else if isExplainTopic(first) {
				explainOpts.Topic = first
				if len(explainPositional) > 1 {
					explainPath = explainPositional[1]
				}
			} else if looksLikePath(first) {
				explainPath = first
			} else {
				// Could be a skill name shorthand
				explainOpts.Topic = first
				if len(explainPositional) > 1 {
					explainPath = explainPositional[1]
				}
			}
		}
		explainAbsPath, err := filepath.Abs(explainPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.Explain(explainAbsPath, explainOpts)
	case "start":
		cmd.Start(absPath)
	case "summary":
		cmd.Summary(absPath, jsonOutput)
	case "check":
		cmd.Check(absPath)
	case "commit":
		cmd.CommitWithOptions(absPath, cmd.CommitOptions{DryRun: dryRun, ChangelogOnly: changelogOnly}, version)
	case "init":
		// Parse init-specific flags
		initOpts := cmd.InitOptions{}
		var initPath string
		for i := 0; i < len(args); i++ {
			a := args[i]
			if a == "--list-templates" {
				initOpts.ListTemplates = true
				continue
			}
			if a == "--template" && i+1 < len(args) {
				initOpts.Template = args[i+1]
				i++
				continue
			}
			if strings.HasPrefix(a, "--template=") {
				initOpts.Template = strings.TrimPrefix(a, "--template=")
				continue
			}
			if strings.HasPrefix(a, "-") {
				continue
			}
			initPath = a
		}
		if initPath == "" {
			initPath = "."
		}
		initAbsPath, err := filepath.Abs(initPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.InitWithOptions(initAbsPath, version, initOpts)
	case "clean":
		cmd.Clean(absPath)
	case "lint":
		cmd.Lint(absPath)
	case "examples":
		// Examples takes optional category as first positional arg
		// The path handling should use "." as default, not interpret category as path
		category := ""
		examplesPath := "."

		if len(positional) > 0 {
			// First arg could be category or path
			// Check if it looks like a path (has / or . or is a directory)
			firstArg := positional[0]
			if strings.Contains(firstArg, "/") || strings.Contains(firstArg, ".") ||
				(len(positional) > 1) {
				// Looks like we have both category and path
				if len(positional) == 1 {
					// Just a path
					examplesPath = firstArg
				} else {
					// category and path
					category = firstArg
					examplesPath = positional[1]
				}
			} else {
				// Single word with no path separators - treat as category
				category = firstArg
			}
		}

		// Resolve the path
		examplesAbsPath, err := filepath.Abs(examplesPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.Examples(examplesAbsPath, category)
	case "guide":
		// Guide takes optional topic as first positional arg
		// Same logic as examples command
		topic := ""
		guidePath := "."

		if len(positional) > 0 {
			firstArg := positional[0]
			if strings.Contains(firstArg, "/") || strings.Contains(firstArg, ".") ||
				(len(positional) > 1) {
				if len(positional) == 1 {
					guidePath = firstArg
				} else {
					topic = firstArg
					guidePath = positional[1]
				}
			} else {
				topic = firstArg
			}
		}

		guideAbsPath, err := filepath.Abs(guidePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.Guide(guideAbsPath, topic)
	case "mcp":
		// MCP stdio server; supports --root flags
		var roots []string
		for i := 0; i < len(args); i++ {
			a := args[i]
			if a == "--root" && i+1 < len(args) {
				roots = append(roots, args[i+1])
				i++
			}
		}
		if err := cmd.MCP(cmd.MCPOptions{Roots: roots}); err != nil {
			fmt.Fprintf(os.Stderr, "mcp error: %v\n", err)
			os.Exit(1)
		}
	case "prompt", "prompts":
		// Prompt takes optional prompt ID as first positional arg
		// Same logic as examples/guide commands
		promptID := ""
		promptPath := "."

		if len(positional) > 0 {
			firstArg := positional[0]
			if strings.Contains(firstArg, "/") || strings.Contains(firstArg, ".") ||
				(len(positional) > 1) {
				if len(positional) == 1 {
					promptPath = firstArg
				} else {
					promptID = firstArg
					promptPath = positional[1]
				}
			} else {
				promptID = firstArg
			}
		}

		promptAbsPath, err := filepath.Abs(promptPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.Prompt(promptAbsPath, promptID, varFlags)
	case "learn":
		// Parse learn-specific flags
		learnOpts := cmd.LearnOptions{}
		var learnPositional []string
		for i := 0; i < len(args); i++ {
			a := args[i]
			if a == "--guideline" {
				learnOpts.Guideline = true
				continue
			}
			if a == "--avoid" {
				learnOpts.Avoid = true
				continue
			}
			if a == "--principle" {
				learnOpts.Principle = true
				continue
			}
			if a == "--pattern" {
				learnOpts.Pattern = true
				continue
			}
			if a == "--tool" {
				learnOpts.Tool = true
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					learnOpts.ToolName = args[i+1]
					i++
				}
				continue
			}
			if strings.HasPrefix(a, "-") {
				continue
			}
			learnPositional = append(learnPositional, a)
		}
		// First positional is content, last might be path
		learnPath := "."
		if len(learnPositional) > 0 {
			learnOpts.Content = learnPositional[0]
			if len(learnPositional) > 1 && looksLikePath(learnPositional[len(learnPositional)-1]) {
				learnPath = learnPositional[len(learnPositional)-1]
			}
		}
		learnAbsPath, err := filepath.Abs(learnPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.Learn(learnAbsPath, learnOpts)
	case "search":
		// Parse search-specific flags
		searchOpts := cmd.SearchOptions{}
		var searchPositional []string
		for i := 0; i < len(args); i++ {
			a := args[i]
			if a == "--failed" {
				searchOpts.Failed = true
				continue
			}
			if a == "--pattern" {
				searchOpts.Pattern = true
				continue
			}
			if a == "--decision" {
				searchOpts.Decision = true
				continue
			}
			if a == "--context" {
				searchOpts.Context = true
				continue
			}
			if a == "--scope" && i+1 < len(args) {
				searchOpts.Scope = args[i+1]
				i++
				continue
			}
			if strings.HasPrefix(a, "--scope=") {
				searchOpts.Scope = strings.TrimPrefix(a, "--scope=")
				continue
			}
			if a == "--recent" && i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &searchOpts.Recent)
				i++
				continue
			}
			if strings.HasPrefix(a, "-") {
				continue
			}
			searchPositional = append(searchPositional, a)
		}
		// First positional is query, last might be path
		searchPath := "."
		if len(searchPositional) > 0 {
			searchOpts.Query = searchPositional[0]
			if len(searchPositional) > 1 {
				searchPath = searchPositional[len(searchPositional)-1]
			}
		}
		searchAbsPath, err := filepath.Abs(searchPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.Search(searchAbsPath, searchOpts)
	case "skill", "skills":
		// Parse skill-specific args
		skillOpts := cmd.SkillOptions{}
		var skillPositional []string
		for i := 0; i < len(args); i++ {
			a := args[i]
			if strings.HasPrefix(a, "-") {
				continue
			}
			skillPositional = append(skillPositional, a)
		}
		// Parse: [action] [name] [path]
		// Actions: list, show, add, create
		skillPath := "."
		if len(skillPositional) > 0 {
			first := skillPositional[0]
			if isSkillAction(first) {
				skillOpts.Action = first
				if len(skillPositional) > 1 {
					skillOpts.SkillName = skillPositional[1]
				}
				if len(skillPositional) > 2 {
					skillPath = skillPositional[2]
				}
			} else if looksLikePath(first) {
				skillPath = first
			} else {
				// Treat as skill name for show
				skillOpts.Action = "show"
				skillOpts.SkillName = first
				if len(skillPositional) > 1 {
					skillPath = skillPositional[1]
				}
			}
		}
		skillAbsPath, err := filepath.Abs(skillPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.Skill(skillAbsPath, skillOpts)
	case "help", "-h", "--help":
		cmd.Help()
	case "version", "-v", "--version":
		fmt.Printf("checkpoint version %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", subcommand)
		cmd.Help()
		os.Exit(1)
	}
}

// isExplainTopic returns true if the string is a known explain topic
func isExplainTopic(s string) bool {
	topics := []string{"project", "tools", "guidelines", "skills", "skill", "history"}
	for _, t := range topics {
		if s == t {
			return true
		}
	}
	return false
}

// looksLikePath returns true if the string looks like a file path
func looksLikePath(s string) bool {
	return strings.Contains(s, "/") || strings.Contains(s, ".") || s == "."
}

// isSkillAction returns true if the string is a known skill action
func isSkillAction(s string) bool {
	actions := []string{"list", "show", "add", "create"}
	for _, a := range actions {
		if s == a {
			return true
		}
	}
	return false
}
