package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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

	// Create lock file to prevent concurrent checkpoints
	lockPath := filepath.Join(projectPath, config.LockFileName)
	if file.Exists(lockPath) {
		fmt.Fprintf(os.Stderr, "error: checkpoint lock file exists at %s\n", lockPath)
		fmt.Fprintf(os.Stderr, "another checkpoint is in progress; run 'checkpoint commit %s' or 'checkpoint clean %s' to resolve\n", projectPath, projectPath)
		os.Exit(1)
	}

	// Prevent overwriting an in-progress checkpoint
	inputPath := filepath.Join(projectPath, config.InputFileName)
	if file.Exists(inputPath) {
		fmt.Fprintf(os.Stderr, "error: input file already exists at %s\n", inputPath)
		fmt.Fprintf(os.Stderr, "another checkpoint may be in progress; run 'checkpoint commit %s' or 'checkpoint clean %s' to resolve\n", projectPath, projectPath)
		os.Exit(1)
	}

	// Create lock file
	if err := file.WriteFile(lockPath, fmt.Sprintf("pid=%d\ntimestamp=%s\n", os.Getpid(), time.Now().Format(time.RFC3339))); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to create lock file: %v\n", err)
		os.Exit(1)
	}

	// Collect git status and diffs
	status, err := git.GetStatus(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to get git status: %v\n", err)
		os.Exit(1)
	}
	diffText, _ := git.GetCombinedDiff(projectPath) // tolerate no HEAD or empty repo

	// Collect file change statistics
	numstat, _ := git.GetDiffNumStat(projectPath) // tolerate no HEAD or empty repo
	stagedNumstat, _ := git.GetStagedDiffNumStat(projectPath)

	// Parse file statistics
	var filesChanged []schema.FileChange
	if numstat != "" {
		filesChanged = append(filesChanged, schema.ParseNumStat(numstat)...)
	}
	if stagedNumstat != "" {
		filesChanged = append(filesChanged, schema.ParseNumStat(stagedNumstat)...)
	}

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
	inputContent := schema.GenerateInputTemplateWithMetadata(status, config.DiffFileName, prevNextSteps, filesChanged, nil)
	if err := file.WriteFile(inputPath, inputContent); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write input file: %v\n", err)
		_ = os.Remove(diffPath)
		_ = os.Remove(lockPath) // Clean up lock file on failure
		os.Exit(1)
	}

	fmt.Printf("✓ Checkpoint input generated\n")
	fmt.Printf("Input: %s\n", inputPath)
	fmt.Printf("Diff:  %s\n", diffPath)
	fmt.Printf("Next: open the input, fill changes[], then run: checkpoint commit %s\n", projectPath)
}
