package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-llm/internal/file"
	"go-llm/internal/schema"
	"go-llm/pkg/config"
)

// Lint checks the checkpoint input for obvious mistakes and issues
func Lint(projectPath string) {
	// Check if input file exists
	inputPath := filepath.Join(projectPath, config.InputFileName)
	if !file.Exists(inputPath) {
		fmt.Fprintf(os.Stderr, "error: input file not found at %s\n", inputPath)
		fmt.Fprintf(os.Stderr, "hint: run 'checkpoint check %s' to generate the input file\n", projectPath)
		os.Exit(1)
	}

	// Read and parse input file
	inputContent, err := file.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to read input file: %v\n", err)
		os.Exit(1)
	}

	entry, err := schema.ParseInputFile(inputContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to parse input file: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: check YAML syntax in %s\n", inputPath)
		os.Exit(1)
	}

	// Run basic validation first
	if err := schema.ValidateEntry(entry); err != nil {
		fmt.Printf("❌ Validation errors found:\n")
		fmt.Printf("   %v\n", err)
		fmt.Printf("\n")
	}

	// Run lint checks
	issues := schema.LintEntry(entry)

	if len(issues) == 0 {
		fmt.Printf("✅ No lint issues found\n")
		fmt.Printf("Changes: %d\n", len(entry.Changes))
		if len(entry.NextSteps) > 0 {
			fmt.Printf("Next steps: %d\n", len(entry.NextSteps))
		}
		return
	}

	fmt.Printf("⚠️  Lint issues found:\n")
	for _, issue := range issues {
		fmt.Printf("   - %s\n", issue)
	}
	fmt.Printf("\nTotal issues: %d\n", len(issues))
	fmt.Printf("\nThese are suggestions - you can still commit if the issues are intentional.\n")
}
