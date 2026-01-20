package guardrail

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// BuildInfo contains version information injected at build time
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// Build holds the current build information
var Build = BuildInfo{
	Version: "dev",
	Commit:  "unknown",
	Date:    "unknown",
}

var rootCmd = &cobra.Command{
	Use:   "guardrail",
	Short: "LLM knowledge and context management",
	Long: `Guardrail manages project knowledge, guidelines, skills, and context
for LLM-assisted development.

It provides tools to initialize, explain, learn, and manage the knowledge base
that helps LLMs understand your project's patterns, conventions, and tooling.`,
}

// Execute runs the root command
func Execute(info BuildInfo) {
	Build = info
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("guardrail %s\n", Build.Version)
			if Build.Commit != "unknown" {
				fmt.Printf("  commit: %s\n", Build.Commit)
			}
			if Build.Date != "unknown" {
				fmt.Printf("  built:  %s\n", Build.Date)
			}
		},
	})
}
