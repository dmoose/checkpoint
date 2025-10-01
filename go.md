# Go Application - File Structure & Code

I'll provide a complete, compilable Go application structure.

## Directory Structure

```
checkpoint-tool/
├── go.mod
├── go.sum
├── main.go
├── cmd/
│   ├── check.go
│   ├── commit.go
│   └── help.go
├── internal/
│   ├── schema/
│   │   └── schema.go
│   ├── git/
│   │   └── git.go
│   ├── file/
│   │   └── file.go
│   └── changelog/
│       └── changelog.go
├── pkg/
│   └── config/
│       └── config.go
└── README.md
```

---

## go.mod

```go
module github.com/yourusername/checkpoint-tool

go 1.21

require gopkg.in/yaml.v3 v3.0.1
```

---

## main.go

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/checkpoint-tool/cmd"
	"github.com/yourusername/checkpoint-tool/pkg/config"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		cmd.Help()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	// Parse flags for path argument
	fs := flag.NewFlagSet(subcommand, flag.ExitOnError)
	fs.Usage = func() {}
	pathFlag := fs.String("path", "", "path to project (optional, uses current dir if git root)")

	// Manual path parsing from args
	var projectPath string
	if len(args) > 0 && !args[0][:1][0:1] == "-" {
		projectPath = args[0]
	} else {
		fs.Parse(args)
		if *pathFlag != "" {
			projectPath = *pathFlag
		}
	}

	// Determine working directory
	if projectPath == "" {
		projectPath = "."
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot resolve path: %v\n", err)
		os.Exit(1)
	}

	switch subcommand {
	case "check":
		cmd.Check(absPath)
	case "commit":
		cmd.Commit(absPath)
	case "help", "-h", "--help":
		cmd.Help()
	case "version", "-v", "--version":
		fmt.Printf("checkpoint-tool version %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", subcommand)
		cmd.Help()
		os.Exit(1)
	}
}
```

---

## cmd/help.go

```go
package cmd

import "fmt"

func Help() {
	fmt.Println(`checkpoint-tool - LLM-assisted development checkpoint tracking

USAGE:
  checkpoint-tool <command> [options] [path]

COMMANDS:
  check       Generate input file for LLM (creates .checkpoint-input and .checkpoint-diff)
  commit      Parse input file, create changelog entry, and commit changes
  help        Display this help message
  version     Display version information

OPTIONS:
  [path]      Path to git repository (optional; uses current directory if git root)

EXAMPLES:
  checkpoint-tool check
  checkpoint-tool check /path/to/project
  checkpoint-tool commit
  checkpoint-tool commit /path/to/project

WORKFLOW:
  1. Run 'check' to generate input file
  2. Share input with LLM or edit manually
  3. Review LLM output and edit if needed
  4. Run 'commit' to finalize and append to changelog

For more information, see README.md
`)
}
```

---

## cmd/check.go

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/checkpoint-tool/internal/changelog"
	"github.com/yourusername/checkpoint-tool/internal/file"
	"github.com/yourusername/checkpoint-tool/internal/git"
	"github.com/yourusername/checkpoint-tool/internal/schema"
	"github.com/yourusername/checkpoint-tool/pkg/config"
)

func Check(projectPath string) {
	// Validate git repository
	if !git.IsGitRepository(projectPath) {
		fmt.Fprintf(os.Stderr, "Error: %s is not a git repository\n", projectPath)
		os.Exit(1)
	}

	// Check if input file already exists
	inputPath := filepath.Join(projectPath, config.InputFileName)
	if file.Exists(inputPath) {
		fmt.Fprintf(os.Stderr, "Error: input file already exists at %s\n", inputPath)
		fmt.Fprintf(os.Stderr, "Another checkpoint may be in progress. Run 'commit' to finalize it.\n")
		os.Exit(1)
	}

	// Get git status
	status, err := git.GetStatus(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get git status: %v\n", err)
		os.Exit(1)
	}

	// Get git diff
	diffOutput, err := git.GetDiff(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get git diff: %v\n", err)
		os.Exit(1)
	}

	// Write diff file
	diffPath := filepath.Join(projectPath, config.DiffFileName)
	if err := file.WriteFile(diffPath, diffOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write diff file: %v\n", err)
		os.Exit(1)
	}

	// Generate input file
	inputContent := schema.GenerateInputTemplate(status, config.DiffFileName)
	if err := file.WriteFile(inputPath, inputContent); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write input file: %v\n", err)
		// Cleanup diff file on failure
		os.Remove(diffPath)
		os.Exit(1)
	}

	fmt.Printf("✓ Checkpoint input generated successfully\n")
	fmt.Printf("\nInput file: %s\n", inputPath)
	fmt.Printf("Diff file: %s\n\n", diffPath)
	fmt.Printf("Next steps:\n")
	fmt.Printf("1. Review the input file and share with LLM, or edit directly\n")
	fmt.Printf("2. LLM should follow the instructions at the top of the input file\n")
	fmt.Printf("3. Once satisfied with the result, run: checkpoint-tool commit %s\n", projectPath)
}
```

---

## cmd/commit.go

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yourusername/checkpoint-tool/internal/changelog"
	"github.com/yourusername/checkpoint-tool/internal/file"
	"github.com/yourusername/checkpoint-tool/internal/git"
	"github.com/yourusername/checkpoint-tool/internal/schema"
	"github.com/yourusername/checkpoint-tool/pkg/config"
)

func Commit(projectPath string) {
	// Validate git repository
	if !git.IsGitRepository(projectPath) {
		fmt.Fprintf(os.Stderr, "Error: %s is not a git repository\n", projectPath)
		os.Exit(1)
	}

	// Check if input file exists
	inputPath := filepath.Join(projectPath, config.InputFileName)
	if !file.Exists(inputPath) {
		fmt.Fprintf(os.Stderr, "Error: input file not found at %s\n", inputPath)
		fmt.Fprintf(os.Stderr, "Run 'checkpoint-tool check %s' first\n", projectPath)
		os.Exit(1)
	}

	// Read and parse input file
	inputContent, err := file.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read input file: %v\n", err)
		os.Exit(1)
	}

	// Parse YAML input
	entry, err := schema.ParseInputFile(inputContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse input file: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please fix the errors and try again\n")
		os.Exit(1)
	}

	// Validate required fields
	if err := schema.ValidateEntry(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Error: validation failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please fill in all required fields\n")
		os.Exit(1)
	}

	// Add current timestamp if not present
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Render changelog entry
	renderedEntry := schema.RenderChangelogEntry(entry)

	// Append to changelog file
	changelogPath := filepath.Join(projectPath, config.ChangelogFileName)
	if err := changelog.AppendEntry(changelogPath, renderedEntry); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to append to changelog: %v\n", err)
		os.Exit(1)
	}

	// Stage and commit
	commitMessage := schema.GenerateCommitMessage(entry)
	commitHash, err := git.StageAndCommit(projectPath, commitMessage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to commit: %v\n", err)
		fmt.Fprintf(os.Stderr, "Changelog has been appended but not committed\n")
		fmt.Fprintf(os.Stderr, "Please resolve git issues and commit manually\n")
		os.Exit(1)
	}

	// Update entry with commit hash
	entry.CommitHash = commitHash

	// Write status file
	statusPath := filepath.Join(projectPath, config.StatusFileName)
	statusContent := schema.GenerateStatusFile(entry, commitMessage)
	if err := file.WriteFile(statusPath, statusContent); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write status file: %v\n", err)
		// Non-fatal; status file is informational
	}

	// Clean up input file
	if err := os.Remove(inputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove input file: %v\n", err)
	}

	// Clean up diff file
	diffPath := filepath.Join(projectPath, config.DiffFileName)
	if file.Exists(diffPath) {
		if err := os.Remove(diffPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove diff file: %v\n", err)
		}
	}

	fmt.Printf("✓ Checkpoint committed successfully\n")
	fmt.Printf("\nCommit hash: %s\n", commitHash)
	fmt.Printf("Summary: %s\n", entry.Summary)
	fmt.Printf("\nYou can now create a new checkpoint with: checkpoint-tool check %s\n", projectPath)
}
```

---

## internal/schema/schema.go

```go
package schema

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ChangelogEntry represents a single changelog entry
type ChangelogEntry struct {
	SchemaVersion string `yaml:"schema_version"`
	Timestamp     string `yaml:"timestamp"`
	Summary       string `yaml:"summary"`
	Details       string `yaml:"details,omitempty"`
	ChangeType    string `yaml:"change_type"`
	Scope         string `yaml:"scope,omitempty"`
	CommitHash    string `yaml:"commit_hash,omitempty"`
	GitStatus     string `yaml:"git_status,omitempty"`
	DiffFile      string `yaml:"diff_file,omitempty"`
}

const (
	SchemaVersion = "1"
	LLMPrompt     = `# INSTRUCTIONS FOR LLM:
# Fill in the fields below to describe the changes made.
# Follow the existing schema structure.
# Be concise but descriptive.
# Do not modify the schema_version or timestamp.
#
# Required fields: summary, change_type
# Optional fields: details, scope
#
# change_type must be one of: feature, fix, refactor, docs, perf, other
`
)

// GenerateInputTemplate creates a template input file for the LLM
func GenerateInputTemplate(gitStatus, diffFileName string) string {
	timestamp := time.Now().Format(time.RFC3339)

	template := fmt.Sprintf(`%s

schema_version: "%s"
timestamp: "%s"

# Git status output (informational):
git_status: |
%s

# Reference to diff context (path to git diff output):
diff_file: "%s"

# SCHEMA FIELDS - Fill these in:
summary: "[FILL IN: one-line summary of changes]"
details: "[FILL IN: longer description of what was changed and why]"
change_type: "[FILL IN: feature|fix|refactor|docs|perf|other]"
scope: "[FILL IN: component or module affected]"
commit_hash: ""  # Leave empty; tool will fill this
`, LLMPrompt, SchemaVersion, timestamp, indentGitStatus(gitStatus), diffFileName)

	return template
}

// ParseInputFile parses a YAML input file into a ChangelogEntry
func ParseInputFile(content string) (*ChangelogEntry, error) {
	// Remove the LLM prompt section if present
	content = stripPrompt(content)

	entry := &ChangelogEntry{}
	if err := yaml.Unmarshal([]byte(content), entry); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return entry, nil
}

// ValidateEntry checks that required fields are present
func ValidateEntry(entry *ChangelogEntry) error {
	var missing []string

	if entry.SchemaVersion == "" {
		missing = append(missing, "schema_version")
	}
	if entry.Timestamp == "" {
		missing = append(missing, "timestamp")
	}
	if entry.Summary == "" {
		missing = append(missing, "summary")
	}
	if entry.ChangeType == "" {
		missing = append(missing, "change_type")
	}

	// Validate change_type enum
	validTypes := map[string]bool{
		"feature":  true,
		"fix":      true,
		"refactor": true,
		"docs":     true,
		"perf":     true,
		"other":    true,
	}
	if !validTypes[entry.ChangeType] {
		return fmt.Errorf("invalid change_type '%s'; must be one of: feature, fix, refactor, docs, perf, other", entry.ChangeType)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

// RenderChangelogEntry renders a changelog entry as YAML
func RenderChangelogEntry(entry *ChangelogEntry) string {
	if entry.SchemaVersion == "" {
		entry.SchemaVersion = SchemaVersion
	}

	output, err := yaml.Marshal(entry)
	if err != nil {
		return ""
	}

	// Add delimiter for entry separation
	return fmt.Sprintf("---\n%s", string(output))
}

// GenerateCommitMessage creates a git commit message from entry
func GenerateCommitMessage(entry *ChangelogEntry) string {
	message := fmt.Sprintf("Checkpoint: %s - %s", entry.ChangeType, entry.Summary)
	if entry.Scope != "" {
		message = fmt.Sprintf("Checkpoint: %s (%s) - %s", entry.ChangeType, entry.Scope, entry.Summary)
	}
	return message
}

// GenerateStatusFile creates status file content
func GenerateStatusFile(entry *ChangelogEntry, commitMessage string) string {
	status := fmt.Sprintf(`last_commit_hash: "%s"
last_commit_timestamp: "%s"
last_commit_message: "%s"
status: "success"
`, entry.CommitHash, entry.Timestamp, commitMessage)
	return status
}

// Helper functions

func indentGitStatus(status string) string {
	lines := strings.Split(status, "\n")
	var indented []string
	for _, line := range lines {
		if line != "" {
			indented = append(indented, "  "+line)
		}
	}
	return strings.Join(indented, "\n")
}

func stripPrompt(content string) string {
	// Find the first YAML line (starts with schema_version:)
	lines := strings.Split(content, "\n")
	var startIdx int
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "schema_version:") {
			startIdx = i
			break
		}
	}
	return strings.Join(lines[startIdx:], "\n")
}
```

---

## internal/git/git.go

```go
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsGitRepository checks if a path is a git repository
func IsGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	return err == nil && info.IsDir()
}

// GetStatus returns git status output
func GetStatus(path string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	return out.String(), nil
}

// GetDiff returns git diff output for staged and unstaged changes
func GetDiff(path string) (string, error) {
	// Get both staged and unstaged changes
	cmd := exec.Command("git", "diff", "HEAD")
	cmd.Dir = path

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		// It's okay if this fails (e.g., no HEAD yet), return what we have
		return out.String(), nil
	}

	return out.String(), nil
}

// StageAndCommit stages all changes and commits with the given message
func StageAndCommit(path, message string) (string, error) {
	// Add all changes
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = path

	if err := addCmd.Run(); err != nil {
		return "", fmt.Errorf("git add failed: %w", err)
	}

	// Commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = path

	var out bytes.Buffer
	commitCmd.Stdout = &out
	commitCmd.Stderr = &out

	if err := commitCmd.Run(); err != nil {
		return "", fmt.Errorf("git commit failed: %w", err)
	}

	// Get commit hash
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = path

	var hashOut bytes.Buffer
	hashCmd.Stdout = &hashOut
	hashCmd.Stderr = &hashOut

	if err := hashCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	return strings.TrimSpace(hashOut.String()), nil
}
```

---

## internal/file/file.go

```go
package file

import (
	"fmt"
	"os"
)

// ReadFile reads a file and returns its contents
func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

// WriteFile writes content to a file
func WriteFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// Exists checks if a file exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// AppendFile appends content to a file
func AppendFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to append to file: %w", err)
	}

	return nil
}
```

---

## internal/changelog/changelog.go

```go
package changelog

import (
	"fmt"
	"os"

	"github.com/yourusername/checkpoint-tool/internal/file"
)

// AppendEntry appends a changelog entry to the changelog file
func AppendEntry(path, entry string) error {
	// Ensure entry ends with newline
	if len(entry) > 0 && entry[len(entry)-1] != '\n' {
		entry += "\n"
	}

	if err := file.AppendFile(path, entry); err != nil {
		return fmt.Errorf("failed to append changelog entry: %w", err)
	}

	return nil
}

// ReadEntries reads all entries from changelog file (for Mac app)
func ReadEntries(path string) (string, error) {
	if !file.Exists(path) {
		return "", nil
	}

	content, err := file.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read changelog: %w", err)
	}

	return content, nil
}

// GetLastModified gets the modification time of the changelog
func GetLastModified(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return info.ModTime().Unix(), nil
}
```

---

## pkg/config/config.go

```go
package config

// Configuration constants for checkpoint tool

const (
	// File names
	ChangelogFileName = ".changelog"
	InputFileName     = ".checkpoint-input"
	DiffFileName      = ".checkpoint-diff"
	StatusFileName    = ".checkpoint-status"
	LockFileName      = ".checkpoint-lock"

	// Git operations
	DefaultCommitPrefix = "Checkpoint"
)

// GlobPatterns for scanning
var ChangelogGlobPattern = "*/.changelog"
```

---

## README.md

```markdown
# Checkpoint Tool

A Go CLI tool for structured change tracking in LLM-assisted development projects.

## Installation

### From Source

```bash
git clone https://github.com/yourusername/checkpoint-tool.git
cd checkpoint-tool
go build -o checkpoint-tool
```

Add to your PATH or run directly: `./checkpoint-tool`

### Requirements

- Go 1.21 or later
- Git

## Usage

### Basic Workflow

```bash
# 1. Create a checkpoint (generates input file for LLM)
checkpoint-tool check /path/to/project

# 2. Share input file with LLM or edit manually

# 3. Finalize checkpoint (parses input, commits, cleans up)
checkpoint-tool commit /path/to/project
```

### Commands

#### `check [path]`

Generates checkpoint input file for review/editing.

**What it does:**
- Validates git repository
- Captures git status
- Captures git diff
- Generates `.checkpoint-input` with template
- Generates `.checkpoint-diff` for reference

**Usage:**
```bash
checkpoint-tool check
checkpoint-tool check /path/to/project
```

#### `commit [path]`

Parses input file, appends to changelog, and commits changes.

**What it does:**
- Validates input file exists
- Parses input YAML
- Validates required fields
- Renders changelog entry
- Stages and commits to git
- Writes status file
- Cleans up input/diff files

**Usage:**
```bash
checkpoint-tool commit
checkpoint-tool commit /path/to/project
```

#### `help`

Displays help information.

```bash
checkpoint-tool help
```

## File Structure

After running `check` and before running `commit`, your project will contain:

```
project-root/
├── .checkpoint-input     # Input file (edit this)
├── .checkpoint-diff      # Git diff output (reference)
└── .changelog            # Append-only changelog (created after first commit)
```

All checkpoint files are git-ignored by default.

## Workflow Details

### Input File Format

The input file is YAML with embedded instructions for the LLM:

```yaml
# INSTRUCTIONS FOR LLM:
# Fill in the fields below...

schema_version: "1"
timestamp: "2025-10-21T16:15:00-07:00"

git_status: |
  M  src/auth.go
  M  src/config.go

diff_file: ".checkpoint-diff"

# SCHEMA FIELDS - Fill these in:
summary: "[FILL IN: one-line summary]"
details: "[FILL IN: longer description]"
change_type: "[FILL IN: feature|fix|refactor|docs|perf|other]"
scope: "[FILL IN: component affected]"
commit_hash: ""
```

### Changelog Format

The `.changelog` file is append-only YAML with entries separated by `---`:

```yaml
---
schema_version: "1"
timestamp: "2025-10-21T16:15:00-07:00"
summary: "Implemented user authentication flow"
details: "Added JWT token validation middleware..."
change_type: "feature"
scope: "auth"
commit_hash: "abc123def456"
---
schema_version: "1"
timestamp: "2025-10-21T16:05:00-07:00"
summary: "Refactored database query layer"
details: "Moved query logic to separate module..."
change_type: "refactor"
scope: "database"
commit_hash: "xyz789uvw012"
```

## Integration with LLM

### Recommended Workflow

1. Run `checkpoint-tool check`
2. Read the generated `.checkpoint-input` file
3. Share the entire file content with your LLM
4. LLM fills in the `summary`, `details`, `change_type`, and `scope` fields
5. Review the output
6. Edit if needed
7. Run `checkpoint-tool commit`

### Example Prompt

```
Please read the checkpoint input file I'm sharing.
It contains git status and diff information about code changes I just made.
Fill in the SCHEMA FIELDS section at the bottom following these guidelines:
- summary: A clear, one-line description
- details: Explain what changed and why
- change_type: Classify as feature, fix, refactor, docs, perf, or other
- scope: What component or module was affected

Leave the other fields unchanged. Return only the modified YAML.
```

## Error Handling

### Common Issues

**"Input file already exists"**
- Another checkpoint is in progress
- Run `commit` to finalize it, or manually delete `.checkpoint-input` to start over

**"Input file not found"**
- Run `check` first to generate the input file

**"Validation failed"**
- Required fields are missing or invalid
- Check that `summary`, `change_type` are filled in
- Valid `change_type` values: feature, fix, refactor, docs, perf, other

**"Not a git repository"**
- Run from a git repository root

### Manual Recovery

If something goes wrong:

1. Review `.checkpoint-input` and `.checkpoint-diff` files
2. Fix the input file manually
3. Run `commit` again

Or start over:

```bash
rm .checkpoint-input .checkpoint-diff
checkpoint-tool check
```

## Configuration

### .gitignore

Automatically managed; checkpoint files are ignored:

```
.checkpoint-input
.checkpoint-diff
.checkpoint-status
.checkpoint-lock
```

## Development

### Building from Source

```bash
git clone https://github.com/yourusername/checkpoint-tool.git
cd checkpoint-tool
go build -o checkpoint-tool main.go
```

### Running Tests

```bash
go test ./...
```

### Directory Structure

```
checkpoint-tool/
├── main.go                    # Entry point
├── cmd/
│   ├── check.go              # check subcommand
│   ├── commit.go             # commit subcommand
│   └── help.go               # help subcommand
├── internal/
│   ├── schema/               # Schema and template logic
│   ├── git/                  # Git operations
│   ├── file/                 # File I/O
│   └── changelog/            # Changelog parsing/writing
└── pkg/
    └── config/               # Configuration constants
```

## License

MIT

## Contributing

Contributions welcome. Please submit issues and pull requests.
```

---

## Compilation Instructions

To compile the application:

```bash
# Clone or create the repository with the above structure
cd checkpoint-tool

# Download dependencies
go mod tidy

# Build the executable
go build -o checkpoint-tool

# Test (optional)
./checkpoint-tool help

# Install globally (optional)
go install
```

The compiled binary will be `checkpoint-tool` (or `checkpoint-tool.exe` on Windows).

---

## Testing the Build

```bash
# Initialize a test git repo
mkdir test-project
cd test-project
git init

# Create a test file
echo "test" > test.txt
git add test.txt
git commit -m "initial"

# Create changes
echo "change" >> test.txt

# Test check command
../checkpoint-tool check .

# Review generated .checkpoint-input file
cat .checkpoint-input

# Edit the input file or let LLM process it
# Then test commit command
../checkpoint-tool commit .

# Verify changelog was created
cat .changelog
```

This structure is ready to compile and run. All dependencies are standard library except for `gopkg.in/yaml.v3` which is widely used and stable.
