package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage checkpoint configuration files",
	Long: `Read and modify checkpoint configuration files programmatically.

Subcommands:
  get <file>                 Read a config file as JSON
  set <file> <path> <value>  Update a config value
  list                       List available config files`,
}

var configGetCmd = &cobra.Command{
	Use:   "get <file>",
	Short: "Read a config file as JSON",
	Long: `Read any .checkpoint/*.yml file and output as JSON.

Examples:
  checkpoint config get project.yml
  checkpoint config get tools.yml
  checkpoint config get guidelines.yml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		configGet(absPath, args[0])
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <file> <path> <value>",
	Short: "Update a config value",
	Long: `Update a configuration value using dot-notation path.

Examples:
  checkpoint config set project.yml name "My Project"
  checkpoint config set tools.yml build.default.command "make build"
  checkpoint config set guidelines.yml rules[0] "New rule"`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		configSet(absPath, args[0], args[1], args[2])
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available config files",
	Long:  `List all configuration files in .checkpoint/ directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		configList(absPath)
	},
}

func configGet(projectPath string, filename string) {
	// Resolve file path
	var filePath string
	if strings.HasPrefix(filename, "/") {
		filePath = filename
	} else {
		filePath = filepath.Join(projectPath, config.CheckpointDir, filename)
	}

	// Read YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot read %s: %v\n", filename, err)
		os.Exit(1)
	}

	// Parse YAML
	var content any
	if err := yaml.Unmarshal(data, &content); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot parse %s: %v\n", filename, err)
		os.Exit(1)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot convert to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}

func configSet(projectPath string, filename string, path string, value string) {
	// Resolve file path
	var filePath string
	if strings.HasPrefix(filename, "/") {
		filePath = filename
	} else {
		filePath = filepath.Join(projectPath, config.CheckpointDir, filename)
	}

	// Read existing YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot read %s: %v\n", filename, err)
		os.Exit(1)
	}

	// Parse YAML into map
	var content map[string]any
	if err := yaml.Unmarshal(data, &content); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot parse %s: %v\n", filename, err)
		os.Exit(1)
	}

	// Parse the path and set the value
	if err := setNestedValue(content, path, value); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Marshal back to YAML
	newData, err := yaml.Marshal(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot serialize config: %v\n", err)
		os.Exit(1)
	}

	// Write back
	if err := os.WriteFile(filePath, newData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot write %s: %v\n", filename, err)
		os.Exit(1)
	}

	fmt.Printf("Updated %s: %s = %s\n", filename, path, value)
}

// setNestedValue sets a value in a nested map using dot-notation path
// Supports paths like "build.default.command" or "rules[0]"
func setNestedValue(data map[string]any, path string, value string) error {
	parts := parsePath(path)
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	// Navigate to the parent
	current := any(data)
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		switch c := current.(type) {
		case map[string]any:
			if next, ok := c[part.key]; ok {
				current = next
			} else {
				// Create missing intermediate map
				newMap := make(map[string]any)
				c[part.key] = newMap
				current = newMap
			}
		case []any:
			if part.index < 0 || part.index >= len(c) {
				return fmt.Errorf("array index %d out of bounds", part.index)
			}
			current = c[part.index]
		default:
			return fmt.Errorf("cannot navigate through %T at %s", current, part.key)
		}
	}

	// Set the final value
	lastPart := parts[len(parts)-1]
	switch c := current.(type) {
	case map[string]any:
		c[lastPart.key] = value
	case []any:
		if lastPart.index < 0 || lastPart.index >= len(c) {
			return fmt.Errorf("array index %d out of bounds", lastPart.index)
		}
		c[lastPart.index] = value
	default:
		return fmt.Errorf("cannot set value in %T", current)
	}

	return nil
}

type pathPart struct {
	key   string
	index int // -1 if not an array access
}

// parsePath parses a dot-notation path with optional array indices
// e.g., "build.default.command" or "rules[0]"
func parsePath(path string) []pathPart {
	var parts []pathPart
	var current strings.Builder

	for i := 0; i < len(path); i++ {
		c := path[i]
		switch c {
		case '.':
			if current.Len() > 0 {
				parts = append(parts, pathPart{key: current.String(), index: -1})
				current.Reset()
			}
		case '[':
			if current.Len() > 0 {
				parts = append(parts, pathPart{key: current.String(), index: -1})
				current.Reset()
			}
			// Parse index
			var idx int
			i++
			for i < len(path) && path[i] != ']' {
				idx = idx*10 + int(path[i]-'0')
				i++
			}
			parts = append(parts, pathPart{key: "", index: idx})
		default:
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, pathPart{key: current.String(), index: -1})
	}

	return parts
}

func configList(projectPath string) {
	checkpointDir := filepath.Join(projectPath, config.CheckpointDir)

	entries, err := os.ReadDir(checkpointDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot read .checkpoint directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration files in .checkpoint/:")
	fmt.Println()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			info, _ := entry.Info()
			size := ""
			if info != nil {
				size = fmt.Sprintf(" (%d bytes)", info.Size())
			}
			fmt.Printf("  %s%s\n", name, size)
		}
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  checkpoint config get <file>              Read as JSON")
	fmt.Println("  checkpoint config set <file> <path> <v>   Update value")
}
