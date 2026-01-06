package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "clean [path]",
	Short: "Remove temporary checkpoint files to abort and restart",
	Long: `Deletes .checkpoint-input, .checkpoint-diff, and .checkpoint-lock files.
Use when you need to start over or resolve conflicts.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		if len(args) > 0 {
			projectPath = args[0]
		}
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		Clean(absPath)
	},
}

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
