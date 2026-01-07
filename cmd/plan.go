package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var planOpts struct {
	fresh bool
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().BoolVar(&planOpts.fresh, "fresh", false, "Create fresh session, replacing any existing one")
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Start a planning session",
	Long: `Create or continue a planning session.

Creates a .checkpoint-session.yaml file with a template for planning your work.
The session file guides development and is cleared on commit.

Examples:
  checkpoint plan          # Create session or show existing
  checkpoint plan --fresh  # Replace existing session with fresh template`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		Plan(absPath, planOpts.fresh)
	},
}

// Plan creates or continues a planning session
func Plan(projectPath string, fresh bool) {
	sessionPath := filepath.Join(projectPath, sessionFileName)

	// Check if session already exists
	if _, err := os.Stat(sessionPath); err == nil && !fresh {
		fmt.Println("Session already exists.")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  checkpoint session       # View current session")
		fmt.Println("  checkpoint plan --fresh  # Start fresh, replacing existing")
		fmt.Println()
		fmt.Println("Current session:")
		fmt.Println()
		showSession(projectPath, false)
		return
	}

	// Create new session from template
	session := createSessionTemplate()

	// Write session file
	data, err := yaml.Marshal(&session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling session: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing session: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Planning session created.")
	fmt.Println()
	fmt.Println("Edit .checkpoint-session.yaml to plan your work:")
	fmt.Println("  - Set your goals for this session")
	fmt.Println("  - Outline your approach")
	fmt.Println("  - List next actions with priorities")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  checkpoint session         # View session")
	fmt.Println("  checkpoint session save    # Update modified files list")
	fmt.Println("  checkpoint session handoff # Prepare for handoff")
	fmt.Println("  checkpoint commit          # Commit and clear session")
}

// createSessionTemplate creates a new session with template structure
func createSessionTemplate() SessionState {
	now := time.Now().Format(time.RFC3339)

	return SessionState{
		SchemaVersion: "1",
		Created:       now,
		Updated:       now,

		// Planning section
		Goals: []string{
			"[What do you want to accomplish this session?]",
		},
		Approach: "[How will you tackle this? What's the general plan?]",
		NextActions: []NextAction{
			{
				Summary:   "[First task to work on]",
				Priority:  "high",
				Status:    "pending",
				BlockedBy: "",
			},
		},
		Risks: []string{
			"[What could go wrong? What should you watch out for?]",
		},
		OpenQuestions: []string{
			"[What needs clarification before or during implementation?]",
		},

		// Active work section
		CurrentFocus: "[What are you working on right now?]",
		Progress:     []string{},
		Blockers:     []Blocker{},
		Decisions:    []SessionDecision{},
		Learnings:    []string{},
		ModifiedFiles: []string{},

		// Handoff section (populated by checkpoint session handoff)
		Handoff: nil,
	}
}
