package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/checkpoint/internal/changelog"
	"github.com/dmoose/checkpoint/internal/file"
	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize checkpoint in a project",
	Long: `Creates .checkpoint/ directory and initializes changelog and project files.
For knowledge base setup (guides, prompts, skills), use 'guardrail init'.`,
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
		InitWithOptions(absPath, Build.Version)
	},
}

// updateProjectGitignore adds checkpoint artifact entries to the project's .gitignore
func updateProjectGitignore(projectPath string) {
	gitignorePath := filepath.Join(projectPath, ".gitignore")

	checkpointEntries := `
# Checkpoint artifacts (temporary files, not tracked)
checkpoint-input
.checkpoint-diff
.checkpoint-lock
.checkpoint-status.yaml
.checkpoint-session.yaml
`

	// Check if .gitignore exists
	existingContent := ""
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existingContent = string(data)
	}

	// Check if checkpoint entries already exist
	if strings.Contains(existingContent, "checkpoint-input") {
		// Already has checkpoint entries
		return
	}

	// Append checkpoint entries
	var newContent string
	if existingContent == "" {
		newContent = strings.TrimPrefix(checkpointEntries, "\n")
	} else {
		// Ensure there's a newline before our entries
		if !strings.HasSuffix(existingContent, "\n") {
			existingContent += "\n"
		}
		newContent = existingContent + checkpointEntries
	}

	if err := os.WriteFile(gitignorePath, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not update .gitignore: %v\n", err)
		return
	}

	if existingContent == "" {
		fmt.Println("✓ Created .gitignore with checkpoint artifacts")
	} else {
		fmt.Println("✓ Updated .gitignore with checkpoint artifacts")
	}
}

// InitWithOptions creates checkpoint files (trimmed: changelog, project file, gitignore only)
func InitWithOptions(projectPath string, version string) {
	// Create .checkpoint/ directory
	checkpointDir := filepath.Join(projectPath, ".checkpoint")
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating .checkpoint directory: %v\n", err)
		os.Exit(1)
	}

	// Update project .gitignore with checkpoint artifacts
	updateProjectGitignore(projectPath)

	// Initialize changelog with meta document (only if it doesn't exist)
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if !file.Exists(changelogPath) {
		if err := changelog.InitializeChangelog(changelogPath, version); err != nil {
			fmt.Fprintf(os.Stderr, "error initializing changelog: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created %s\n", config.ChangelogFileName)
	} else {
		fmt.Printf("  %s already exists (skipped)\n", config.ChangelogFileName)
	}

	fmt.Printf("\n✓ Checkpoint initialization complete\n")
	fmt.Printf("  .checkpoint/ directory is ready\n")
	fmt.Printf("\nFor knowledge base setup (guides, prompts, skills), run: guardrail init\n")
}
