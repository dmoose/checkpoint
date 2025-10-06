package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-llm/pkg/config"
)

// Clean removes artifacts created by the 'check' command so the user can abort and re-run
func Clean(projectPath string) {
	inputPath := filepath.Join(projectPath, config.InputFileName)
	diffPath := filepath.Join(projectPath, config.DiffFileName)

	removedAny := false
	if err := os.Remove(inputPath); err == nil {
		fmt.Printf("Removed %s\n", inputPath)
		removedAny = true
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warning: failed to remove %s: %v\n", inputPath, err)
	}

	if err := os.Remove(diffPath); err == nil {
		fmt.Printf("Removed %s\n", diffPath)
		removedAny = true
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warning: failed to remove %s: %v\n", diffPath, err)
	}

	if !removedAny {
		fmt.Println("Nothing to clean (no checkpoint artifacts found)")
	} else {
		fmt.Println("✓ Checkpoint artifacts cleaned")
	}
}