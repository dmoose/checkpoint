package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-llm/internal/file"
	"go-llm/internal/git"
	"go-llm/internal/schema"
	"go-llm/pkg/config"

	"gopkg.in/yaml.v3"
)

// Start validates project readiness and shows next steps
func Start(projectPath string) {
	if !startInternal(projectPath) {
		os.Exit(1)
	}
}

// startInternal is the testable implementation
func startInternal(projectPath string) bool {
	hasErrors := false
	hasWarnings := false

	fmt.Println("\nCHECKPOINT START")
	fmt.Println(strings.Repeat("━", 60))

	// Check 1: Git repository
	if ok, err := git.IsGitRepository(projectPath); !ok {
		fmt.Println("✗ Not a git repository")
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		}
		fmt.Println("  Hint: run 'git init' to initialize a repository")
		hasErrors = true
	} else {
		fmt.Println("✓ Git repository detected")
	}

	// Check 2: Checkpoint initialized
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if !file.Exists(changelogPath) {
		fmt.Println("✗ Checkpoint not initialized")
		fmt.Println("  Hint: run 'checkpoint init' to set up checkpoint")
		hasErrors = true
	} else {
		// Count checkpoints
		content, err := file.ReadFile(changelogPath)
		if err != nil {
			fmt.Println("✓ Checkpoint initialized (unable to read changelog)")
		} else {
			count := countCheckpoints(content)
			fmt.Printf("✓ Checkpoint initialized: %d checkpoint(s) in history\n", count)
		}
	}

	// Check 3: No checkpoint in progress
	lockPath := filepath.Join(projectPath, config.LockFileName)
	inputPath := filepath.Join(projectPath, config.InputFileName)
	if file.Exists(lockPath) || file.Exists(inputPath) {
		fmt.Println("⚠ Checkpoint in progress")
		fmt.Println("  You have an unfinished checkpoint")
		fmt.Println("  Options:")
		fmt.Println("    - Continue: edit .checkpoint-input and run 'checkpoint commit'")
		fmt.Println("    - Abort: run 'checkpoint clean' to start over")
		hasWarnings = true
	} else {
		fmt.Println("✓ No checkpoint in progress")
	}

	// Check 4: Git status
	if hasErrors {
		// Skip git status if we don't have a valid git repo
	} else {
		status, err := git.GetStatus(projectPath)
		if err != nil {
			fmt.Printf("⚠ Unable to check git status: %v\n", err)
			hasWarnings = true
		} else {
			status = strings.TrimSpace(status)
			if status == "" {
				fmt.Println("✓ Working directory clean")
			} else {
				lines := strings.Split(status, "\n")
				fmt.Printf("ℹ Working directory has changes (%d file(s))\n", len(lines))
			}
		}
	}

	// Check 5: Pending recommendations
	if !hasErrors {
		projectFilePath := filepath.Join(projectPath, config.ProjectFileName)
		recCount := countPendingRecommendations(projectFilePath)
		if recCount > 0 {
			fmt.Printf("⚠ %d pending recommendation(s) in .checkpoint-project.yml\n", recCount)
			fmt.Println("  Hint: review and curate recommendations periodically")
			hasWarnings = true
		}
	}

	fmt.Println(strings.Repeat("━", 60))

	// Exit early if errors
	if hasErrors {
		fmt.Println("\n❌ Cannot start - fix errors above first")
		return false
	}

	// Show next steps from last checkpoint
	if hasWarnings {
		fmt.Println()
	}

	showNextSteps(projectPath)

	fmt.Println("\nREADY TO WORK")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println("Make your changes, then run:")
	fmt.Println("  checkpoint check    # When ready to create a checkpoint")
	fmt.Println()

	return true
}

// countCheckpoints counts the number of checkpoint documents in changelog
func countCheckpoints(changelogContent string) int {
	// Count "---" separators - in YAML multi-doc, separators indicate doc boundaries
	// Format: --- (meta doc) --- (checkpoint 1) --- (checkpoint 2) ...
	// So: separators - 1 = checkpoint count (first separator is for meta doc)
	separatorCount := 0
	for _, line := range strings.Split(changelogContent, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			separatorCount++
		}
	}

	// First separator is for meta doc, rest are checkpoints
	if separatorCount > 1 {
		return separatorCount - 1
	}
	return 0
}

// countPendingRecommendations counts recommendation documents in project file
func countPendingRecommendations(projectPath string) int {
	if !file.Exists(projectPath) {
		return 0
	}

	content, err := file.ReadFile(projectPath)
	if err != nil {
		return 0
	}

	// Parse multi-document YAML
	decoder := yaml.NewDecoder(strings.NewReader(content))
	count := 0
	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			break
		}
		// Check if this is a recommendations document
		if docType, ok := doc["document_type"].(string); ok && docType == "recommendations" {
			count++
		}
	}
	return count
}

// showNextSteps displays next steps from last checkpoint
func showNextSteps(projectPath string) {
	statusPath := filepath.Join(projectPath, config.StatusFileName)
	if !file.Exists(statusPath) {
		return
	}

	content, err := file.ReadFile(statusPath)
	if err != nil {
		return
	}

	nextSteps := schema.ExtractNextStepsFromStatus(content)
	if len(nextSteps) == 0 {
		return
	}

	fmt.Println("\nNEXT STEPS (from last checkpoint)")
	fmt.Println(strings.Repeat("━", 60))

	// Group by priority
	high := []schema.NextStep{}
	med := []schema.NextStep{}
	low := []schema.NextStep{}
	none := []schema.NextStep{}

	for _, step := range nextSteps {
		switch strings.ToLower(step.Priority) {
		case "high":
			high = append(high, step)
		case "med":
			med = append(med, step)
		case "low":
			low = append(low, step)
		default:
			none = append(none, step)
		}
	}

	// Print in priority order
	num := 1
	for _, step := range high {
		fmt.Printf(" %d. [HIGH] %s", num, step.Summary)
		if step.Scope != "" {
			fmt.Printf(" (%s)", step.Scope)
		}
		fmt.Println()
		if step.Details != "" {
			fmt.Printf("    %s\n", step.Details)
		}
		num++
	}
	for _, step := range med {
		fmt.Printf(" %d. [MED]  %s", num, step.Summary)
		if step.Scope != "" {
			fmt.Printf(" (%s)", step.Scope)
		}
		fmt.Println()
		if step.Details != "" {
			fmt.Printf("    %s\n", step.Details)
		}
		num++
	}
	for _, step := range low {
		fmt.Printf(" %d. [LOW]  %s", num, step.Summary)
		if step.Scope != "" {
			fmt.Printf(" (%s)", step.Scope)
		}
		fmt.Println()
		if step.Details != "" {
			fmt.Printf("    %s\n", step.Details)
		}
		num++
	}
	for _, step := range none {
		fmt.Printf(" %d. %s", num, step.Summary)
		if step.Scope != "" {
			fmt.Printf(" (%s)", step.Scope)
		}
		fmt.Println()
		if step.Details != "" {
			fmt.Printf("    %s\n", step.Details)
		}
		num++
	}
}
