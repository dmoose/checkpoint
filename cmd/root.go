package cmd

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
	Use:   "checkpoint",
	Short: "LLM-assisted development checkpoint tracking",
	Long: `Checkpoint captures structured development history in git-tracked YAML files.

It solves the problem of LLM-assisted development losing context between sessions
by creating an append-only changelog linking every commit to its reasoning,
decisions, and failed approaches.`,
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
			fmt.Printf("checkpoint %s\n", Build.Version)
			if Build.Commit != "unknown" {
				fmt.Printf("  commit: %s\n", Build.Commit)
			}
			if Build.Date != "unknown" {
				fmt.Printf("  built:  %s\n", Build.Date)
			}
		},
	})
}
