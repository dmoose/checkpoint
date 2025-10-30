package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-llm/internal/file"
)

// Guide displays guide documents from .checkpoint/guides/
func Guide(projectPath string, topic string) {
	guidesDir := filepath.Join(projectPath, ".checkpoint", "guides")

	// Check if guides directory exists
	if !file.Exists(guidesDir) {
		fmt.Fprintf(os.Stderr, "Guides directory not found at %s\n", guidesDir)
		fmt.Fprintf(os.Stderr, "Hint: Run 'checkpoint init' to create the directory structure\n")
		os.Exit(1)
	}

	// If no topic specified, list available guides
	if topic == "" {
		listGuides(guidesDir)
		return
	}

	// Show specific guide
	showGuide(guidesDir, topic)
}

// listGuides shows available guide topics
func listGuides(guidesDir string) {
	fmt.Println("\nCHECKPOINT GUIDES")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println("\nAvailable guides:")
	fmt.Println()

	guides := map[string]string{
		"first-time-user": "Complete walkthrough for first-time users",
		"llm-workflow":    "LLM integration patterns and workflow",
		"best-practices":  "Best practices for effective checkpoints",
	}

	for name, desc := range guides {
		guidePath := filepath.Join(guidesDir, name+".md")

		if file.Exists(guidePath) {
			fmt.Printf("  %-20s %s\n", name, desc)
		}
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  checkpoint guide [topic]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  checkpoint guide first-time-user")
	fmt.Println("  checkpoint guide llm-workflow")
	fmt.Println("  checkpoint guide best-practices")
	fmt.Println()
}

// showGuide displays a specific guide
func showGuide(guidesDir string, topic string) {
	// Map topic to filename
	filename := topic + ".md"

	guidePath := filepath.Join(guidesDir, filename)

	// Check if guide exists
	if !file.Exists(guidePath) {
		fmt.Fprintf(os.Stderr, "Guide '%s' not found\n", topic)
		fmt.Fprintf(os.Stderr, "Run 'checkpoint guide' to see available guides\n")
		os.Exit(1)
	}

	// Read and display the guide
	content, err := file.ReadFile(guidePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading guide: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("GUIDE: %s\n", strings.ToUpper(strings.ReplaceAll(topic, "-", " ")))
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()
	fmt.Println(content)
}
