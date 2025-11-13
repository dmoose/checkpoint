package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/checkpoint/internal/file"
)

// Examples displays example checkpoint entries from .checkpoint/examples/
func Examples(projectPath string, category string) {
	examplesDir := filepath.Join(projectPath, ".checkpoint", "examples")

	// Check if examples directory exists
	if !file.Exists(examplesDir) {
		fmt.Fprintf(os.Stderr, "Examples directory not found at %s\n", examplesDir)
		fmt.Fprintf(os.Stderr, "Hint: Run 'checkpoint init' to create the directory structure\n")
		os.Exit(1)
	}

	// If no category specified, list available examples
	if category == "" {
		listExamples(examplesDir)
		return
	}

	// Show specific example
	showExample(examplesDir, category)
}

// listExamples shows available example categories
func listExamples(examplesDir string) {
	fmt.Println("\nCHECKPOINT EXAMPLES")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println("\nAvailable examples:")
	fmt.Println()

	examples := map[string]string{
		"feature":       "Good example of adding a new feature",
		"bugfix":        "Good example of fixing a bug with investigation",
		"refactor":      "Good example of refactoring with rationale",
		"context":       "Effective context capture examples",
		"anti-patterns": "Common mistakes to avoid",
	}

	for name, desc := range examples {
		examplePath := filepath.Join(examplesDir, name+"-example.yaml")
		if name == "anti-patterns" {
			examplePath = filepath.Join(examplesDir, "anti-patterns.yaml")
		}
		if name == "context" {
			examplePath = filepath.Join(examplesDir, "context-examples.yaml")
		}

		if file.Exists(examplePath) {
			fmt.Printf("  %-15s %s\n", name, desc)
		}
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  checkpoint examples [category]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  checkpoint examples feature")
	fmt.Println("  checkpoint examples bugfix")
	fmt.Println("  checkpoint examples anti-patterns")
	fmt.Println()
}

// showExample displays a specific example
func showExample(examplesDir string, category string) {
	// Map category to filename
	filename := category + "-example.yaml"
	if category == "anti-patterns" {
		filename = "anti-patterns.yaml"
	}
	if category == "context" {
		filename = "context-examples.yaml"
	}

	examplePath := filepath.Join(examplesDir, filename)

	// Check if example exists
	if !file.Exists(examplePath) {
		fmt.Fprintf(os.Stderr, "Example '%s' not found\n", category)
		fmt.Fprintf(os.Stderr, "Run 'checkpoint examples' to see available examples\n")
		os.Exit(1)
	}

	// Read and display the example
	content, err := file.ReadFile(examplePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading example: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("EXAMPLE: %s\n", strings.ToUpper(category))
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()
	fmt.Println(content)
}
