package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var searchOpts struct {
	failed   bool
	pattern  bool
	decision bool
	scope    string
	recent   int
	context  bool
	json     bool
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolVar(&searchOpts.failed, "failed", false, "Search failed approaches")
	searchCmd.Flags().BoolVar(&searchOpts.pattern, "pattern", false, "Search established patterns")
	searchCmd.Flags().BoolVar(&searchOpts.decision, "decision", false, "Search decisions made")
	searchCmd.Flags().StringVar(&searchOpts.scope, "scope", "", "Filter by scope")
	searchCmd.Flags().IntVar(&searchOpts.recent, "recent", 0, "Limit to recent N checkpoints")
	searchCmd.Flags().BoolVar(&searchOpts.context, "context", false, "Search context file")
	searchCmd.Flags().BoolVar(&searchOpts.json, "json", false, "Output as JSON")
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search checkpoint history",
	Long:  `Search changelog and context files for patterns, decisions, and failures.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}

		opts := SearchOptions{
			Failed:   searchOpts.failed,
			Pattern:  searchOpts.pattern,
			Decision: searchOpts.decision,
			Scope:    searchOpts.scope,
			Recent:   searchOpts.recent,
			Context:  searchOpts.context,
			JSON:     searchOpts.json,
		}
		if len(args) > 0 {
			opts.Query = args[0]
		}
		Search(absPath, opts)
	},
}

// SearchOptions holds flags for the search command
type SearchOptions struct {
	Query    string
	Failed   bool   // Search failed approaches
	Pattern  bool   // Search established patterns
	Decision bool   // Search decisions
	Scope    string // Filter by scope
	Recent   int    // Limit to recent N entries
	Context  bool   // Search context file instead of changelog
	JSON     bool   // Output as JSON
}

// SearchResult represents a search match
type SearchResult struct {
	Source     string `json:"source"`      // "changelog" or "context"
	Timestamp  string `json:"timestamp"`   // Timestamp of the entry
	CommitHash string `json:"commit_hash"` // Commit hash if available
	Section    string `json:"section"`     // "changes", "context", "next_steps", etc.
	Field      string `json:"field"`       // Specific field that matched
	Content    string `json:"content"`     // Matched content
	MatchLine  string `json:"match_line"`  // Line containing match
}

// Search searches checkpoint history
func Search(projectPath string, opts SearchOptions) {
	if opts.Query == "" && !opts.Failed && !opts.Pattern && !opts.Decision {
		fmt.Fprintf(os.Stderr, "error: search query required\n")
		fmt.Fprintf(os.Stderr, "usage: checkpoint search <query> [flags]\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		fmt.Fprintf(os.Stderr, "  --failed      Search failed approaches\n")
		fmt.Fprintf(os.Stderr, "  --pattern     Search established patterns\n")
		fmt.Fprintf(os.Stderr, "  --decision    Search decisions made\n")
		fmt.Fprintf(os.Stderr, "  --scope <s>   Filter by scope\n")
		fmt.Fprintf(os.Stderr, "  --recent <n>  Limit to recent N checkpoints\n")
		fmt.Fprintf(os.Stderr, "  --context     Search context file\n")
		os.Exit(1)
	}

	var results []SearchResult

	// Search changelog
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if changelogResults, err := searchChangelog(changelogPath, opts); err == nil {
		results = append(results, changelogResults...)
	}

	// Search context file
	contextPath := filepath.Join(projectPath, config.ContextFileName)
	if contextResults, err := searchContext(contextPath, opts); err == nil {
		results = append(results, contextResults...)
	}

	// Display results
	if opts.JSON {
		output := struct {
			Query   string         `json:"query"`
			Results []SearchResult `json:"results"`
		}{
			Query:   opts.Query,
			Results: results,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(output)
		return
	}

	if len(results) == 0 {
		fmt.Println("No matches found.")
		return
	}

	fmt.Printf("Found %d match(es):\n\n", len(results))
	for i, r := range results {
		if i > 0 {
			fmt.Println("---")
		}
		fmt.Printf("[%s] %s\n", r.Source, r.Timestamp)
		if r.CommitHash != "" {
			fmt.Printf("Commit: %s\n", r.CommitHash[:min(8, len(r.CommitHash))])
		}
		fmt.Printf("Section: %s", r.Section)
		if r.Field != "" {
			fmt.Printf(" > %s", r.Field)
		}
		fmt.Println()
		fmt.Printf("\n%s\n\n", r.Content)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// searchChangelog searches the changelog file
func searchChangelog(path string, opts SearchOptions) ([]SearchResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	docs := splitYAMLDocuments(string(data))

	// Apply recent limit (skip meta doc)
	startIdx := 1
	if opts.Recent > 0 && len(docs) > opts.Recent+1 {
		startIdx = len(docs) - opts.Recent
	}

	for i := startIdx; i < len(docs); i++ {
		doc := docs[i]
		var entry map[string]interface{}
		if err := yaml.Unmarshal([]byte(doc), &entry); err != nil {
			continue
		}

		// Skip meta document
		if docType, ok := entry["document_type"].(string); ok && docType == "meta" {
			continue
		}

		timestamp, _ := entry["timestamp"].(string)
		commitHash, _ := entry["commit_hash"].(string)

		// Search changes
		if changes, ok := entry["changes"].([]interface{}); ok {
			for _, change := range changes {
				if changeMap, ok := change.(map[string]interface{}); ok {
					if matchesSearch(changeMap, opts) {
						content := formatChangeContent(changeMap)
						results = append(results, SearchResult{
							Source:     "changelog",
							Timestamp:  timestamp,
							CommitHash: commitHash,
							Section:    "changes",
							Content:    content,
						})
					}
				}
			}
		}

		// Search next_steps
		if steps, ok := entry["next_steps"].([]interface{}); ok {
			for _, step := range steps {
				if stepMap, ok := step.(map[string]interface{}); ok {
					if matchesSearch(stepMap, opts) {
						content := formatStepContent(stepMap)
						results = append(results, SearchResult{
							Source:     "changelog",
							Timestamp:  timestamp,
							CommitHash: commitHash,
							Section:    "next_steps",
							Content:    content,
						})
					}
				}
			}
		}
	}

	return results, nil
}

// searchContext searches the context file
func searchContext(path string, opts SearchOptions) ([]SearchResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	docs := splitYAMLDocuments(string(data))

	// Apply recent limit
	startIdx := 0
	if opts.Recent > 0 && len(docs) > opts.Recent {
		startIdx = len(docs) - opts.Recent
	}

	for i := startIdx; i < len(docs); i++ {
		doc := docs[i]
		var entry map[string]interface{}
		if err := yaml.Unmarshal([]byte(doc), &entry); err != nil {
			continue
		}

		timestamp, _ := entry["timestamp"].(string)
		commitHash, _ := entry["commit_hash"].(string)

		// Get context section
		context, ok := entry["context"].(map[string]interface{})
		if !ok {
			continue
		}

		// Search failed approaches
		if opts.Failed || (opts.Query != "" && !opts.Pattern && !opts.Decision) {
			if failed, ok := context["failed_approaches"].([]interface{}); ok {
				for _, item := range failed {
					if matchesQuery(item, opts.Query) || opts.Failed {
						content := formatContextItem("failed_approach", item)
						results = append(results, SearchResult{
							Source:     "context",
							Timestamp:  timestamp,
							CommitHash: commitHash,
							Section:    "context",
							Field:      "failed_approaches",
							Content:    content,
						})
					}
				}
			}
		}

		// Search patterns
		if opts.Pattern || (opts.Query != "" && !opts.Failed && !opts.Decision) {
			if patterns, ok := context["established_patterns"].([]interface{}); ok {
				for _, item := range patterns {
					if matchesQuery(item, opts.Query) || opts.Pattern {
						content := formatContextItem("pattern", item)
						results = append(results, SearchResult{
							Source:     "context",
							Timestamp:  timestamp,
							CommitHash: commitHash,
							Section:    "context",
							Field:      "established_patterns",
							Content:    content,
						})
					}
				}
			}
		}

		// Search decisions
		if opts.Decision || (opts.Query != "" && !opts.Failed && !opts.Pattern) {
			if decisions, ok := context["decisions_made"].([]interface{}); ok {
				for _, item := range decisions {
					if matchesQuery(item, opts.Query) || opts.Decision {
						content := formatContextItem("decision", item)
						results = append(results, SearchResult{
							Source:     "context",
							Timestamp:  timestamp,
							CommitHash: commitHash,
							Section:    "context",
							Field:      "decisions_made",
							Content:    content,
						})
					}
				}
			}
		}

		// Search key insights
		if opts.Query != "" && !opts.Failed && !opts.Pattern && !opts.Decision {
			if insights, ok := context["key_insights"].([]interface{}); ok {
				for _, item := range insights {
					if matchesQuery(item, opts.Query) {
						content := formatContextItem("insight", item)
						results = append(results, SearchResult{
							Source:     "context",
							Timestamp:  timestamp,
							CommitHash: commitHash,
							Section:    "context",
							Field:      "key_insights",
							Content:    content,
						})
					}
				}
			}
		}

		// Search problem statement
		if opts.Query != "" {
			if problem, ok := context["problem_statement"].(string); ok {
				if matchesQueryString(problem, opts.Query) {
					results = append(results, SearchResult{
						Source:     "context",
						Timestamp:  timestamp,
						CommitHash: commitHash,
						Section:    "context",
						Field:      "problem_statement",
						Content:    problem,
					})
				}
			}
		}
	}

	return results, nil
}

func splitYAMLDocuments(content string) []string {
	var docs []string
	parts := strings.Split(content, "\n---")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && part != "---" {
			docs = append(docs, part)
		}
	}
	return docs
}

func matchesSearch(m map[string]interface{}, opts SearchOptions) bool {
	// Check scope filter
	if opts.Scope != "" {
		if scope, ok := m["scope"].(string); ok {
			if !strings.Contains(strings.ToLower(scope), strings.ToLower(opts.Scope)) {
				return false
			}
		} else {
			return false
		}
	}

	// Check query
	if opts.Query != "" {
		return matchesMapQuery(m, opts.Query)
	}

	return true
}

func matchesMapQuery(m map[string]interface{}, query string) bool {
	query = strings.ToLower(query)
	for _, v := range m {
		switch val := v.(type) {
		case string:
			if strings.Contains(strings.ToLower(val), query) {
				return true
			}
		case []interface{}:
			for _, item := range val {
				if str, ok := item.(string); ok {
					if strings.Contains(strings.ToLower(str), query) {
						return true
					}
				}
			}
		}
	}
	return false
}

func matchesQuery(item interface{}, query string) bool {
	if query == "" {
		return true
	}
	query = strings.ToLower(query)

	switch v := item.(type) {
	case string:
		return strings.Contains(strings.ToLower(v), query)
	case map[string]interface{}:
		return matchesMapQuery(v, query)
	}
	return false
}

func matchesQueryString(s, query string) bool {
	if query == "" {
		return false
	}
	// Support regex if query contains special chars
	if strings.ContainsAny(query, ".*+?[](){}|\\^$") {
		if re, err := regexp.Compile("(?i)" + query); err == nil {
			return re.MatchString(s)
		}
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(query))
}

func formatChangeContent(m map[string]interface{}) string {
	var sb strings.Builder
	if summary, ok := m["summary"].(string); ok {
		sb.WriteString(fmt.Sprintf("Summary: %s\n", summary))
	}
	if details, ok := m["details"].(string); ok && details != "" {
		sb.WriteString(fmt.Sprintf("Details: %s\n", details))
	}
	if changeType, ok := m["change_type"].(string); ok {
		sb.WriteString(fmt.Sprintf("Type: %s\n", changeType))
	}
	if scope, ok := m["scope"].(string); ok {
		sb.WriteString(fmt.Sprintf("Scope: %s\n", scope))
	}
	return sb.String()
}

func formatStepContent(m map[string]interface{}) string {
	var sb strings.Builder
	if summary, ok := m["summary"].(string); ok {
		sb.WriteString(fmt.Sprintf("Summary: %s\n", summary))
	}
	if details, ok := m["details"].(string); ok && details != "" {
		sb.WriteString(fmt.Sprintf("Details: %s\n", details))
	}
	if priority, ok := m["priority"].(string); ok {
		sb.WriteString(fmt.Sprintf("Priority: %s\n", priority))
	}
	if scope, ok := m["scope"].(string); ok {
		sb.WriteString(fmt.Sprintf("Scope: %s\n", scope))
	}
	return sb.String()
}

func formatContextItem(itemType string, item interface{}) string {
	switch v := item.(type) {
	case string:
		return v
	case map[string]interface{}:
		var sb strings.Builder
		// Common fields
		for _, key := range []string{"insight", "pattern", "decision", "approach", "description"} {
			if val, ok := v[key].(string); ok {
				sb.WriteString(val)
				sb.WriteString("\n")
				break
			}
		}
		// Additional details
		if rationale, ok := v["rationale"].(string); ok {
			sb.WriteString(fmt.Sprintf("Rationale: %s\n", rationale))
		}
		if why, ok := v["why_failed"].(string); ok {
			sb.WriteString(fmt.Sprintf("Why failed: %s\n", why))
		}
		if lessons, ok := v["lessons_learned"].(string); ok {
			sb.WriteString(fmt.Sprintf("Lessons: %s\n", lessons))
		}
		if scope, ok := v["scope"].(string); ok {
			sb.WriteString(fmt.Sprintf("Scope: %s\n", scope))
		}
		return sb.String()
	default:
		return fmt.Sprintf("%v", item)
	}
}
