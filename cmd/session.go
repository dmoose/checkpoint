package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var sessionOpts struct {
	status string
	json   bool
}

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.Flags().StringVar(&sessionOpts.status, "status", "", "Set status (in_progress, blocked, complete, handoff)")
	sessionCmd.Flags().BoolVar(&sessionOpts.json, "json", false, "Output as JSON (for show)")
}

var sessionCmd = &cobra.Command{
	Use:   "session [action] [summary]",
	Short: "Manage session state for LLM handoff",
	Long: `Capture and restore session state across LLM conversations.
Actions: show, save <summary>, clear, handoff`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}

		opts := SessionOptions{
			Status: sessionOpts.status,
			JSON:   sessionOpts.json,
		}
		if len(args) > 0 {
			opts.Action = args[0]
		}
		if len(args) > 1 {
			opts.Summary = args[1]
		}
		Session(absPath, opts)
	},
}

// SessionOptions holds flags for the session command
type SessionOptions struct {
	Action  string // save, show, clear
	Summary string // session summary when saving
	Status  string // in_progress, blocked, complete, handoff
	JSON    bool   // output as JSON
}

// SessionState represents captured session state
type SessionState struct {
	SchemaVersion   string        `yaml:"schema_version" json:"schema_version"`
	Timestamp       string        `yaml:"timestamp" json:"timestamp"`
	Status          string        `yaml:"status" json:"status"` // in_progress, blocked, complete, handoff
	Summary         string        `yaml:"summary" json:"summary"`
	ModifiedFiles   []string      `yaml:"modified_files,omitempty" json:"modified_files,omitempty"`
	CurrentWork     string        `yaml:"current_work,omitempty" json:"current_work,omitempty"`
	BlockedOn       string        `yaml:"blocked_on,omitempty" json:"blocked_on,omitempty"`
	NextActions     []string      `yaml:"next_actions,omitempty" json:"next_actions,omitempty"`
	KeyDecisions    []string      `yaml:"key_decisions,omitempty" json:"key_decisions,omitempty"`
	SessionNotes    string        `yaml:"session_notes,omitempty" json:"session_notes,omitempty"`
	PreviousSession *SessionState `yaml:"previous_session,omitempty" json:"previous_session,omitempty"`
}

const sessionFileName = ".checkpoint-session.yaml"

// Session manages session state for LLM handoff
func Session(projectPath string, opts SessionOptions) {
	switch opts.Action {
	case "", "show":
		showSession(projectPath, opts.JSON)
	case "save":
		saveSession(projectPath, opts)
	case "clear":
		clearSession(projectPath)
	case "handoff":
		handoffSession(projectPath, opts)
	default:
		fmt.Fprintf(os.Stderr, "unknown action: %s\n", opts.Action)
		fmt.Fprintf(os.Stderr, "available: show, save, clear, handoff\n")
		os.Exit(1)
	}
}

func showSession(projectPath string, jsonOutput bool) {
	sessionPath := filepath.Join(projectPath, sessionFileName)
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			if jsonOutput {
				fmt.Println(`{"session": null}`)
			} else {
				fmt.Println("No active session state.")
				fmt.Println("\nhint: Use 'checkpoint session save \"summary\"' to capture session state")
			}
			return
		}
		fmt.Fprintf(os.Stderr, "error reading session: %v\n", err)
		os.Exit(1)
	}

	var session SessionState
	if err := yaml.Unmarshal(data, &session); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing session: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(session)
		return
	}

	renderSession(&session)
}

func renderSession(session *SessionState) {
	fmt.Println("# Session State")
	fmt.Println()
	fmt.Printf("**Status:** %s\n", session.Status)
	fmt.Printf("**Saved:** %s\n", session.Timestamp)
	fmt.Println()

	if session.Summary != "" {
		fmt.Println("## Summary")
		fmt.Println()
		fmt.Println(session.Summary)
		fmt.Println()
	}

	if session.CurrentWork != "" {
		fmt.Println("## Current Work")
		fmt.Println()
		fmt.Println(session.CurrentWork)
		fmt.Println()
	}

	if session.BlockedOn != "" {
		fmt.Println("## Blocked On")
		fmt.Println()
		fmt.Println(session.BlockedOn)
		fmt.Println()
	}

	if len(session.ModifiedFiles) > 0 {
		fmt.Println("## Modified Files")
		fmt.Println()
		for _, f := range session.ModifiedFiles {
			fmt.Printf("- %s\n", f)
		}
		fmt.Println()
	}

	if len(session.NextActions) > 0 {
		fmt.Println("## Next Actions")
		fmt.Println()
		for _, a := range session.NextActions {
			fmt.Printf("- %s\n", a)
		}
		fmt.Println()
	}

	if len(session.KeyDecisions) > 0 {
		fmt.Println("## Key Decisions Made")
		fmt.Println()
		for _, d := range session.KeyDecisions {
			fmt.Printf("- %s\n", d)
		}
		fmt.Println()
	}

	if session.SessionNotes != "" {
		fmt.Println("## Notes")
		fmt.Println()
		fmt.Println(session.SessionNotes)
		fmt.Println()
	}
}

func saveSession(projectPath string, opts SessionOptions) {
	session := SessionState{
		SchemaVersion: "1",
		Timestamp:     time.Now().Format(time.RFC3339),
		Status:        opts.Status,
		Summary:       opts.Summary,
	}

	if session.Status == "" {
		session.Status = "in_progress"
	}

	// Auto-detect modified files from git
	modifiedFiles := getModifiedFiles(projectPath)
	session.ModifiedFiles = modifiedFiles

	// Load existing session to preserve context
	sessionPath := filepath.Join(projectPath, sessionFileName)
	if existing, err := os.ReadFile(sessionPath); err == nil {
		var prev SessionState
		if yaml.Unmarshal(existing, &prev) == nil {
			// Preserve previous session context
			if session.CurrentWork == "" && prev.CurrentWork != "" {
				session.CurrentWork = prev.CurrentWork
			}
			if len(session.NextActions) == 0 && len(prev.NextActions) > 0 {
				session.NextActions = prev.NextActions
			}
			if len(session.KeyDecisions) == 0 && len(prev.KeyDecisions) > 0 {
				session.KeyDecisions = prev.KeyDecisions
			}
		}
	}

	// Write session state
	data, err := yaml.Marshal(&session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling session: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Session state saved (%s)\n", session.Status)
	if len(modifiedFiles) > 0 {
		fmt.Printf("  %d modified files tracked\n", len(modifiedFiles))
	}
}

func clearSession(projectPath string) {
	sessionPath := filepath.Join(projectPath, sessionFileName)
	if err := os.Remove(sessionPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No session to clear.")
			return
		}
		fmt.Fprintf(os.Stderr, "error clearing session: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Session cleared.")
}

func handoffSession(projectPath string, opts SessionOptions) {
	// Create a comprehensive handoff document
	var sb strings.Builder

	sb.WriteString("# Session Handoff\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC3339)))

	// Include session state if exists
	sessionPath := filepath.Join(projectPath, sessionFileName)
	if data, err := os.ReadFile(sessionPath); err == nil {
		var session SessionState
		if yaml.Unmarshal(data, &session) == nil {
			sb.WriteString("## Previous Session\n\n")
			if session.Summary != "" {
				sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", session.Summary))
			}
			if session.CurrentWork != "" {
				sb.WriteString(fmt.Sprintf("**Was working on:** %s\n\n", session.CurrentWork))
			}
			if session.BlockedOn != "" {
				sb.WriteString(fmt.Sprintf("**Blocked on:** %s\n\n", session.BlockedOn))
			}
			if len(session.NextActions) > 0 {
				sb.WriteString("**Next actions:**\n")
				for _, a := range session.NextActions {
					sb.WriteString(fmt.Sprintf("- %s\n", a))
				}
				sb.WriteString("\n")
			}
			if len(session.KeyDecisions) > 0 {
				sb.WriteString("**Key decisions:**\n")
				for _, d := range session.KeyDecisions {
					sb.WriteString(fmt.Sprintf("- %s\n", d))
				}
				sb.WriteString("\n")
			}
		}
	}

	// Include current git status
	sb.WriteString("## Current State\n\n")

	modifiedFiles := getModifiedFiles(projectPath)
	if len(modifiedFiles) > 0 {
		sb.WriteString("**Modified files:**\n")
		for _, f := range modifiedFiles {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("No uncommitted changes.\n\n")
	}

	// Include recent checkpoints from changelog
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if _, err := os.Stat(changelogPath); err == nil {
		sb.WriteString("## Recent Checkpoints\n\n")
		sb.WriteString("See `checkpoint explain history` for recent checkpoint history.\n\n")
	}

	// Include outstanding next steps
	sb.WriteString("## Outstanding Next Steps\n\n")
	sb.WriteString("See `checkpoint explain next` for all outstanding next steps.\n\n")

	// Suggest next steps
	sb.WriteString("## Recommended Entry Points\n\n")
	sb.WriteString("1. `checkpoint explain` - Get project overview\n")
	sb.WriteString("2. `checkpoint explain history` - See recent work\n")
	sb.WriteString("3. `checkpoint explain next` - See outstanding tasks\n")
	sb.WriteString("4. `checkpoint start` - Check readiness and planned work\n")

	fmt.Print(sb.String())
}

func getModifiedFiles(projectPath string) []string {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var files []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 3 {
			// Status is first 2 chars, then space, then filename
			file := strings.TrimSpace(line[2:])
			// Skip checkpoint temporary files
			if strings.HasPrefix(file, ".checkpoint-") && !strings.HasSuffix(file, ".yaml") {
				continue
			}
			files = append(files, file)
		}
	}
	return files
}
