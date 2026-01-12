package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dmoose/checkpoint/pkg/config"
)

// Clean removes artifacts created by the 'check' command so the user can abort and re-run
func Clean(projectPath string) {
	inputPath := filepath.Join(projectPath, config.InputFileName)
	diffPath := filepath.Join(projectPath, config.DiffFileName)
	lockPath := filepath.Join(projectPath, config.LockFileName)

	filesToRemove := []string{inputPath, diffPath, lockPath}
	removedAny := false

	for _, filePath := range filesToRemove {
		if err := os.Remove(filePath); err == nil {
			fmt.Printf("Removed %s\n", filePath)
			removedAny = true
		} else if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: failed to remove %s: %v\n", filePath, err)
		}
	}

	if !removedAny {
		fmt.Println("Nothing to clean (no checkpoint artifacts found)")
	} else {
		fmt.Println("✓ Checkpoint artifacts cleaned")
	}
}
