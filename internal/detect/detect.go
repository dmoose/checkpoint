package detect

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectInfo holds auto-detected project information
type ProjectInfo struct {
	Name        string
	Language    string
	Languages   []string // Additional languages detected
	BuildCmd    string
	TestCmd     string
	LintCmd     string
	FormatCmd   string
	DevCmd      string
	CleanCmd    string
	Description string
	Frameworks  []string
	HasGit      bool
	GitRemote   string
}

// DetectProject analyzes a project directory and returns detected info
func DetectProject(projectPath string) *ProjectInfo {
	info := &ProjectInfo{}

	// Detect project name from directory
	info.Name = filepath.Base(projectPath)

	// Check for git
	if _, err := os.Stat(filepath.Join(projectPath, ".git")); err == nil {
		info.HasGit = true
		info.GitRemote = detectGitRemote(projectPath)
	}

	// Detect language and package info
	detectLanguage(projectPath, info)

	// Detect commands from Makefile
	detectMakefileCommands(projectPath, info)

	// Detect commands from package.json
	detectPackageJsonCommands(projectPath, info)

	// Detect commands from common patterns
	detectCommonCommands(projectPath, info)

	return info
}

func detectGitRemote(projectPath string) string {
	configPath := filepath.Join(projectPath, ".git", "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	// Simple regex to find origin URL
	re := regexp.MustCompile(`\[remote "origin"\][^\[]*url\s*=\s*(\S+)`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func detectLanguage(projectPath string, info *ProjectInfo) {
	// Go
	if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); err == nil {
		info.Language = "go"
		info.Languages = append(info.Languages, "go")
		detectGoInfo(projectPath, info)
	}

	// Node.js
	if _, err := os.Stat(filepath.Join(projectPath, "package.json")); err == nil {
		if info.Language == "" {
			info.Language = "javascript"
		}
		info.Languages = append(info.Languages, "javascript")
		detectNodeInfo(projectPath, info)
	}

	// TypeScript
	if _, err := os.Stat(filepath.Join(projectPath, "tsconfig.json")); err == nil {
		if info.Language == "" {
			info.Language = "typescript"
		}
		info.Languages = append(info.Languages, "typescript")
	}

	// Python
	if hasPythonProject(projectPath) {
		if info.Language == "" {
			info.Language = "python"
		}
		info.Languages = append(info.Languages, "python")
		detectPythonInfo(projectPath, info)
	}

	// Rust
	if _, err := os.Stat(filepath.Join(projectPath, "Cargo.toml")); err == nil {
		if info.Language == "" {
			info.Language = "rust"
		}
		info.Languages = append(info.Languages, "rust")
		detectRustInfo(projectPath, info)
	}

	// Java/Kotlin (Maven)
	if _, err := os.Stat(filepath.Join(projectPath, "pom.xml")); err == nil {
		if info.Language == "" {
			info.Language = "java"
		}
		info.Languages = append(info.Languages, "java")
		info.BuildCmd = "mvn package"
		info.TestCmd = "mvn test"
	}

	// Java/Kotlin (Gradle)
	if _, err := os.Stat(filepath.Join(projectPath, "build.gradle")); err == nil {
		if info.Language == "" {
			info.Language = "java"
		}
		if !contains(info.Languages, "java") {
			info.Languages = append(info.Languages, "java")
		}
		info.BuildCmd = "./gradlew build"
		info.TestCmd = "./gradlew test"
	}
}

func hasPythonProject(projectPath string) bool {
	pythonFiles := []string{"pyproject.toml", "setup.py", "setup.cfg", "requirements.txt", "Pipfile"}
	for _, f := range pythonFiles {
		if _, err := os.Stat(filepath.Join(projectPath, f)); err == nil {
			return true
		}
	}
	return false
}

func detectGoInfo(projectPath string, info *ProjectInfo) {
	// Read go.mod for module name
	data, err := os.ReadFile(filepath.Join(projectPath, "go.mod"))
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			moduleName := strings.TrimPrefix(line, "module ")
			moduleName = strings.TrimSpace(moduleName)
			// Use last part of module path as name if it looks better
			parts := strings.Split(moduleName, "/")
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				if lastPart != "" && lastPart != info.Name {
					info.Name = lastPart
				}
			}
			break
		}
	}

	// Default Go commands if not already set
	if info.BuildCmd == "" {
		info.BuildCmd = "go build ./..."
	}
	if info.TestCmd == "" {
		info.TestCmd = "go test ./..."
	}

	// Check for golangci-lint
	if _, err := os.Stat(filepath.Join(projectPath, ".golangci.yml")); err == nil {
		info.LintCmd = "golangci-lint run"
	} else if _, err := os.Stat(filepath.Join(projectPath, ".golangci.yaml")); err == nil {
		info.LintCmd = "golangci-lint run"
	}

	// Check for gofmt/goimports
	info.FormatCmd = "go fmt ./..."
}

func detectNodeInfo(projectPath string, info *ProjectInfo) {
	data, err := os.ReadFile(filepath.Join(projectPath, "package.json"))
	if err != nil {
		return
	}

	var pkg struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Scripts     map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	if pkg.Name != "" {
		info.Name = pkg.Name
	}
	if pkg.Description != "" {
		info.Description = pkg.Description
	}

	// Detect frameworks from dependencies
	detectNodeFrameworks(projectPath, info)
}

func detectNodeFrameworks(projectPath string, info *ProjectInfo) {
	data, err := os.ReadFile(filepath.Join(projectPath, "package.json"))
	if err != nil {
		return
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	allDeps := make(map[string]bool)
	for k := range pkg.Dependencies {
		allDeps[k] = true
	}
	for k := range pkg.DevDependencies {
		allDeps[k] = true
	}

	frameworks := []string{}
	if allDeps["react"] {
		frameworks = append(frameworks, "react")
	}
	if allDeps["vue"] {
		frameworks = append(frameworks, "vue")
	}
	if allDeps["next"] {
		frameworks = append(frameworks, "next.js")
	}
	if allDeps["express"] {
		frameworks = append(frameworks, "express")
	}
	if allDeps["fastify"] {
		frameworks = append(frameworks, "fastify")
	}
	if allDeps["nest"] || allDeps["@nestjs/core"] {
		frameworks = append(frameworks, "nestjs")
	}

	info.Frameworks = frameworks
}

func detectPythonInfo(projectPath string, info *ProjectInfo) {
	// Check for pyproject.toml (modern Python)
	if data, err := os.ReadFile(filepath.Join(projectPath, "pyproject.toml")); err == nil {
		// Simple extraction of project name
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "name") {
				re := regexp.MustCompile(`name\s*=\s*"([^"]+)"`)
				if matches := re.FindStringSubmatch(line); len(matches) > 1 {
					info.Name = matches[1]
					break
				}
			}
		}
	}

	// Default Python commands
	if info.TestCmd == "" {
		// Check for pytest
		if _, err := os.Stat(filepath.Join(projectPath, "pytest.ini")); err == nil {
			info.TestCmd = "pytest"
		} else if _, err := os.Stat(filepath.Join(projectPath, "pyproject.toml")); err == nil {
			info.TestCmd = "pytest"
		} else {
			info.TestCmd = "python -m pytest"
		}
	}

	// Check for common linters
	if _, err := os.Stat(filepath.Join(projectPath, ".flake8")); err == nil {
		info.LintCmd = "flake8"
	} else if _, err := os.Stat(filepath.Join(projectPath, "pyproject.toml")); err == nil {
		// Could be ruff or other modern tools
		info.LintCmd = "ruff check ."
	}

	if _, err := os.Stat(filepath.Join(projectPath, "pyproject.toml")); err == nil {
		info.FormatCmd = "ruff format ."
	} else {
		info.FormatCmd = "black ."
	}
}

func detectRustInfo(projectPath string, info *ProjectInfo) {
	// Read Cargo.toml for package name
	data, err := os.ReadFile(filepath.Join(projectPath, "Cargo.toml"))
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	inPackage := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[package]" {
			inPackage = true
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			inPackage = false
			continue
		}
		if inPackage && strings.HasPrefix(trimmed, "name") {
			re := regexp.MustCompile(`name\s*=\s*"([^"]+)"`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				info.Name = matches[1]
				break
			}
		}
	}

	// Default Rust commands
	info.BuildCmd = "cargo build"
	info.TestCmd = "cargo test"
	info.LintCmd = "cargo clippy"
	info.FormatCmd = "cargo fmt"
}

func detectMakefileCommands(projectPath string, info *ProjectInfo) {
	makefilePath := filepath.Join(projectPath, "Makefile")
	if _, err := os.Stat(makefilePath); err != nil {
		return
	}

	file, err := os.Open(makefilePath)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	targets := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Match target definitions (target: or target:deps)
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			target := strings.TrimSpace(parts[0])
			if target != "" && !strings.HasPrefix(target, ".") {
				targets[target] = true
			}
		}
	}

	// Map common target names to commands
	if targets["build"] && info.BuildCmd == "" {
		info.BuildCmd = "make build"
	}
	if targets["test"] && info.TestCmd == "" {
		info.TestCmd = "make test"
	}
	if targets["lint"] && info.LintCmd == "" {
		info.LintCmd = "make lint"
	}
	if targets["check"] && info.LintCmd == "" {
		info.LintCmd = "make check"
	}
	if targets["fmt"] && info.FormatCmd == "" {
		info.FormatCmd = "make fmt"
	}
	if targets["format"] && info.FormatCmd == "" {
		info.FormatCmd = "make format"
	}
	if targets["dev"] && info.DevCmd == "" {
		info.DevCmd = "make dev"
	}
	if targets["run"] && info.DevCmd == "" {
		info.DevCmd = "make run"
	}
	if targets["clean"] {
		info.CleanCmd = "make clean"
	}
}

func detectPackageJsonCommands(projectPath string, info *ProjectInfo) {
	data, err := os.ReadFile(filepath.Join(projectPath, "package.json"))
	if err != nil {
		return
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	// Detect package manager
	pm := "npm"
	if _, err := os.Stat(filepath.Join(projectPath, "pnpm-lock.yaml")); err == nil {
		pm = "pnpm"
	} else if _, err := os.Stat(filepath.Join(projectPath, "yarn.lock")); err == nil {
		pm = "yarn"
	} else if _, err := os.Stat(filepath.Join(projectPath, "bun.lockb")); err == nil {
		pm = "bun"
	}

	run := pm + " run"
	if pm == "yarn" || pm == "bun" {
		run = pm // yarn and bun don't need "run"
	}

	// Map script names to commands
	if _, ok := pkg.Scripts["build"]; ok && info.BuildCmd == "" {
		info.BuildCmd = run + " build"
	}
	if _, ok := pkg.Scripts["test"]; ok && info.TestCmd == "" {
		info.TestCmd = run + " test"
	}
	if _, ok := pkg.Scripts["lint"]; ok && info.LintCmd == "" {
		info.LintCmd = run + " lint"
	}
	if _, ok := pkg.Scripts["format"]; ok && info.FormatCmd == "" {
		info.FormatCmd = run + " format"
	}
	if _, ok := pkg.Scripts["fmt"]; ok && info.FormatCmd == "" {
		info.FormatCmd = run + " fmt"
	}
	if _, ok := pkg.Scripts["dev"]; ok && info.DevCmd == "" {
		info.DevCmd = run + " dev"
	}
	if _, ok := pkg.Scripts["start"]; ok && info.DevCmd == "" {
		info.DevCmd = run + " start"
	}
}

func detectCommonCommands(projectPath string, info *ProjectInfo) {
	// Fill in any remaining gaps with language-specific defaults
	switch info.Language {
	case "go":
		if info.BuildCmd == "" {
			info.BuildCmd = "go build ./..."
		}
		if info.TestCmd == "" {
			info.TestCmd = "go test ./..."
		}
	case "javascript", "typescript":
		if info.TestCmd == "" {
			// Check for test frameworks
			if _, err := os.Stat(filepath.Join(projectPath, "jest.config.js")); err == nil {
				info.TestCmd = "npx jest"
			} else if _, err := os.Stat(filepath.Join(projectPath, "vitest.config.js")); err == nil {
				info.TestCmd = "npx vitest"
			}
		}
		if info.LintCmd == "" {
			if _, err := os.Stat(filepath.Join(projectPath, ".eslintrc.js")); err == nil {
				info.LintCmd = "npx eslint ."
			} else if _, err := os.Stat(filepath.Join(projectPath, ".eslintrc.json")); err == nil {
				info.LintCmd = "npx eslint ."
			} else if _, err := os.Stat(filepath.Join(projectPath, "eslint.config.js")); err == nil {
				info.LintCmd = "npx eslint ."
			}
		}
	case "python":
		if info.TestCmd == "" {
			info.TestCmd = "pytest"
		}
	case "rust":
		if info.BuildCmd == "" {
			info.BuildCmd = "cargo build"
		}
		if info.TestCmd == "" {
			info.TestCmd = "cargo test"
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
