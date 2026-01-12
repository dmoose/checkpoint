package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/checkpoint/internal/explain"
	"github.com/dmoose/checkpoint/pkg/config"

	"gopkg.in/yaml.v3"
)

// SkillOptions holds flags for the skill command
type SkillOptions struct {
	Action    string // list, show, add, create, or empty for list
	SkillName string // skill name for show/add/create
}

// Skill manages skills for a project
func Skill(projectPath string, opts SkillOptions) {
	switch opts.Action {
	case "", "list":
		listSkills(projectPath)
	case "show":
		showSkill(projectPath, opts.SkillName)
	case "add":
		addSkill(projectPath, opts.SkillName)
	case "create":
		createSkill(projectPath, opts.SkillName)
	default:
		fmt.Fprintf(os.Stderr, "unknown action: %s\n", opts.Action)
		fmt.Fprintf(os.Stderr, "usage: checkpoint skill [list|show|add|create] [name]\n")
		os.Exit(1)
	}
}

func listSkills(projectPath string) {
	ctx, err := explain.LoadExplainContext(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading context: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available Skills")
	fmt.Println()

	// Local skills
	fmt.Println("Local (project-specific):")
	if ctx.Skills != nil && len(ctx.Skills.Local) > 0 {
		for _, name := range ctx.Skills.Local {
			loaded := skillLoaded(ctx.SkillDefs, name)
			status := ""
			if !loaded {
				status = " (not found)"
			}
			fmt.Printf("  - %s%s\n", name, status)
		}
	} else {
		fmt.Println("  (none)")
	}
	fmt.Println()

	// Global skills
	fmt.Println("Global (from ~/.config/checkpoint/skills/):")
	if ctx.Skills != nil && len(ctx.Skills.Global) > 0 {
		for _, name := range ctx.Skills.Global {
			loaded := skillLoaded(ctx.SkillDefs, name)
			status := ""
			if !loaded {
				status = " (not found)"
			}
			fmt.Printf("  - %s%s\n", name, status)
		}
	} else {
		fmt.Println("  (none configured)")
	}
	fmt.Println()

	// Show available global skills not yet added
	availableGlobal := listAvailableGlobalSkills()
	if len(availableGlobal) > 0 {
		fmt.Println("Available global skills (not yet added):")
		currentGlobal := make(map[string]bool)
		if ctx.Skills != nil {
			for _, name := range ctx.Skills.Global {
				currentGlobal[name] = true
			}
		}
		for _, name := range availableGlobal {
			if !currentGlobal[name] {
				fmt.Printf("  - %s\n", name)
			}
		}
		fmt.Println()
	}

	fmt.Println("Commands:")
	fmt.Println("  checkpoint skill show <name>   - View skill details")
	fmt.Println("  checkpoint skill add <name>    - Add global skill to project")
	fmt.Println("  checkpoint skill create <name> - Create new local skill")
}

func skillLoaded(skills []explain.Skill, name string) bool {
	for _, s := range skills {
		if s.Name == name {
			return true
		}
	}
	return false
}

func listAvailableGlobalSkills() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	globalDir := filepath.Join(homeDir, config.GlobalConfigDir, config.GlobalSkillsDir)
	entries, err := os.ReadDir(globalDir)
	if err != nil {
		return nil
	}

	var skills []string
	for _, entry := range entries {
		if entry.IsDir() {
			skillFile := filepath.Join(globalDir, entry.Name(), "skill.md")
			if _, err := os.Stat(skillFile); err == nil {
				skills = append(skills, entry.Name())
			}
		}
	}
	return skills
}

func showSkill(projectPath string, name string) {
	if name == "" {
		fmt.Fprintf(os.Stderr, "error: skill name required\n")
		fmt.Fprintf(os.Stderr, "usage: checkpoint skill show <name>\n")
		os.Exit(1)
	}

	ctx, err := explain.LoadExplainContext(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading context: %v\n", err)
		os.Exit(1)
	}

	// Try to find in loaded skills
	for _, s := range ctx.SkillDefs {
		if s.Name == name {
			fmt.Print(s.Content)
			return
		}
	}

	// Try to load from global even if not configured
	homeDir, err := os.UserHomeDir()
	if err == nil {
		skillPath := filepath.Join(homeDir, config.GlobalConfigDir, config.GlobalSkillsDir, name, "skill.md")
		if content, err := os.ReadFile(skillPath); err == nil {
			fmt.Print(string(content))
			fmt.Println()
			fmt.Printf("hint: This skill is not configured for this project. Run 'checkpoint skill add %s' to add it.\n", name)
			return
		}
	}

	// Try local skills directory
	localPath := filepath.Join(projectPath, config.CheckpointDir, config.SkillsDir, name, "skill.md")
	if content, err := os.ReadFile(localPath); err == nil {
		fmt.Print(string(content))
		return
	}

	fmt.Fprintf(os.Stderr, "skill '%s' not found\n", name)
	fmt.Fprintf(os.Stderr, "hint: Run 'checkpoint skill list' to see available skills\n")
	os.Exit(1)
}

func addSkill(projectPath string, name string) {
	if name == "" {
		fmt.Fprintf(os.Stderr, "error: skill name required\n")
		fmt.Fprintf(os.Stderr, "usage: checkpoint skill add <name>\n")
		os.Exit(1)
	}

	// Verify skill exists in global
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot get home directory: %v\n", err)
		os.Exit(1)
	}

	skillPath := filepath.Join(homeDir, config.GlobalConfigDir, config.GlobalSkillsDir, name, "skill.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: global skill '%s' not found\n", name)
		fmt.Fprintf(os.Stderr, "hint: Available global skills are in ~/.config/checkpoint/skills/\n")
		os.Exit(1)
	}

	// Load current skills.yml
	skillsPath := filepath.Join(projectPath, config.CheckpointDir, config.ExplainSkillsYml)
	var skillsConfig explain.SkillsConfig

	if data, err := os.ReadFile(skillsPath); err == nil {
		if err := yaml.Unmarshal(data, &skillsConfig); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing skills.yml: %v\n", err)
			os.Exit(1)
		}
	} else {
		skillsConfig.SchemaVersion = "1"
	}

	// Check if already added
	for _, existing := range skillsConfig.Global {
		if existing == name {
			fmt.Printf("skill '%s' is already configured\n", name)
			return
		}
	}

	// Add skill
	skillsConfig.Global = append(skillsConfig.Global, name)

	// Write back
	data, err := yaml.Marshal(&skillsConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling skills.yml: %v\n", err)
		os.Exit(1)
	}

	content := "schema_version: \"1\"\n\n" + strings.TrimPrefix(string(data), "schema_version: \"1\"\n")
	if err := os.WriteFile(skillsPath, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing skills.yml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Added global skill '%s' to project\n", name)
}

func createSkill(projectPath string, name string) {
	if name == "" {
		fmt.Fprintf(os.Stderr, "error: skill name required\n")
		fmt.Fprintf(os.Stderr, "usage: checkpoint skill create <name>\n")
		os.Exit(1)
	}

	// Create skill directory
	skillDir := filepath.Join(projectPath, config.CheckpointDir, config.SkillsDir, name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating skill directory: %v\n", err)
		os.Exit(1)
	}

	// Create skill.md template
	skillPath := filepath.Join(skillDir, "skill.md")
	if _, err := os.Stat(skillPath); err == nil {
		fmt.Fprintf(os.Stderr, "error: skill '%s' already exists\n", name)
		os.Exit(1)
	}

	template := fmt.Sprintf(`# %s

(Describe what this skill is and when to use it)

## Purpose

(What does this skill help accomplish?)

## When to Use

- (Scenario 1)
- (Scenario 2)

## Usage

`+"```"+`
(Commands or instructions)
`+"```"+`

## Tips

- (Helpful tip 1)
- (Helpful tip 2)

## Related

- (Related skills or resources)
`, name)

	if err := os.WriteFile(skillPath, []byte(template), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing skill.md: %v\n", err)
		os.Exit(1)
	}

	// Add to skills.yml
	skillsPath := filepath.Join(projectPath, config.CheckpointDir, config.ExplainSkillsYml)
	var skillsConfig explain.SkillsConfig

	if data, err := os.ReadFile(skillsPath); err == nil {
		yaml.Unmarshal(data, &skillsConfig)
	}
	skillsConfig.SchemaVersion = "1"

	// Check if already in local
	for _, existing := range skillsConfig.Local {
		if existing == name {
			fmt.Printf("✓ Created skill at %s\n", skillPath)
			return
		}
	}

	skillsConfig.Local = append(skillsConfig.Local, name)

	data, _ := yaml.Marshal(&skillsConfig)
	content := "schema_version: \"1\"\n\n" + strings.TrimPrefix(string(data), "schema_version: \"1\"\n")
	os.WriteFile(skillsPath, []byte(content), 0644)

	fmt.Printf("✓ Created skill '%s' at %s\n", name, skillPath)
	fmt.Printf("  Edit the skill.md file to add content\n")
}

// InitGlobalSkills creates the global skills directory with default skills
func InitGlobalSkills() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get home directory: %w", err)
	}

	globalSkillsDir := filepath.Join(homeDir, config.GlobalConfigDir, config.GlobalSkillsDir)
	if err := os.MkdirAll(globalSkillsDir, 0755); err != nil {
		return fmt.Errorf("create global skills directory: %w", err)
	}

	// Create default skills
	defaultSkills := map[string]string{
		"git":     gitSkillContent,
		"ripgrep": ripgrepSkillContent,
		"go":      goSkillContent,
	}

	for name, content := range defaultSkills {
		skillDir := filepath.Join(globalSkillsDir, name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			continue
		}
		skillPath := filepath.Join(skillDir, "skill.md")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			os.WriteFile(skillPath, []byte(content), 0644)
		}
	}

	return nil
}

const gitSkillContent = `# Git

Version control system for tracking changes.

## Purpose

Track code changes, collaborate with others, and maintain project history.

## Common Commands

` + "```bash" + `
git status              # Show working tree status
git diff                # Show unstaged changes
git diff --staged       # Show staged changes
git add <file>          # Stage file for commit
git add -A              # Stage all changes
git commit -m "msg"     # Commit staged changes
git log --oneline -10   # Show recent commits
git branch              # List branches
git checkout <branch>   # Switch branches
git pull                # Fetch and merge remote
git push                # Push to remote
` + "```" + `

## Tips

- Use descriptive commit messages
- Commit often, push regularly
- Review changes before committing with git diff
- Use branches for features/experiments

## With Checkpoint

Always use ` + "`checkpoint commit`" + ` instead of raw ` + "`git commit`" + ` to maintain the changelog.
`

const ripgrepSkillContent = `# ripgrep (rg)

Fast regex search tool for codebases.

## Purpose

Search for patterns across files quickly. Much faster than grep for large codebases.

## Common Commands

` + "```bash" + `
rg "pattern"                    # Basic search
rg "pattern" --type go          # Filter by filetype
rg "pattern" -g "*.go"          # Glob filter
rg "pattern" -A 3 -B 3          # With context lines
rg "pattern" -l                 # Just filenames
rg "pattern" -c                 # Count matches
rg "pattern" -i                 # Case insensitive
rg "pattern" -w                 # Word boundaries
rg "func \w+\(" --type go       # Regex: find functions
rg "TODO|FIXME"                 # Multiple patterns
` + "```" + `

## Tips

- Use --type instead of glob for common languages (go, py, js, etc.)
- Use -w for whole word matching
- Use -F for literal strings (no regex)
- Use --hidden to include dotfiles
- Use -g '!vendor' to exclude directories

## File Types

` + "```bash" + `
rg --type-list                  # Show all known types
rg "pattern" -t go -t rust      # Multiple types
` + "```" + `
`

const goSkillContent = `# Go

Go programming language toolchain.

## Purpose

Build, test, and manage Go applications.

## Common Commands

` + "```bash" + `
go build .              # Build current package
go run .                # Build and run
go test ./...           # Run all tests
go test -v ./...        # Verbose tests
go test -cover ./...    # With coverage
go test -race ./...     # With race detector
go fmt ./...            # Format code
go vet ./...            # Static analysis
go mod init <name>      # Initialize module
go mod tidy             # Clean up go.mod
go mod download         # Download dependencies
go get <pkg>            # Add dependency
go get -u ./...         # Update all dependencies
` + "```" + `

## Project Structure

` + "```" + `
myproject/
├── main.go             # Entry point
├── go.mod              # Module definition
├── cmd/                # Command implementations
├── internal/           # Private packages
├── pkg/                # Public packages
└── *_test.go           # Test files
` + "```" + `

## Tips

- Run go fmt before committing
- Use go vet to catch common mistakes
- Use -race flag during development
- Keep go.mod tidy with go mod tidy
- Use internal/ for packages not meant for external use
`
