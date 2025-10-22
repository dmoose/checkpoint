package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-llm/internal/file"
	"go-llm/internal/git"
	"go-llm/internal/schema"
	"go-llm/pkg/config"
)

// Check implements Phase 2: generate .checkpoint-input and .checkpoint-diff
func Check(projectPath string) {
	// Validate git repository (robust to worktrees)
	if ok, err := git.IsGitRepository(projectPath); !ok {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: git repository check failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "error: %s is not a git repository\n", projectPath)
		}
		os.Exit(1)
	}

	// Prevent overwriting an in-progress checkpoint
	inputPath := filepath.Join(projectPath, config.InputFileName)
	if file.Exists(inputPath) {
		fmt.Fprintf(os.Stderr, "error: input file already exists at %s\n", inputPath)
		fmt.Fprintf(os.Stderr, "another checkpoint may be in progress; run 'checkpoint commit %s' or remove the file\n", projectPath)
		os.Exit(1)
	}

	// Collect git status and diffs
	status, err := git.GetStatus(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to get git status: %v\n", err)
		os.Exit(1)
	}
	diffText, _ := git.GetCombinedDiff(projectPath) // tolerate no HEAD or empty repo

	// Write diff file
	diffPath := filepath.Join(projectPath, config.DiffFileName)
	if err := file.WriteFile(diffPath, diffText); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write diff file: %v\n", err)
		os.Exit(1)
	}

	// Load previous next_steps from status (if present)
	var prevNextSteps []schema.NextStep
	statusPath := filepath.Join(projectPath, config.StatusFileName)
	if file.Exists(statusPath) {
		if content, err := file.ReadFile(statusPath); err == nil {
			if ns := schema.ExtractNextStepsFromStatus(content); len(ns) > 0 {
				prevNextSteps = ns
			}
		}
	}

	// Generate input file content (multi-change schema)
	inputContent := schema.GenerateInputTemplate(status, config.DiffFileName, prevNextSteps)
	if err := file.WriteFile(inputPath, inputContent); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write input file: %v\n", err)
		_ = os.Remove(diffPath)
		os.Exit(1)
	}

	fmt.Printf("✓ Checkpoint input generated\n")
	fmt.Printf("Input: %s\n", inputPath)
	fmt.Printf("Diff:  %s\n", diffPath)
	fmt.Printf("Next: open the input, fill changes[], then run: checkpoint commit %s\n", projectPath)
}