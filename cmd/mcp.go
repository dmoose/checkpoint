package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"

	"github.com/dmoose/checkpoint/internal/prompts"
	"github.com/dmoose/checkpoint/pkg/config"
)

// MCP starts a stdio MCP server that serves read-only project data.
// Roots precedence: --root flags (passed via options) > CHECKPOINT_ROOTS env (comma-separated) > ~/.config/checkpoint/config.json {"roots":[]}
// Each call requires params.project_id; server resolves the project directory by scanning roots for .checkpoint-status.yaml.
// On each call, the cached project_id->path mapping is verified against the status file; on mismatch, roots are rescanned once.
// Responses include project_id for sanity checking.
func MCP(options MCPOptions) error {
	roots := resolveRoots(options.Roots)
	if len(roots) == 0 {
		return fmt.Errorf("no roots configured; set --root, CHECKPOINT_ROOTS, or ~/.config/checkpoint/config.json")
	}

	// Create service with cache and roots
	svc := &mcpService{
		roots: roots,
		cache: make(map[string]string),
	}
	// Initial lazy: do not pre-scan; scan on first miss

	s := server.NewMCPServer("checkpoint-mcp", "0.1.0",
		server.WithToolCapabilities(true),
	)

	// Define the 'project' tool
	s.AddTool(
		mcp.NewTool("project",
			mcp.WithDescription("Return structured project info by project_id"),
			mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ULID from .checkpoint-status.yaml")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectID, err := req.RequireString("project_id")
			if err != nil {
				return mcp.NewToolResultError("missing project_id"), nil
			}

			info, err := svc.getProjectInfo(projectID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			b, _ := json.Marshal(info)
			return mcp.NewToolResultText(string(b)), nil
		},
	)

	return server.ServeStdio(s)
}

type MCPOptions struct {
	Roots []string // repeatable --root
}

type mcpService struct {
	mu    sync.Mutex
	roots []string
	cache map[string]string // project_id -> abs path of dir containing status file
}

type projectInfo struct {
	ProjectID               string           `json:"project_id"`
	ProjectRoot             string           `json:"project_root"`
	PathHash                string           `json:"path_hash"`
	CheckpointCount         int              `json:"checkpoint_count"`
	LastCheckpointTimestamp string           `json:"last_checkpoint_timestamp,omitempty"`
	NextSteps               []map[string]any `json:"next_steps,omitempty"`
	PromptIDs               []string         `json:"prompt_ids,omitempty"`
}

func (s *mcpService) getProjectInfo(projectID string) (*projectInfo, error) {
	// Resolve path from cache or rescan
	path, ok := s.lookup(projectID)
	if !ok {
		if err := s.rescan(projectID); err != nil {
			return nil, err
		}
		var ok2 bool
		path, ok2 = s.lookup(projectID)
		if !ok2 {
			return nil, fmt.Errorf("not_found: project_id %s", projectID)
		}
	}

	// Verify status still matches
	statusPath := filepath.Join(path, config.StatusFileName)
	sid, ph, lastTS, steps, err := readStatus(statusPath)
	if err != nil {
		// status missing or unreadable: force rescan
		if err := s.rescan(projectID); err != nil {
			return nil, err
		}
		path2, ok3 := s.lookup(projectID)
		if !ok3 {
			return nil, fmt.Errorf("not_found: project_id %s", projectID)
		}
		path = path2
		statusPath = filepath.Join(path, config.StatusFileName)
		sid, ph, lastTS, steps, err = readStatus(statusPath)
		if err != nil {
			return nil, fmt.Errorf("status_mismatch: %v", err)
		}
	}
	if sid != projectID {
		// Cache stale; rescan
		if err := s.rescan(projectID); err != nil {
			return nil, err
		}
		p2, ok4 := s.lookup(projectID)
		if !ok4 {
			return nil, fmt.Errorf("status_mismatch: cached path no longer matches and project not found")
		}
		path = p2
	}

	// Compute extras
	cc := countChangelog(filepath.Join(path, config.ChangelogFileName))
	promptIDs := listPromptIDs(filepath.Join(path, ".checkpoint", "prompts"))

	// Convert next steps to generic maps to keep handler simple
	var ns []map[string]any
	for _, st := range steps {
		m := map[string]any{"summary": st["summary"]}
		if v, ok := st["details"]; ok {
			m["details"] = v
		}
		if v, ok := st["priority"]; ok {
			m["priority"] = v
		}
		if v, ok := st["scope"]; ok {
			m["scope"] = v
		}
		ns = append(ns, m)
	}

	return &projectInfo{
		ProjectID:               projectID,
		ProjectRoot:             path,
		PathHash:                ph,
		CheckpointCount:         cc,
		LastCheckpointTimestamp: lastTS,
		NextSteps:               ns,
		PromptIDs:               promptIDs,
	}, nil
}

func (s *mcpService) lookup(projectID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.cache[projectID]
	return p, ok
}

func (s *mcpService) rescan(targetProjectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Build fresh map; if target found multiple times, error
	foundPaths := make(map[string]string)
	for _, root := range s.roots {
		abs := root
		if !filepath.IsAbs(abs) {
			if a2, err := filepath.Abs(abs); err == nil {
				abs = a2
			}
		}
		walkRoot := abs
		filepath.WalkDir(walkRoot, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			statusPath := filepath.Join(p, config.StatusFileName)
			if _, err := os.Stat(statusPath); err == nil {
				pid, _, _, _, err := readStatus(statusPath)
				if err == nil {
					if _, exists := foundPaths[pid]; exists {
						// duplicate project_id found
						foundPaths[pid] = "__DUPLICATE__"
					} else {
						foundPaths[pid] = p
					}
				}
				// Do not descend further from a project root
				return filepath.SkipDir
			}
			return nil
		})
	}
	// Check duplicate for target if provided
	if targetProjectID != "" {
		if v, ok := foundPaths[targetProjectID]; ok && v == "__DUPLICATE__" {
			return fmt.Errorf("duplicate_project_id: %s", targetProjectID)
		}
	}
	// Replace cache
	s.cache = map[string]string{}
	for k, v := range foundPaths {
		if v != "__DUPLICATE__" {
			s.cache[k] = v
		}
	}
	return nil
}

func resolveRoots(flagRoots []string) []string {
	roots := make([]string, 0, 8)
	// 1) flags
	for _, r := range flagRoots {
		if r != "" {
			roots = append(roots, r)
		}
	}
	if len(roots) > 0 {
		return dedupeAbs(roots)
	}
	// 2) env
	if env := strings.TrimSpace(os.Getenv("CHECKPOINT_ROOTS")); env != "" {
		parts := strings.Split(env, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				roots = append(roots, p)
			}
		}
		if len(roots) > 0 {
			return dedupeAbs(roots)
		}
	}
	// 3) config file
	home, _ := os.UserHomeDir()
	cfg := filepath.Join(home, ".config", "checkpoint", "config.json")
	if b, err := os.ReadFile(cfg); err == nil {
		var aux struct {
			Roots []string `json:"roots"`
		}
		if json.Unmarshal(b, &aux) == nil {
			for _, r := range aux.Roots {
				if r != "" {
					roots = append(roots, r)
				}
			}
		}
	}
	return dedupeAbs(roots)
}

func dedupeAbs(in []string) []string {
	m := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, r := range in {
		abs := r
		if !filepath.IsAbs(abs) {
			if a2, err := filepath.Abs(abs); err == nil {
				abs = a2
			}
		}
		if _, ok := m[abs]; !ok {
			m[abs] = struct{}{}
			out = append(out, abs)
		}
	}
	sort.Strings(out)
	return out
}

// readStatus parses .checkpoint-status.yaml and returns (project_id, path_hash, last_commit_timestamp, next_steps)
func readStatus(statusPath string) (string, string, string, []map[string]any, error) {
	b, err := os.ReadFile(statusPath)
	if err != nil {
		return "", "", "", nil, err
	}
	var aux struct {
		ProjectID string           `yaml:"project_id"`
		PathHash  string           `yaml:"path_hash"`
		LastTS    string           `yaml:"last_commit_timestamp"`
		NextSteps []map[string]any `yaml:"next_steps"`
	}
	if err := yaml.Unmarshal(b, &aux); err != nil {
		return "", "", "", nil, err
	}
	if aux.ProjectID == "" {
		return "", "", "", nil, errors.New("status missing project_id")
	}
	return aux.ProjectID, aux.PathHash, aux.LastTS, aux.NextSteps, nil
}

func countChangelog(changelogPath string) int {
	b, err := os.ReadFile(changelogPath)
	if err != nil {
		return 0
	}
	content := string(b)
	sep := 0
	for _, ln := range strings.Split(content, "\n") {
		if strings.TrimSpace(ln) == "---" {
			sep++
		}
	}
	if sep > 1 {
		return sep - 1
	}
	return 0
}

func listPromptIDs(promptsDir string) []string {
	// prompts.yaml is expected in promptsDir
	cfg, err := prompts.LoadPromptsConfig(promptsDir)
	if err != nil {
		return nil
	}
	infos := prompts.ListPrompts(cfg)
	ids := make([]string, 0, len(infos))
	for _, p := range infos {
		ids = append(ids, p.ID)
	}
	sort.Strings(ids)
	return ids
}
