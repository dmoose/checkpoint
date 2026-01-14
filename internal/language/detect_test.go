package language

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestDetectLanguages(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "language-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	testFiles := map[string]string{
		"go.mod":                    "module test",
		"main.go":                   "package main",
		"src/handler.go":            "package src",
		"package.json":              `{"name": "test"}`,
		"app.js":                    "console.log('hello')",
		"src/component.tsx":         "export const Component = () => {}",
		"requirements.txt":          "django==3.2.0",
		"app.py":                    "print('hello')",
		"Cargo.toml":                "[package]",
		"src/main.rs":               "fn main() {}",
		"README.md":                 "# Test Project",
		"node_modules/lib/index.js": "// should be ignored",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tmpDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir for %s: %v", filePath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", filePath, err)
		}
	}

	// Run detection
	languages, err := DetectLanguages(tmpDir)
	if err != nil {
		t.Fatalf("DetectLanguages failed: %v", err)
	}

	// Sort languages by name for consistent testing
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].Name < languages[j].Name
	})

	// Verify we detected the expected languages
	expectedLanguages := []string{"Go", "JavaScript/TypeScript", "Python", "Rust"}
	actualLanguages := make([]string, len(languages))
	for i, lang := range languages {
		actualLanguages[i] = lang.Name
	}

	if !reflect.DeepEqual(expectedLanguages, actualLanguages) {
		t.Errorf("expected languages %v, got %v", expectedLanguages, actualLanguages)
	}

	// Verify all languages have indicators (since we only detect reliable ones)
	for _, lang := range languages {
		if len(lang.Indicators) == 0 {
			t.Errorf("language %s has no indicators", lang.Name)
		}
	}
}

func TestDetectLanguagesWithOnlySourceFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "language-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create only source files, no project files
	testFiles := map[string]string{
		"main.go":   "package main",
		"utils.go":  "package main",
		"script.py": "print('hello')",
		"README.md": "# Test",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tmpDir, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", filePath, err)
		}
	}

	languages, err := DetectLanguages(tmpDir)
	if err != nil {
		t.Fatalf("DetectLanguages failed: %v", err)
	}

	// Should detect nothing since we only look for reliable project files
	if len(languages) != 0 {
		t.Errorf("expected no languages detected, got %d: %v", len(languages), languages)
	}
}

func TestDetectLanguagesSwiftProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "swift-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create Swift project structure
	testFiles := map[string]string{
		"Package.swift":                   "// swift-tools-version:5.5",
		"Sources/App/main.swift":          "print(\"Hello, World!\")",
		"MyApp.xcodeproj/project.pbxproj": "// Xcode project",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tmpDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir for %s: %v", filePath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", filePath, err)
		}
	}

	languages, err := DetectLanguages(tmpDir)
	if err != nil {
		t.Fatalf("DetectLanguages failed: %v", err)
	}

	// Should detect both Swift and Swift/Objective-C
	languageNames := make([]string, len(languages))
	for i, lang := range languages {
		languageNames[i] = lang.Name
	}

	expectedToFind := []string{"Swift", "Swift/Objective-C"}
	for _, expected := range expectedToFind {
		found := false
		for _, actual := range languageNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find language %s, but didn't. Found: %v", expected, languageNames)
		}
	}
}

func TestGetPrimaryLanguage(t *testing.T) {
	tests := []struct {
		name      string
		languages []Language
		expected  string
	}{
		{
			name: "single language",
			languages: []Language{
				{Name: "Go", Indicators: []string{"go.mod"}},
			},
			expected: "Go",
		},
		{
			name: "multiple languages, pick one with most indicators",
			languages: []Language{
				{Name: "Go", Indicators: []string{"go.mod"}},
				{Name: "JavaScript/TypeScript", Indicators: []string{"package.json", "yarn.lock", "package-lock.json"}},
			},
			expected: "JavaScript/TypeScript",
		},
		{
			name: "equal indicators, pick first",
			languages: []Language{
				{Name: "Python", Indicators: []string{"requirements.txt"}},
				{Name: "JavaScript/TypeScript", Indicators: []string{"package.json"}},
			},
			expected: "Python",
		},
		{
			name:      "empty languages",
			languages: []Language{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPrimaryLanguage(tt.languages)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDetermineConfidence(t *testing.T) {

}

func TestDedupeStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", "d"}
	expected := []string{"a", "b", "c", "d"}
	result := dedupeStrings(input)

	if len(result) != len(expected) {
		t.Errorf("expected length %d, got %d", len(expected), len(result))
		return
	}

	// Convert to maps for comparison since order might differ
	expectedMap := make(map[string]bool)
	resultMap := make(map[string]bool)

	for _, s := range expected {
		expectedMap[s] = true
	}
	for _, s := range result {
		resultMap[s] = true
	}

	if !reflect.DeepEqual(expectedMap, resultMap) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestIgnorePatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ignore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create files that should be ignored
	ignoredFiles := map[string]string{
		"node_modules/lib/index.js": "// should be ignored",
		"vendor/lib/main.go":        "// should be ignored",
		"target/release/main.rs":    "// should be ignored",
		".git/hooks/pre-commit":     "// should be ignored",
		"build/output/app.js":       "// should be ignored",
		"dist/bundle.js":            "// should be ignored",
	}

	// Create files that should NOT be ignored
	validFiles := map[string]string{
		"go.mod":       "module test",
		"package.json": `{"name": "test"}`,
	}

	allFiles := make(map[string]string)
	for k, v := range ignoredFiles {
		allFiles[k] = v
	}
	for k, v := range validFiles {
		allFiles[k] = v
	}

	for filePath, content := range allFiles {
		fullPath := filepath.Join(tmpDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir for %s: %v", filePath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", filePath, err)
		}
	}

	languages, err := DetectLanguages(tmpDir)
	if err != nil {
		t.Fatalf("DetectLanguages failed: %v", err)
	}

	// Should only detect Go and JavaScript from the valid files
	languageNames := make([]string, len(languages))
	for i, lang := range languages {
		languageNames[i] = lang.Name
	}

	// Should contain Go and JavaScript/TypeScript but not from ignored paths
	foundGo := false
	foundJS := false
	for _, lang := range languageNames {
		if lang == "Go" {
			foundGo = true
		}
		if lang == "JavaScript/TypeScript" {
			foundJS = true
		}
	}

	if !foundGo {
		t.Error("expected to find Go language from valid files")
	}
	if !foundJS {
		t.Error("expected to find JavaScript/TypeScript from valid files")
	}

	// Verify that the indicators don't mention ignored paths
	for _, lang := range languages {
		for _, indicator := range lang.Indicators {
			if strings.Contains(indicator, "node_modules") ||
				strings.Contains(indicator, "vendor") ||
				strings.Contains(indicator, "target") ||
				strings.Contains(indicator, "build") ||
				strings.Contains(indicator, "dist") {
				t.Errorf("language %s has indicator from ignored path: %s", lang.Name, indicator)
			}
		}
	}
}
