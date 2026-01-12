package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGitRemote(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	tests := []struct {
		name     string
		config   string
		expected string
	}{
		{
			name: "https remote",
			config: `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = https://github.com/user/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*`,
			expected: "https://github.com/user/repo.git",
		},
		{
			name: "ssh remote",
			config: `[remote "origin"]
	url = git@github.com:user/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*`,
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "no origin",
			config:   `[core]\n\trepositoryformatversion = 0`,
			expected: "",
		},
		{
			name: "multiple remotes",
			config: `[remote "upstream"]
	url = https://github.com/upstream/repo.git
[remote "origin"]
	url = https://github.com/fork/repo.git`,
			expected: "https://github.com/fork/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(gitDir, "config")
			if err := os.WriteFile(configPath, []byte(tt.config), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			result := detectGitRemote(tmpDir)
			if result != tt.expected {
				t.Errorf("detectGitRemote() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHasPythonProject(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected bool
	}{
		{
			name:     "pyproject.toml",
			files:    []string{"pyproject.toml"},
			expected: true,
		},
		{
			name:     "requirements.txt",
			files:    []string{"requirements.txt"},
			expected: true,
		},
		{
			name:     "setup.py",
			files:    []string{"setup.py"},
			expected: true,
		},
		{
			name:     "Pipfile",
			files:    []string{"Pipfile"},
			expected: true,
		},
		{
			name:     "no python files",
			files:    []string{"main.go", "package.json"},
			expected: false,
		},
		{
			name:     "empty directory",
			files:    []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for _, f := range tt.files {
				if err := os.WriteFile(filepath.Join(tmpDir, f), []byte(""), 0644); err != nil {
					t.Fatalf("failed to create file %s: %v", f, err)
				}
			}

			result := hasPythonProject(tmpDir)
			if result != tt.expected {
				t.Errorf("hasPythonProject() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name        string
		lockFile    string
		expectedPM  string
		expectedRun string
	}{
		{
			name:        "npm (default)",
			lockFile:    "",
			expectedPM:  "npm",
			expectedRun: "npm run",
		},
		{
			name:        "yarn",
			lockFile:    "yarn.lock",
			expectedPM:  "yarn",
			expectedRun: "yarn",
		},
		{
			name:        "pnpm",
			lockFile:    "pnpm-lock.yaml",
			expectedPM:  "pnpm",
			expectedRun: "pnpm run",
		},
		{
			name:        "bun",
			lockFile:    "bun.lockb",
			expectedPM:  "bun",
			expectedRun: "bun",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create package.json with a test script
			pkgJSON := `{"name": "test", "scripts": {"test": "jest"}}`
			if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
				t.Fatalf("failed to create package.json: %v", err)
			}

			// Create lock file if specified
			if tt.lockFile != "" {
				if err := os.WriteFile(filepath.Join(tmpDir, tt.lockFile), []byte(""), 0644); err != nil {
					t.Fatalf("failed to create lock file: %v", err)
				}
			}

			info := &ProjectInfo{}
			detectPackageJsonCommands(tmpDir, info)

			// The test command should use the correct package manager
			expectedTestCmd := tt.expectedRun + " test"
			if info.TestCmd != expectedTestCmd {
				t.Errorf("TestCmd = %q, want %q", info.TestCmd, expectedTestCmd)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"go", "javascript"}, "go", true},
	}

	for _, tt := range tests {
		result := contains(tt.slice, tt.item)
		if result != tt.expected {
			t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, result, tt.expected)
		}
	}
}

func TestDetectProjectGo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goMod := `module github.com/user/myproject

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	info := DetectProject(tmpDir)

	if info.Language != "go" {
		t.Errorf("Language = %q, want 'go'", info.Language)
	}
	if info.Name != "myproject" {
		t.Errorf("Name = %q, want 'myproject'", info.Name)
	}
	if info.BuildCmd == "" {
		t.Error("BuildCmd should not be empty for Go project")
	}
	if info.TestCmd == "" {
		t.Error("TestCmd should not be empty for Go project")
	}
}

func TestDetectProjectNode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package.json
	pkgJSON := `{
  "name": "my-node-app",
  "description": "A test app",
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "lint": "eslint ."
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	info := DetectProject(tmpDir)

	if info.Language != "javascript" {
		t.Errorf("Language = %q, want 'javascript'", info.Language)
	}
	if info.Name != "my-node-app" {
		t.Errorf("Name = %q, want 'my-node-app'", info.Name)
	}
	if info.Description != "A test app" {
		t.Errorf("Description = %q, want 'A test app'", info.Description)
	}

	// Should detect express framework
	foundExpress := false
	for _, fw := range info.Frameworks {
		if fw == "express" {
			foundExpress = true
			break
		}
	}
	if !foundExpress {
		t.Errorf("Frameworks = %v, should contain 'express'", info.Frameworks)
	}
}
