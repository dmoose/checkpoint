package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/checkpoint/internal/file"
	"github.com/dmoose/checkpoint/internal/prompts"
)

// Prompt displays LLM prompts from .checkpoint/prompts/
func Prompt(projectPath string, promptID string, vars map[string]string) {
	promptsDir := filepath.Join(projectPath, ".checkpoint", "prompts")

	// Check if prompts directory exists
	if !file.Exists(promptsDir) {
		fmt.Fprintf(os.Stderr, "Prompts directory not found at %s\n", promptsDir)
		fmt.Fprintf(os.Stderr, "Hint: Run 'checkpoint init' to create the directory structure\n")
		os.Exit(1)
	}

	// Load prompts configuration
	config, err := prompts.LoadPromptsConfig(promptsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading prompts configuration: %v\n", err)
		fmt.Fprintf(os.Stderr, "Hint: Check that .checkpoint/prompts/prompts.yaml exists and is valid\n")
		os.Exit(1)
	}

	// If no prompt ID specified, list all prompts
	if promptID == "" {
		listPrompts(config)
		return
	}

	// Show specific prompt with variable substitution
	showPrompt(config, promptsDir, projectPath, promptID, vars)
}

// listPrompts displays all available prompts grouped by category
func listPrompts(config *prompts.PromptsConfig) {
	fmt.Println()
	fmt.Println("CHECKPOINT PROMPTS")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()
	fmt.Println("Available prompts:")
	fmt.Println()

	// Get all prompts
	allPrompts := prompts.ListPrompts(config)

	// Group by category
	categories := make(map[string][]prompts.PromptInfo)
	for _, p := range allPrompts {
		categories[p.Category] = append(categories[p.Category], p)
	}

	// Display by category
	categoryOrder := []string{"checkpoint", "development"}
	for _, category := range categoryOrder {
		if prompts, ok := categories[category]; ok {
			// Capitalize category name
			categoryName := strings.Title(category)
			if category == "checkpoint" {
				categoryName = "Checkpoint Workflow"
			} else if category == "development" {
				categoryName = "Development"
			}

			fmt.Printf("%s:\n", categoryName)
			for _, p := range prompts {
				fmt.Printf("  %-20s %s\n", p.ID, p.Description)
			}
			fmt.Println()
			delete(categories, category)
		}
	}

	// Display any remaining categories
	for category, prompts := range categories {
		categoryName := strings.Title(category)
		fmt.Printf("%s:\n", categoryName)
		for _, p := range prompts {
			fmt.Printf("  %-20s %s\n", p.ID, p.Description)
		}
		fmt.Println()
	}

	fmt.Println("Usage:")
	fmt.Println("  checkpoint prompt <id>")
	fmt.Println("  checkpoint prompt <id> --var key=value")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  checkpoint prompt fill-checkpoint")
	fmt.Println("  checkpoint prompt implement-feature --var feature_name=\"User Auth\"")
	fmt.Println()
}

// showPrompt displays a specific prompt with variable substitution
func showPrompt(config *prompts.PromptsConfig, promptsDir string, projectPath string, promptID string, userVars map[string]string) {
	// Get the prompt
	prompt, err := prompts.GetPrompt(config, promptsDir, promptID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'checkpoint prompt' to see available prompts\n")
		os.Exit(1)
	}

	// Build variables map: automatic + global + user
	vars := buildVariables(projectPath, config.Variables, userVars)

	// Substitute variables in template
	output := prompts.SubstituteVariables(prompt.Template, vars)

	// Display the prompt
	fmt.Println()
	fmt.Println(output)
}

// buildVariables creates a complete variables map from automatic, global, and user-provided variables
// Priority order: user-provided > global > automatic
func buildVariables(projectPath string, globalVars map[string]string, userVars map[string]string) map[string]string {
	vars := make(map[string]string)

	// 1. Automatic variables (lowest priority)
	vars["project_name"] = filepath.Base(projectPath)
	vars["project_path"] = projectPath

	// 2. Global variables from prompts.yaml (medium priority)
	for key, value := range globalVars {
		vars[key] = value
	}

	// 3. User-provided variables from --var flags (highest priority)
	for key, value := range userVars {
		vars[key] = value
	}

	return vars
}
