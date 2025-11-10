package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-llm/internal/explain"
	"go-llm/pkg/config"

	"gopkg.in/yaml.v3"
)

// LearnOptions holds flags for the learn command
type LearnOptions struct {
	Content   string // The content to learn
	Guideline bool   // Add as a guideline rule
	Tool      bool   // Add as a tool
	Avoid     bool   // Add as an anti-pattern
	Principle bool   // Add as a design principle
	Pattern   bool   // Add as a pattern
	ToolName  string // Tool name when adding a tool
}

// Learn captures knowledge during development
func Learn(projectPath string, opts LearnOptions) {
	if opts.Content == "" {
		fmt.Fprintf(os.Stderr, "error: content required\n")
		fmt.Fprintf(os.Stderr, "usage: checkpoint learn <content> [flags]\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		fmt.Fprintf(os.Stderr, "  --guideline     Add as a rule to follow\n")
		fmt.Fprintf(os.Stderr, "  --avoid         Add as an anti-pattern to avoid\n")
		fmt.Fprintf(os.Stderr, "  --principle     Add as a design principle\n")
		fmt.Fprintf(os.Stderr, "  --pattern       Add as an established pattern\n")
		fmt.Fprintf(os.Stderr, "  --tool <name>   Add as a tool command\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  checkpoint learn \"Always validate input at API boundaries\" --guideline\n")
		fmt.Fprintf(os.Stderr, "  checkpoint learn \"Don't use global mutable state\" --avoid\n")
		fmt.Fprintf(os.Stderr, "  checkpoint learn \"make test-race\" --tool race\n")
		os.Exit(1)
	}

	checkpointDir := filepath.Join(projectPath, config.CheckpointDir)

	// Ensure checkpoint is initialized
	if _, err := os.Stat(checkpointDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: checkpoint not initialized\n")
		fmt.Fprintf(os.Stderr, "hint: Run 'checkpoint init' first\n")
		os.Exit(1)
	}

	var err error
	switch {
	case opts.Guideline:
		err = addGuideline(checkpointDir, opts.Content)
	case opts.Avoid:
		err = addAvoid(checkpointDir, opts.Content)
	case opts.Principle:
		err = addPrinciple(checkpointDir, opts.Content)
	case opts.Pattern:
		err = addPattern(checkpointDir, opts.Content)
	case opts.Tool || opts.ToolName != "":
		err = addTool(checkpointDir, opts.ToolName, opts.Content)
	default:
		// Default: add to learnings log
		err = addLearning(checkpointDir, opts.Content)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func addGuideline(checkpointDir, content string) error {
	guidelinesPath := filepath.Join(checkpointDir, config.ExplainGuidelinesYml)

	var guidelines explain.GuidelinesConfig
	if data, err := os.ReadFile(guidelinesPath); err == nil {
		yaml.Unmarshal(data, &guidelines)
	}
	guidelines.SchemaVersion = "1"

	// Check if already exists
	for _, rule := range guidelines.Rules {
		if rule == content {
			fmt.Printf("Rule already exists: %s\n", content)
			return nil
		}
	}

	guidelines.Rules = append(guidelines.Rules, content)

	if err := writeGuidelinesFile(guidelinesPath, &guidelines); err != nil {
		return err
	}

	fmt.Printf("✓ Added rule: %s\n", content)
	return nil
}

func addAvoid(checkpointDir, content string) error {
	guidelinesPath := filepath.Join(checkpointDir, config.ExplainGuidelinesYml)

	var guidelines explain.GuidelinesConfig
	if data, err := os.ReadFile(guidelinesPath); err == nil {
		yaml.Unmarshal(data, &guidelines)
	}
	guidelines.SchemaVersion = "1"

	// Check if already exists
	for _, item := range guidelines.Avoid {
		if item == content {
			fmt.Printf("Anti-pattern already exists: %s\n", content)
			return nil
		}
	}

	guidelines.Avoid = append(guidelines.Avoid, content)

	if err := writeGuidelinesFile(guidelinesPath, &guidelines); err != nil {
		return err
	}

	fmt.Printf("✓ Added anti-pattern: %s\n", content)
	return nil
}

func addPrinciple(checkpointDir, content string) error {
	guidelinesPath := filepath.Join(checkpointDir, config.ExplainGuidelinesYml)

	var guidelines explain.GuidelinesConfig
	if data, err := os.ReadFile(guidelinesPath); err == nil {
		yaml.Unmarshal(data, &guidelines)
	}
	guidelines.SchemaVersion = "1"

	// Check if already exists
	for _, p := range guidelines.Principles {
		if p == content {
			fmt.Printf("Principle already exists: %s\n", content)
			return nil
		}
	}

	guidelines.Principles = append(guidelines.Principles, content)

	if err := writeGuidelinesFile(guidelinesPath, &guidelines); err != nil {
		return err
	}

	fmt.Printf("✓ Added principle: %s\n", content)
	return nil
}

func addPattern(checkpointDir, content string) error {
	guidelinesPath := filepath.Join(checkpointDir, config.ExplainGuidelinesYml)

	var guidelines explain.GuidelinesConfig
	if data, err := os.ReadFile(guidelinesPath); err == nil {
		yaml.Unmarshal(data, &guidelines)
	}
	guidelines.SchemaVersion = "1"

	// For patterns, we'll add to a patterns section in naming (flexible structure)
	// Actually, let's create a separate patterns array if not exists
	// Since the schema uses interface{}, we need to handle this carefully

	// For simplicity, add patterns to principles with a prefix
	patternContent := fmt.Sprintf("Pattern: %s", content)
	for _, p := range guidelines.Principles {
		if p == patternContent || p == content {
			fmt.Printf("Pattern already exists: %s\n", content)
			return nil
		}
	}

	guidelines.Principles = append(guidelines.Principles, patternContent)

	if err := writeGuidelinesFile(guidelinesPath, &guidelines); err != nil {
		return err
	}

	fmt.Printf("✓ Added pattern: %s\n", content)
	return nil
}

func addTool(checkpointDir, name, command string) error {
	if name == "" {
		// Try to extract name from command
		parts := strings.Fields(command)
		if len(parts) > 0 {
			name = parts[len(parts)-1] // Use last word as name
		} else {
			name = "custom"
		}
	}

	toolsPath := filepath.Join(checkpointDir, config.ExplainToolsYml)

	var tools explain.ToolsConfig
	if data, err := os.ReadFile(toolsPath); err == nil {
		yaml.Unmarshal(data, &tools)
	}
	tools.SchemaVersion = "1"

	// Add to maintenance section (most likely place for custom tools)
	if tools.Maintenance == nil {
		tools.Maintenance = make(map[string]explain.ToolCommand)
	}

	if _, exists := tools.Maintenance[name]; exists {
		fmt.Printf("Tool '%s' already exists, updating...\n", name)
	}

	tools.Maintenance[name] = explain.ToolCommand{
		Command: command,
		Notes:   fmt.Sprintf("Added via 'checkpoint learn' on %s", time.Now().Format("2006-01-02")),
	}

	if err := writeToolsFile(toolsPath, &tools); err != nil {
		return err
	}

	fmt.Printf("✓ Added tool '%s': %s\n", name, command)
	return nil
}

func addLearning(checkpointDir, content string) error {
	// Add to a learnings.yml file (append-only log)
	learningsPath := filepath.Join(checkpointDir, "learnings.yml")

	entry := fmt.Sprintf("---\ntimestamp: %s\nlearning: %s\n",
		time.Now().Format(time.RFC3339),
		content)

	f, err := os.OpenFile(learningsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open learnings file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("write learning: %w", err)
	}

	fmt.Printf("✓ Captured learning: %s\n", content)
	fmt.Printf("  (saved to .checkpoint/learnings.yml)\n")
	return nil
}

func writeGuidelinesFile(path string, g *explain.GuidelinesConfig) error {
	data, err := yaml.Marshal(g)
	if err != nil {
		return fmt.Errorf("marshal guidelines: %w", err)
	}

	// Clean up the output
	content := string(data)
	content = strings.ReplaceAll(content, "schema_version: \"1\"\n", "")
	content = "schema_version: \"1\"\n\n" + content

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write guidelines: %w", err)
	}

	return nil
}

func writeToolsFile(path string, t *explain.ToolsConfig) error {
	data, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshal tools: %w", err)
	}

	// Clean up the output
	content := string(data)
	content = strings.ReplaceAll(content, "schema_version: \"1\"\n", "")
	content = "schema_version: \"1\"\n\n" + content

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write tools: %w", err)
	}

	return nil
}
