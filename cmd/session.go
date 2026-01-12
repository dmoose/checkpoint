package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var sessionOpts struct {
	json bool
}

func init() {
	rootCmd.AddCommand(sessionCmd)
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
			JSON: sessionOpts.json,
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
	Action  string // save, show, clear, handoff
	Summary string // session summary when saving
	JSON    bool   // output as JSON
}

// SessionState represents the session planning document
type SessionState struct {
	SchemaVersion string `yaml:"schema_version" json:"schema_version"`
	Created       string `yaml:"created" json:"created"`
	Updated       string `yaml:"updated" json:"updated"`

	// Planning section
	Goals         []string        `yaml:"goals,omitempty" json:"goals,omitempty"`
	Approach      string          `yaml:"approach,omitempty" json:"approach,omitempty"`
	NextActions   []NextAction    `yaml:"next_actions,omitempty" json:"next_actions,omitempty"`
	Risks         []string        `yaml:"risks,omitempty" json:"risks,omitempty"`
	OpenQuestions []string        `yaml:"open_questions,omitempty" json:"open_questions,omitempty"`

	// Active work section
	CurrentFocus  string            `yaml:"current_focus,omitempty" json:"current_focus,omitempty"`
	Progress      []string          `yaml:"progress,omitempty" json:"progress,omitempty"`
	Blockers      []Blocker         `yaml:"blockers,omitempty" json:"blockers,omitempty"`
	Decisions     []SessionDecision `yaml:"decisions,omitempty" json:"decisions,omitempty"`
	Learnings     []string          `yaml:"learnings,omitempty" json:"learnings,omitempty"`
	ModifiedFiles []string          `yaml:"modified_files,omitempty" json:"modified_files,omitempty"`

	// Handoff section (appended by handoff command)
	Handoff *SessionHandoff `yaml:"handoff,omitempty" json:"handoff,omitempty"`
}

// NextAction represents a planned task with status tracking
type NextAction struct {
	Summary   string `yaml:"summary" json:"summary"`
	Priority  string `yaml:"priority,omitempty" json:"priority,omitempty"`     // high, med, low
	Status    string `yaml:"status,omitempty" json:"status,omitempty"`         // pending, in_progress, done, blocked
	BlockedBy string `yaml:"blocked_by,omitempty" json:"blocked_by,omitempty"` // what's blocking this action
}

// Blocker represents something blocking progress
type Blocker struct {
	Issue     string `yaml:"issue" json:"issue"`
	WaitingOn string `yaml:"waiting_on,omitempty" json:"waiting_on,omitempty"`
}

// SessionDecision records a decision made during the session
type SessionDecision struct {
	Decision  string `yaml:"decision" json:"decision"`
	Rationale string `yaml:"rationale,omitempty" json:"rationale,omitempty"`
}

// SessionHandoff contains context for handing off to another session
type SessionHandoff struct {
	Timestamp        string   `yaml:"timestamp" json:"timestamp"`
	Summary          string   `yaml:"summary" json:"summary"`
	Unfinished       []string `yaml:"unfinished,omitempty" json:"unfinished,omitempty"`
	ContextForNext   string   `yaml:"context_for_next,omitempty" json:"context_for_next,omitempty"`
	RecommendedStart string   `yaml:"recommended_start,omitempty" json:"recommended_start,omitempty"`
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
	fmt.Println("# Session")
	fmt.Println()
	fmt.Printf("**Created:** %s\n", session.Created)
	if session.Updated != session.Created {
		fmt.Printf("**Updated:** %s\n", session.Updated)
	}
	fmt.Println()

	// Planning section
	if len(session.Goals) > 0 {
		fmt.Println("## Goals")
		fmt.Println()
		for _, g := range session.Goals {
			fmt.Printf("- %s\n", g)
		}
		fmt.Println()
	}

	if session.Approach != "" {
		fmt.Println("## Approach")
		fmt.Println()
		fmt.Println(session.Approach)
		fmt.Println()
	}

	if len(session.NextActions) > 0 {
		fmt.Println("## Next Actions")
		fmt.Println()
		for _, a := range session.NextActions {
			status := a.Status
			if status == "" {
				status = "pending"
			}
			priority := ""
			if a.Priority != "" {
				priority = fmt.Sprintf("[%s] ", strings.ToUpper(a.Priority))
			}
			fmt.Printf("- %s%s (%s)", priority, a.Summary, status)
			if a.BlockedBy != "" {
				fmt.Printf(" - blocked by: %s", a.BlockedBy)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	if len(session.Risks) > 0 {
		fmt.Println("## Risks")
		fmt.Println()
		for _, r := range session.Risks {
			fmt.Printf("- %s\n", r)
		}
		fmt.Println()
	}

	if len(session.OpenQuestions) > 0 {
		fmt.Println("## Open Questions")
		fmt.Println()
		for _, q := range session.OpenQuestions {
			fmt.Printf("- %s\n", q)
		}
		fmt.Println()
	}

	// Active work section
	if session.CurrentFocus != "" {
		fmt.Println("## Current Focus")
		fmt.Println()
		fmt.Println(session.CurrentFocus)
		fmt.Println()
	}

	if len(session.Progress) > 0 {
		fmt.Println("## Progress")
		fmt.Println()
		for _, p := range session.Progress {
			fmt.Printf("- %s\n", p)
		}
		fmt.Println()
	}

	if len(session.Blockers) > 0 {
		fmt.Println("## Blockers")
		fmt.Println()
		for _, b := range session.Blockers {
			fmt.Printf("- %s", b.Issue)
			if b.WaitingOn != "" {
				fmt.Printf(" (waiting on: %s)", b.WaitingOn)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	if len(session.Decisions) > 0 {
		fmt.Println("## Decisions")
		fmt.Println()
		for _, d := range session.Decisions {
			fmt.Printf("- **%s**", d.Decision)
			if d.Rationale != "" {
				fmt.Printf(": %s", d.Rationale)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	if len(session.Learnings) > 0 {
		fmt.Println("## Learnings")
		fmt.Println()
		for _, l := range session.Learnings {
			fmt.Printf("- %s\n", l)
		}
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

	// Handoff section
	if session.Handoff != nil {
		fmt.Println("## Handoff")
		fmt.Println()
		fmt.Printf("**Timestamp:** %s\n", session.Handoff.Timestamp)
		fmt.Println()
		if session.Handoff.Summary != "" {
			fmt.Println("### Summary")
			fmt.Println()
			fmt.Println(session.Handoff.Summary)
			fmt.Println()
		}
		if len(session.Handoff.Unfinished) > 0 {
			fmt.Println("### Unfinished")
			fmt.Println()
			for _, u := range session.Handoff.Unfinished {
				fmt.Printf("- %s\n", u)
			}
			fmt.Println()
		}
		if session.Handoff.ContextForNext != "" {
			fmt.Println("### Context for Next Session")
			fmt.Println()
			fmt.Println(session.Handoff.ContextForNext)
			fmt.Println()
		}
		if session.Handoff.RecommendedStart != "" {
			fmt.Println("### Recommended Start")
			fmt.Println()
			fmt.Println(session.Handoff.RecommendedStart)
			fmt.Println()
		}
	}
}

func saveSession(projectPath string, opts SessionOptions) {
	sessionPath := filepath.Join(projectPath, sessionFileName)

	// Load existing session or create new one
	var session SessionState
	if existing, err := os.ReadFile(sessionPath); err == nil {
		if err := yaml.Unmarshal(existing, &session); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing existing session: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Create new session
		session = SessionState{
			SchemaVersion: "1",
			Created:       time.Now().Format(time.RFC3339),
		}
	}

	// Update timestamp and modified files
	session.Updated = time.Now().Format(time.RFC3339)
	session.ModifiedFiles = getModifiedFiles(projectPath)

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

	fmt.Println("Session updated.")
	if len(session.ModifiedFiles) > 0 {
		fmt.Printf("  %d modified files tracked\n", len(session.ModifiedFiles))
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
	sessionPath := filepath.Join(projectPath, sessionFileName)

	// Load existing session
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "No active session to hand off.\n")
			fmt.Fprintf(os.Stderr, "hint: Use 'checkpoint plan' to create a session first\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "error reading session: %v\n", err)
		os.Exit(1)
	}

	var session SessionState
	if err := yaml.Unmarshal(data, &session); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing session: %v\n", err)
		os.Exit(1)
	}

	// Build handoff section
	handoff := SessionHandoff{
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Auto-generate summary from session state
	var summaryParts []string
	if session.CurrentFocus != "" {
		summaryParts = append(summaryParts, fmt.Sprintf("Was working on: %s", session.CurrentFocus))
	}
	if len(session.Progress) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("Completed %d items.", len(session.Progress)))
	}
	if len(session.Blockers) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("Blocked on %d issue(s).", len(session.Blockers)))
	}
	if len(summaryParts) > 0 {
		handoff.Summary = strings.Join(summaryParts, " ")
	}

	// Collect unfinished actions
	for _, a := range session.NextActions {
		if a.Status != "done" {
			handoff.Unfinished = append(handoff.Unfinished, a.Summary)
		}
	}

	// Build context for next session
	var contextParts []string
	if len(session.Decisions) > 0 {
		contextParts = append(contextParts, "Key decisions were made - review the Decisions section.")
	}
	if len(session.Learnings) > 0 {
		contextParts = append(contextParts, "Learnings captured - review before continuing.")
	}
	if len(contextParts) > 0 {
		handoff.ContextForNext = strings.Join(contextParts, " ")
	}

	// Recommended start
	if len(handoff.Unfinished) > 0 {
		handoff.RecommendedStart = fmt.Sprintf("Continue with: %s", handoff.Unfinished[0])
	}

	// Update session with handoff
	session.Handoff = &handoff
	session.Updated = time.Now().Format(time.RFC3339)
	session.ModifiedFiles = getModifiedFiles(projectPath)

	// Write updated session
	newData, err := yaml.Marshal(&session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling session: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(sessionPath, newData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing session: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Session prepared for handoff.")
	fmt.Println()
	fmt.Println("The session file now contains handoff context for the next LLM.")
	fmt.Println("Next session can read it with: checkpoint session show")
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
