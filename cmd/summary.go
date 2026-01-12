package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dmoose/checkpoint/internal/file"
	"github.com/dmoose/checkpoint/internal/git"
	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var summaryOpts struct {
	json bool
}

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().BoolVar(&summaryOpts.json, "json", false, "Output as JSON")
}

var summaryCmd = &cobra.Command{
	Use:   "summary [path]",
	Short: "Show project overview and recent activity",
	Long:  `Displays checkpoint count, recent activity, next steps, and patterns.`,
	Args:  cobra.MaximumNArgs(1),
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
		Summary(absPath, summaryOpts.json)
	},
}

// Summary displays project overview and status
func Summary(projectPath string, jsonOutput bool) {
	// Check if checkpoint is initialized
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if !file.Exists(changelogPath) {
		fmt.Fprintf(os.Stderr, "Checkpoint not initialized in %s\n", projectPath)
		fmt.Fprintf(os.Stderr, "Hint: Run 'checkpoint init' to initialize\n")
		os.Exit(1)
	}

	// Gather summary data
	data := gatherSummaryData(projectPath, changelogPath)

	if jsonOutput {
		printJSONSummary(data)
	} else {
		printHumanSummary(data)
	}
}

type summaryData struct {
	projectName            string
	checkpointCount        int
	lastCheckpointTime     string
	lastCheckpointHash     string
	recentCheckpoints      []recentCheckpoint
	gitStatus              string
	gitClean               bool
	pendingRecommendations int
	nextSteps              []nextStepItem
	recentPatterns         []string
}

type recentCheckpoint struct {
	timestamp string
	summary   string
	hash      string
}

type nextStepItem struct {
	summary  string
	priority string
	scope    string
}

func gatherSummaryData(projectPath, changelogPath string) summaryData {
	data := summaryData{
		projectName: filepath.Base(projectPath),
	}

	// Parse changelog
	changelogContent, err := file.ReadFile(changelogPath)
	if err == nil {
		data.checkpointCount = countCheckpointsInChangelog(changelogContent)
		data.recentCheckpoints = extractRecentCheckpoints(changelogContent, 5)
	}

	// Get last checkpoint info from status
	statusPath := filepath.Join(projectPath, config.StatusFileName)
	if file.Exists(statusPath) {
		statusContent, err := file.ReadFile(statusPath)
		if err == nil {
			data.lastCheckpointHash, data.lastCheckpointTime = extractLastCheckpointInfo(statusContent)
			data.nextSteps = extractNextStepsFromStatusFile(statusContent)
		}
	}

	// Check git status
	if ok, _ := git.IsGitRepository(projectPath); ok {
		status, err := git.GetStatus(projectPath)
		if err == nil {
			data.gitStatus = strings.TrimSpace(status)
			data.gitClean = data.gitStatus == ""
		}
	}

	// Count pending recommendations
	projectFilePath := filepath.Join(projectPath, config.ProjectFileName)
	if file.Exists(projectFilePath) {
		data.pendingRecommendations = countRecommendations(projectFilePath)
	}

	// Extract recent patterns from context
	contextPath := filepath.Join(projectPath, config.ContextFileName)
	if file.Exists(contextPath) {
		data.recentPatterns = extractRecentPatterns(contextPath, 5)
	}

	return data
}

func countCheckpointsInChangelog(content string) int {
	separatorCount := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "---" {
			separatorCount++
		}
	}
	// First separator is meta doc, rest are checkpoints
	if separatorCount > 1 {
		return separatorCount - 1
	}
	return 0
}

func extractRecentCheckpoints(content string, count int) []recentCheckpoint {
	var checkpoints []recentCheckpoint
	decoder := yaml.NewDecoder(strings.NewReader(content))

	// Skip meta document
	var metaDoc map[string]interface{}
	decoder.Decode(&metaDoc)

	// Parse checkpoint documents
	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			break
		}

		// Check if this is a checkpoint document
		if timestamp, ok := doc["timestamp"].(string); ok {
			cp := recentCheckpoint{
				timestamp: timestamp,
			}

			if hash, ok := doc["commit_hash"].(string); ok {
				cp.hash = hash
			}

			// Extract first change summary
			if changes, ok := doc["changes"].([]interface{}); ok && len(changes) > 0 {
				if change, ok := changes[0].(map[string]interface{}); ok {
					if summary, ok := change["summary"].(string); ok {
						cp.summary = summary
					}
				}
			}

			checkpoints = append(checkpoints, cp)
		}
	}

	// Return last N checkpoints (they're already in order)
	if len(checkpoints) > count {
		return checkpoints[len(checkpoints)-count:]
	}
	return checkpoints
}

func extractLastCheckpointInfo(statusContent string) (string, string) {
	var status struct {
		LastCommitHash      string `yaml:"last_commit_hash"`
		LastCommitTimestamp string `yaml:"last_commit_timestamp"`
	}
	yaml.Unmarshal([]byte(statusContent), &status)
	return status.LastCommitHash, status.LastCommitTimestamp
}

func extractNextStepsFromStatusFile(statusContent string) []nextStepItem {
	var status struct {
		NextSteps []struct {
			Summary  string `yaml:"summary"`
			Priority string `yaml:"priority"`
			Scope    string `yaml:"scope"`
		} `yaml:"next_steps"`
	}
	yaml.Unmarshal([]byte(statusContent), &status)

	var items []nextStepItem
	for _, ns := range status.NextSteps {
		items = append(items, nextStepItem{
			summary:  ns.Summary,
			priority: ns.Priority,
			scope:    ns.Scope,
		})
	}
	return items
}

func countRecommendations(projectPath string) int {
	content, err := file.ReadFile(projectPath)
	if err != nil {
		return 0
	}

	decoder := yaml.NewDecoder(strings.NewReader(content))
	count := 0
	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			break
		}
		if docType, ok := doc["document_type"].(string); ok && docType == "recommendations" {
			count++
		}
	}
	return count
}

func extractRecentPatterns(contextPath string, maxPatterns int) []string {
	content, err := file.ReadFile(contextPath)
	if err != nil {
		return nil
	}

	var patterns []string
	decoder := yaml.NewDecoder(strings.NewReader(content))

	// Parse all context documents
	var allDocs []map[string]interface{}
	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			break
		}
		allDocs = append(allDocs, doc)
	}

	// Process last few documents
	startIdx := 0
	if len(allDocs) > 3 {
		startIdx = len(allDocs) - 3
	}

	for i := startIdx; i < len(allDocs) && len(patterns) < maxPatterns; i++ {
		doc := allDocs[i]
		if ctx, ok := doc["context"].(map[string]interface{}); ok {
			// Extract established patterns
			if patternsData, ok := ctx["established_patterns"].([]interface{}); ok {
				for _, p := range patternsData {
					if patternMap, ok := p.(map[string]interface{}); ok {
						if pattern, ok := patternMap["pattern"].(string); ok {
							if scope, ok := patternMap["scope"].(string); ok && scope == "project" {
								if len(patterns) < maxPatterns {
									patterns = append(patterns, pattern)
								}
							}
						}
					}
				}
			}
		}
	}

	return patterns
}

func printHumanSummary(data summaryData) {
	fmt.Println()
	fmt.Println("PROJECT SUMMARY")
	fmt.Println(strings.Repeat("━", 60))

	// Project info
	fmt.Printf("Project: %s\n", data.projectName)
	fmt.Printf("Checkpoints: %d total", data.checkpointCount)
	if data.lastCheckpointTime != "" {
		fmt.Printf(" | Last: %s", formatTimeAgo(data.lastCheckpointTime))
	}
	fmt.Println()
	fmt.Println()

	// Current status
	fmt.Println("CURRENT STATUS")
	fmt.Println(strings.Repeat("━", 60))
	if data.gitClean {
		fmt.Println("✓ Working directory clean")
	} else {
		lines := strings.Split(data.gitStatus, "\n")
		fmt.Printf("ℹ Working directory has changes (%d file(s))\n", len(lines))
	}

	if data.pendingRecommendations > 0 {
		fmt.Printf("⚠ %d pending recommendation(s) in .checkpoint-project.yml\n", data.pendingRecommendations)
	}
	fmt.Println()

	// Recent activity
	if len(data.recentCheckpoints) > 0 {
		fmt.Println("RECENT ACTIVITY")
		fmt.Println(strings.Repeat("━", 60))
		for _, cp := range data.recentCheckpoints {
			fmt.Printf("• %s", cp.summary)
			if cp.timestamp != "" {
				fmt.Printf(" [%s]", formatTimeAgo(cp.timestamp))
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Next steps
	if len(data.nextSteps) > 0 {
		fmt.Println("NEXT STEPS")
		fmt.Println(strings.Repeat("━", 60))
		for i, step := range data.nextSteps {
			priority := strings.ToUpper(step.priority)
			if priority == "" {
				priority = "   "
			}
			fmt.Printf("%d. [%s] %s", i+1, priority, step.summary)
			if step.scope != "" {
				fmt.Printf(" (%s)", step.scope)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Recent patterns
	if len(data.recentPatterns) > 0 {
		fmt.Println("RECENT PATTERNS")
		fmt.Println(strings.Repeat("━", 60))
		for _, pattern := range data.recentPatterns {
			fmt.Printf("• %s\n", pattern)
		}
		fmt.Println()
	}

	// Footer
	fmt.Println("💡 Tip: Run 'checkpoint start' for detailed status checks")
	fmt.Println("💡 Tip: Run 'checkpoint guide' to view available guides")
	fmt.Println()
}

func printJSONSummary(data summaryData) {
	fmt.Println("{")
	fmt.Printf("  \"project_name\": \"%s\",\n", data.projectName)
	fmt.Printf("  \"checkpoint_count\": %d,\n", data.checkpointCount)
	fmt.Printf("  \"last_checkpoint_time\": \"%s\",\n", data.lastCheckpointTime)
	fmt.Printf("  \"last_checkpoint_hash\": \"%s\",\n", data.lastCheckpointHash)
	fmt.Printf("  \"git_clean\": %t,\n", data.gitClean)
	fmt.Printf("  \"pending_recommendations\": %d,\n", data.pendingRecommendations)

	// Recent checkpoints
	fmt.Println("  \"recent_checkpoints\": [")
	for i, cp := range data.recentCheckpoints {
		fmt.Printf("    {\"timestamp\": \"%s\", \"summary\": \"%s\", \"hash\": \"%s\"}", cp.timestamp, cp.summary, cp.hash)
		if i < len(data.recentCheckpoints)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}
	fmt.Println("  ],")

	// Next steps
	fmt.Println("  \"next_steps\": [")
	for i, step := range data.nextSteps {
		fmt.Printf("    {\"summary\": \"%s\", \"priority\": \"%s\", \"scope\": \"%s\"}", step.summary, step.priority, step.scope)
		if i < len(data.nextSteps)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}
	fmt.Println("  ],")

	// Recent patterns
	fmt.Println("  \"recent_patterns\": [")
	for i, pattern := range data.recentPatterns {
		fmt.Printf("    \"%s\"", pattern)
		if i < len(data.recentPatterns)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}
	fmt.Println("  ]")
	fmt.Println("}")
}

func formatTimeAgo(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
}
