package guardrail

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmoose/checkpoint/internal/detect"
	"github.com/dmoose/checkpoint/internal/explain"
	"github.com/dmoose/checkpoint/pkg/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var doctorOpts struct {
	fix     bool
	verbose bool
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().BoolVar(&doctorOpts.fix, "fix", false, "Auto-fix issues where possible")
	doctorCmd.Flags().BoolVarP(&doctorOpts.verbose, "verbose", "v", false, "Show detected project info")
}

var doctorCmd = &cobra.Command{
	Use:   "doctor [path]",
	Short: "Check knowledge base health and suggest fixes",
	Long:  `Validates knowledge base configuration, detects missing files, suggests improvements.`,
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
		Doctor(absPath, DoctorOptions{Fix: doctorOpts.fix, Verbose: doctorOpts.verbose})
	},
}

// DoctorOptions holds flags for the doctor command
type DoctorOptions struct {
	Fix     bool // --fix flag to auto-fix issues
	Verbose bool // --verbose flag for more detail
}

// CheckResult represents the result of a single check
type CheckResult struct {
	Name    string
	Status  string // ok, warning, error, missing
	Message string
	Fix     string // Suggested fix command
	AutoFix bool   // Can be auto-fixed
}

// Doctor validates knowledge base health and suggests fixes
func Doctor(projectPath string, opts DoctorOptions) {
	fmt.Println("Guardrail Doctor")
	fmt.Println("================")
	fmt.Println()

	results := []CheckResult{}

	// Run all checks
	results = append(results, checkCheckpointDir(projectPath))
	results = append(results, checkProjectYml(projectPath))
	results = append(results, checkToolsYml(projectPath))
	results = append(results, checkGuidelinesYml(projectPath))
	results = append(results, checkSkills(projectPath))
	results = append(results, checkPrompts(projectPath))
	results = append(results, checkExamples(projectPath))

	// Count results
	okCount := 0
	warnCount := 0
	errCount := 0
	missingCount := 0

	for _, r := range results {
		switch r.Status {
		case "ok":
			okCount++
		case "warning":
			warnCount++
		case "error":
			errCount++
		case "missing":
			missingCount++
		}
	}

	// Display results
	for _, r := range results {
		icon := getStatusIcon(r.Status)
		fmt.Printf("%s %s: %s\n", icon, r.Name, r.Message)

		if r.Fix != "" && (r.Status == "error" || r.Status == "missing" || r.Status == "warning") {
			if opts.Fix && r.AutoFix {
				fmt.Printf("   -> Auto-fixing...\n")
				// Auto-fix would go here
			} else {
				fmt.Printf("   fix: %s\n", r.Fix)
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("Summary")
	fmt.Println("-------")
	fmt.Printf("  %d passed, %d warnings, %d errors, %d missing\n", okCount, warnCount, errCount, missingCount)

	if errCount > 0 || missingCount > 0 {
		fmt.Println()
		fmt.Println("Run suggested fix commands to resolve issues.")
		fmt.Println("For initial setup, run: guardrail init")
	} else if warnCount > 0 {
		fmt.Println()
		fmt.Println("Warnings won't prevent guardrail from working, but")
		fmt.Println("addressing them will improve your experience.")
	} else {
		fmt.Println()
		fmt.Println("All checks passed! Your knowledge base is well configured.")
	}

	// Show detected info if verbose
	if opts.Verbose {
		fmt.Println()
		fmt.Println("Detected Project Info")
		fmt.Println("---------------------")
		info := detect.DetectProject(projectPath)
		fmt.Printf("  Name:     %s\n", info.Name)
		fmt.Printf("  Language: %s\n", info.Language)
		if len(info.Languages) > 1 {
			fmt.Printf("  Also:     %s\n", strings.Join(info.Languages[1:], ", "))
		}
		if info.BuildCmd != "" {
			fmt.Printf("  Build:    %s\n", info.BuildCmd)
		}
		if info.TestCmd != "" {
			fmt.Printf("  Test:     %s\n", info.TestCmd)
		}
		if info.LintCmd != "" {
			fmt.Printf("  Lint:     %s\n", info.LintCmd)
		}
		if len(info.Frameworks) > 0 {
			fmt.Printf("  Frameworks: %s\n", strings.Join(info.Frameworks, ", "))
		}
	}
}

func getStatusIcon(status string) string {
	switch status {
	case "ok":
		return "[OK]"
	case "warning":
		return "[!!]"
	case "error":
		return "[ERR]"
	case "missing":
		return "[--]"
	default:
		return "[??]"
	}
}

func checkCheckpointDir(projectPath string) CheckResult {
	checkpointPath := filepath.Join(projectPath, config.CheckpointDir)
	if _, err := os.Stat(checkpointPath); err != nil {
		return CheckResult{
			Name:    "Checkpoint Directory",
			Status:  "missing",
			Message: ".checkpoint/ directory not found",
			Fix:     "guardrail init",
			AutoFix: false,
		}
	}
	return CheckResult{
		Name:    "Checkpoint Directory",
		Status:  "ok",
		Message: ".checkpoint/ directory exists",
	}
}

func checkProjectYml(projectPath string) CheckResult {
	projectYamlPath := filepath.Join(projectPath, config.CheckpointDir, config.ExplainProjectYaml)
	data, err := os.ReadFile(projectYamlPath)
	if err != nil {
		return CheckResult{
			Name:    "Project Config",
			Status:  "missing",
			Message: "project.yaml not found",
			Fix:     "guardrail init",
			AutoFix: false,
		}
	}

	var proj explain.ProjectConfig
	if err := yaml.Unmarshal(data, &proj); err != nil {
		return CheckResult{
			Name:    "Project Config",
			Status:  "error",
			Message: fmt.Sprintf("project.yaml parse error: %v", err),
			Fix:     "Check YAML syntax in .checkpoint/project.yaml",
		}
	}

	// Check for placeholder values
	issues := []string{}
	if proj.Name == "" || proj.Name == "project-name" {
		issues = append(issues, "name not set")
	}
	if proj.Purpose == "" || strings.Contains(proj.Purpose, "TODO") {
		issues = append(issues, "purpose not set")
	}
	if proj.Languages.Primary == "" {
		issues = append(issues, "language not set")
	}

	if len(issues) > 0 {
		return CheckResult{
			Name:    "Project Config",
			Status:  "warning",
			Message: fmt.Sprintf("project.yaml incomplete: %s", strings.Join(issues, ", ")),
			Fix:     "Edit .checkpoint/project.yaml to fill in project details",
		}
	}

	return CheckResult{
		Name:    "Project Config",
		Status:  "ok",
		Message: "project.yaml configured",
	}
}

func checkToolsYml(projectPath string) CheckResult {
	toolsYamlPath := filepath.Join(projectPath, config.CheckpointDir, config.ExplainToolsYaml)
	data, err := os.ReadFile(toolsYamlPath)
	if err != nil {
		return CheckResult{
			Name:    "Tools Config",
			Status:  "missing",
			Message: "tools.yaml not found",
			Fix:     "guardrail init",
			AutoFix: false,
		}
	}

	var tools explain.ToolsConfig
	if err := yaml.Unmarshal(data, &tools); err != nil {
		return CheckResult{
			Name:    "Tools Config",
			Status:  "error",
			Message: fmt.Sprintf("tools.yaml parse error: %v", err),
			Fix:     "Check YAML syntax in .checkpoint/tools.yaml",
		}
	}

	// Check for common commands (build and test are typical; lint is optional)
	hasTest := len(tools.Test) > 0
	hasBuild := len(tools.Build) > 0

	missing := []string{}
	if !hasTest {
		missing = append(missing, "test")
	}
	if !hasBuild {
		missing = append(missing, "build")
	}

	if len(missing) > 0 {
		// Detect what we can
		info := detect.DetectProject(projectPath)
		fixes := []string{}
		if !hasTest && info.TestCmd != "" {
			fixes = append(fixes, fmt.Sprintf("guardrail learn --tool test '%s'", info.TestCmd))
		}
		if !hasBuild && info.BuildCmd != "" {
			fixes = append(fixes, fmt.Sprintf("guardrail learn --tool build '%s'", info.BuildCmd))
		}

		fix := strings.Join(fixes, " && ")
		if fix == "" {
			fix = "Add commands to .checkpoint/tools.yaml"
		}

		return CheckResult{
			Name:    "Tools Config",
			Status:  "warning",
			Message: fmt.Sprintf("tools.yaml missing: %s", strings.Join(missing, ", ")),
			Fix:     fix,
		}
	}

	// Count total commands
	cmdCount := len(tools.Build) + len(tools.Test) + len(tools.Lint) + len(tools.Check) + len(tools.Run) + len(tools.Checkpoint) + len(tools.Maintenance)

	return CheckResult{
		Name:    "Tools Config",
		Status:  "ok",
		Message: fmt.Sprintf("tools.yaml has %d command group(s)", cmdCount),
	}
}

func checkGuidelinesYml(projectPath string) CheckResult {
	guidelinesYamlPath := filepath.Join(projectPath, config.CheckpointDir, config.ExplainGuidelinesYaml)
	data, err := os.ReadFile(guidelinesYamlPath)
	if err != nil {
		return CheckResult{
			Name:    "Guidelines Config",
			Status:  "missing",
			Message: "guidelines.yaml not found",
			Fix:     "guardrail init",
			AutoFix: false,
		}
	}

	var guidelines explain.GuidelinesConfig
	if err := yaml.Unmarshal(data, &guidelines); err != nil {
		return CheckResult{
			Name:    "Guidelines Config",
			Status:  "error",
			Message: fmt.Sprintf("guidelines.yaml parse error: %v", err),
			Fix:     "Check YAML syntax in .checkpoint/guidelines.yaml",
		}
	}

	// Check if it has any content
	hasContent := len(guidelines.Naming) > 0 ||
		len(guidelines.Structure) > 0 ||
		len(guidelines.Rules) > 0 ||
		len(guidelines.Avoid) > 0 ||
		len(guidelines.Principles) > 0

	if !hasContent {
		return CheckResult{
			Name:    "Guidelines Config",
			Status:  "warning",
			Message: "guidelines.yaml is empty",
			Fix:     "Add guidelines with: guardrail learn --guideline 'Your guideline here'",
		}
	}

	return CheckResult{
		Name:    "Guidelines Config",
		Status:  "ok",
		Message: "guidelines.yaml configured",
	}
}

func checkSkills(projectPath string) CheckResult {
	skillsYamlPath := filepath.Join(projectPath, config.CheckpointDir, config.ExplainSkillsYaml)
	if _, err := os.Stat(skillsYamlPath); err != nil {
		return CheckResult{
			Name:    "Skills",
			Status:  "warning",
			Message: "No skills configured",
			Fix:     "guardrail init",
		}
	}

	// Check skills directory
	skillsDir := filepath.Join(projectPath, config.CheckpointDir, config.SkillsDir)
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return CheckResult{
			Name:    "Skills",
			Status:  "ok",
			Message: "Skills config exists, no local skills defined",
		}
	}

	skillCount := 0
	for _, e := range entries {
		if e.IsDir() {
			skillCount++
		}
	}

	if skillCount > 0 {
		return CheckResult{
			Name:    "Skills",
			Status:  "ok",
			Message: fmt.Sprintf("%d local skills defined", skillCount),
		}
	}

	return CheckResult{
		Name:    "Skills",
		Status:  "ok",
		Message: "Skills configured",
	}
}

func checkPrompts(projectPath string) CheckResult {
	promptsDir := filepath.Join(projectPath, config.CheckpointDir, "prompts")
	if _, err := os.Stat(promptsDir); err != nil {
		return CheckResult{
			Name:    "Prompts",
			Status:  "missing",
			Message: "prompts directory not found",
			Fix:     "guardrail init",
			AutoFix: false,
		}
	}

	entries, err := os.ReadDir(promptsDir)
	if err != nil {
		return CheckResult{
			Name:    "Prompts",
			Status:  "error",
			Message: fmt.Sprintf("cannot read prompts directory: %v", err),
		}
	}

	fileCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			fileCount++
		}
	}

	if fileCount == 0 {
		return CheckResult{
			Name:    "Prompts",
			Status:  "warning",
			Message: "prompts directory is empty",
			Fix:     "guardrail init",
		}
	}

	return CheckResult{
		Name:    "Prompts",
		Status:  "ok",
		Message: fmt.Sprintf("prompts directory has %d file(s)", fileCount),
	}
}

func checkExamples(projectPath string) CheckResult {
	examplesDir := filepath.Join(projectPath, config.CheckpointDir, "examples")
	if _, err := os.Stat(examplesDir); err != nil {
		return CheckResult{
			Name:    "Examples",
			Status:  "missing",
			Message: "examples directory not found",
			Fix:     "guardrail init",
			AutoFix: false,
		}
	}

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		return CheckResult{
			Name:    "Examples",
			Status:  "error",
			Message: fmt.Sprintf("cannot read examples directory: %v", err),
		}
	}

	fileCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			fileCount++
		}
	}

	if fileCount == 0 {
		return CheckResult{
			Name:    "Examples",
			Status:  "warning",
			Message: "examples directory is empty",
			Fix:     "Add example checkpoint entries to .checkpoint/examples/",
		}
	}

	return CheckResult{
		Name:    "Examples",
		Status:  "ok",
		Message: fmt.Sprintf("examples directory has %d file(s)", fileCount),
	}
}
