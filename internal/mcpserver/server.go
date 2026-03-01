package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"

	cpcontext "github.com/dmoose/checkpoint/internal/context"
	"github.com/dmoose/checkpoint/internal/explain"
	"github.com/dmoose/checkpoint/internal/guides"
	"github.com/dmoose/checkpoint/internal/schema"
	"github.com/dmoose/checkpoint/pkg/config"
)

// NewServer creates an MCP server exposing read-only checkpoint project context tools.
func NewServer(projectPath string) *server.MCPServer {
	s := server.NewMCPServer(
		"checkpoint",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
		server.WithLogging(),
	)

	addExplainTool(s, projectPath)
	addSearchTool(s, projectPath)
	addGuideTool(s)
	addStatusTool(s, projectPath)
	addContextTemplateTool(s)

	return s
}

func addExplainTool(s *server.MCPServer, projectPath string) {
	tool := mcp.NewTool("explain",
		mcp.WithDescription("Get project context for LLMs - architecture, patterns, guidelines, tools"),
		mcp.WithString("topic",
			mcp.Description("Topic to explain: project, tools, guidelines, skills, learnings, or empty for summary"),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		topic := req.GetString("topic", "")

		explainCtx, err := explain.LoadExplainContext(projectPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to load project context: %v", err)), nil
		}

		var output string
		switch topic {
		case "project":
			output = explainCtx.RenderProject()
		case "tools":
			output = explainCtx.RenderTools()
		case "guidelines":
			output = explainCtx.RenderGuidelines()
		case "skills":
			output = explainCtx.RenderSkills()
		case "learnings":
			output = explainCtx.RenderLearnings()
		case "":
			output = explainCtx.RenderSummary()
		default:
			// Try as a skill name
			output = explainCtx.RenderSkill(topic)
		}

		return mcp.NewToolResultText(output), nil
	})
}

func addSearchTool(s *server.MCPServer, projectPath string) {
	tool := mcp.NewTool("search",
		mcp.WithDescription("Search checkpoint history for decisions, patterns, and failed approaches"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query string")),
		mcp.WithBoolean("failed", mcp.Description("Search failed approaches only")),
		mcp.WithBoolean("decision", mcp.Description("Search decisions only")),
		mcp.WithNumber("recent", mcp.Description("Limit to N most recent entries")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := req.GetString("query", "")
		searchFailed := req.GetBool("failed", false)
		searchDecision := req.GetBool("decision", false)
		recent := req.GetInt("recent", 0)

		queryLower := strings.ToLower(query)
		var results []string

		// Search changelog entries
		changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
		if data, err := os.ReadFile(changelogPath); err == nil {
			entries := parseChangelogEntries(data)
			if recent > 0 && len(entries) > recent {
				entries = entries[len(entries)-recent:]
			}
			for _, entry := range entries {
				for _, change := range entry.Changes {
					if matchesQuery(change.Summary, queryLower) || matchesQuery(change.Details, queryLower) || matchesQuery(change.Scope, queryLower) {
						results = append(results, fmt.Sprintf("[%s] %s: %s (%s)",
							entry.Timestamp, change.ChangeType, change.Summary, change.Scope))
					}
				}
			}
		}

		// Search context entries (decisions, failed approaches)
		contextPath := filepath.Join(projectPath, config.ContextFileName)
		if data, err := os.ReadFile(contextPath); err == nil {
			ctxEntries := parseContextEntries(data)
			if recent > 0 && len(ctxEntries) > recent {
				ctxEntries = ctxEntries[len(ctxEntries)-recent:]
			}
			for _, entry := range ctxEntries {
				if !searchFailed {
					for _, d := range entry.Context.DecisionsMade {
						if matchesQuery(d.Decision, queryLower) || matchesQuery(d.Rationale, queryLower) {
							results = append(results, fmt.Sprintf("[%s] DECISION: %s | Rationale: %s",
								entry.Timestamp, d.Decision, d.Rationale))
						}
					}
				}
				if !searchDecision {
					for _, f := range entry.Context.FailedApproaches {
						if matchesQuery(f.Approach, queryLower) || matchesQuery(f.WhyFailed, queryLower) || matchesQuery(f.LessonsLearned, queryLower) {
							results = append(results, fmt.Sprintf("[%s] FAILED: %s | Why: %s | Lesson: %s",
								entry.Timestamp, f.Approach, f.WhyFailed, f.LessonsLearned))
						}
					}
				}
				if !searchFailed && !searchDecision {
					for _, i := range entry.Context.KeyInsights {
						if matchesQuery(i.Insight, queryLower) || matchesQuery(i.Impact, queryLower) {
							results = append(results, fmt.Sprintf("[%s] INSIGHT: %s | Impact: %s",
								entry.Timestamp, i.Insight, i.Impact))
						}
					}
				}
			}
		}

		if len(results) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No results found for query: %s", query)), nil
		}

		output := fmt.Sprintf("Found %d results for \"%s\":\n\n%s", len(results), query, strings.Join(results, "\n"))
		return mcp.NewToolResultText(output), nil
	})
}

func addGuideTool(s *server.MCPServer) {
	tool := mcp.NewTool("guide",
		mcp.WithDescription("Get detailed guides on checkpoint workflow, best practices, LLM integration"),
		mcp.WithString("topic", mcp.Required(),
			mcp.Description("Guide topic: first-time-user, llm-workflow, best-practices"),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		topic := req.GetString("topic", "")

		guide, ok := guides.GetGuide(topic)
		if !ok {
			available := guides.ListGuides()
			return mcp.NewToolResultError(fmt.Sprintf("Guide '%s' not found. Available guides: %s",
				topic, strings.Join(available, ", "))), nil
		}

		return mcp.NewToolResultText(guide.Content), nil
	})
}

func addStatusTool(s *server.MCPServer, projectPath string) {
	tool := mcp.NewTool("status",
		mcp.WithDescription("Get project checkpoint status - recent activity, next steps, pending work"),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var sb strings.Builder
		sb.WriteString("# Checkpoint Status\n\n")

		// Count changelog entries
		changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
		if data, err := os.ReadFile(changelogPath); err == nil {
			entries := parseChangelogEntries(data)
			sb.WriteString(fmt.Sprintf("Total checkpoints: %d\n", len(entries)))

			// Show last entry info
			if len(entries) > 0 {
				last := entries[len(entries)-1]
				sb.WriteString(fmt.Sprintf("Last checkpoint: %s\n", last.Timestamp))
				if last.CommitHash != "" {
					sb.WriteString(fmt.Sprintf("Last commit: %s\n", last.CommitHash))
				}
				if len(last.Changes) > 0 {
					sb.WriteString("\nLast changes:\n")
					for _, c := range last.Changes {
						sb.WriteString(fmt.Sprintf("  - [%s] %s\n", c.ChangeType, c.Summary))
					}
				}
			}
		} else {
			sb.WriteString("No changelog found. Run 'checkpoint init' to get started.\n")
		}

		// Read status file for next steps
		statusPath := filepath.Join(projectPath, config.StatusFileName)
		if data, err := os.ReadFile(statusPath); err == nil {
			nextSteps := schema.ExtractNextStepsFromStatus(string(data))
			if len(nextSteps) > 0 {
				sb.WriteString("\nNext steps:\n")
				for _, step := range nextSteps {
					priority := ""
					if step.Priority != "" {
						priority = fmt.Sprintf(" [%s]", step.Priority)
					}
					sb.WriteString(fmt.Sprintf("  - %s%s\n", step.Summary, priority))
				}
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
	})
}

func addContextTemplateTool(s *server.MCPServer) {
	tool := mcp.NewTool("context_template",
		mcp.WithDescription("Get the context template for filling checkpoint input"),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(cpcontext.GenerateContextTemplate()), nil
	})
}

// parseChangelogEntries parses multi-doc YAML changelog into checkpoint entries,
// skipping meta documents.
func parseChangelogEntries(data []byte) []schema.CheckpointEntry {
	var entries []schema.CheckpointEntry
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var entry schema.CheckpointEntry
		if err := decoder.Decode(&entry); err != nil {
			break
		}
		// Skip meta documents and empty entries
		if entry.SchemaVersion != "" && len(entry.Changes) > 0 {
			entries = append(entries, entry)
		}
	}
	return entries
}

// parseContextEntries parses multi-doc YAML context file.
func parseContextEntries(data []byte) []cpcontext.ContextEntry {
	var entries []cpcontext.ContextEntry
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var entry cpcontext.ContextEntry
		if err := decoder.Decode(&entry); err != nil {
			break
		}
		if entry.SchemaVersion != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}

// matchesQuery performs case-insensitive substring matching.
func matchesQuery(text, queryLower string) bool {
	return text != "" && strings.Contains(strings.ToLower(text), queryLower)
}
