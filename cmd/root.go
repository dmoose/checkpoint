package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set by main.go
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "LLM-assisted development checkpoint tracking",
	Long: `Checkpoint captures structured development history in git-tracked YAML files.

It solves the problem of LLM-assisted development losing context between sessions
by creating an append-only changelog linking every commit to its reasoning,
decisions, and failed approaches.`,
}

// Execute runs the root command
func Execute(version string) {
	Version = version
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
			fmt.Printf("checkpoint version %s\n", Version)
		},
	})
}
