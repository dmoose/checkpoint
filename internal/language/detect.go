package language

import (
	"os"
	"path/filepath"
	"strings"
)

// Language represents a detected programming language
type Language struct {
	Name       string   `yaml:"name"`
	Indicators []string `yaml:"indicators,omitempty"`
}

// DetectLanguages analyzes a project directory and returns detected languages
func DetectLanguages(projectPath string) ([]Language, error) {
	var languages []Language
	indicators := make(map[string][]string)

	// Walk the project directory looking for language indicators
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking on errors
		}

		// Skip hidden directories and common ignore patterns
		if info.IsDir() {
			name := filepath.Base(path)
			if strings.HasPrefix(name, ".") && name != "." {
				if name == ".git" || name == ".svn" || name == ".hg" {
					return filepath.SkipDir
				}
			}
			if name == "node_modules" || name == "vendor" || name == "target" || name == "build" || name == "dist" {
				return filepath.SkipDir
			}

			// Check for Xcode project directories
			if strings.HasSuffix(name, ".xcodeproj") {
				indicators["Swift/Objective-C"] = append(indicators["Swift/Objective-C"], name)
			}
			if strings.HasSuffix(name, ".xcworkspace") {
				indicators["Swift/Objective-C"] = append(indicators["Swift/Objective-C"], name)
			}

			return nil
		}

		filename := filepath.Base(path)

		// Check for specific project files (most reliable indicators only)
		switch filename {
		case "go.mod", "go.sum":
			indicators["Go"] = append(indicators["Go"], filename)
		case "package.json", "package-lock.json", "yarn.lock":
			indicators["JavaScript/TypeScript"] = append(indicators["JavaScript/TypeScript"], filename)
		case "Cargo.toml", "Cargo.lock":
			indicators["Rust"] = append(indicators["Rust"], filename)
		case "requirements.txt", "setup.py", "pyproject.toml", "Pipfile":
			indicators["Python"] = append(indicators["Python"], filename)
		case "Gemfile", "Gemfile.lock":
			indicators["Ruby"] = append(indicators["Ruby"], filename)
		case "composer.json", "composer.lock":
			indicators["PHP"] = append(indicators["PHP"], filename)
		case "pom.xml", "build.gradle", "gradle.properties":
			indicators["Java"] = append(indicators["Java"], filename)
		case "Package.swift":
			indicators["Swift"] = append(indicators["Swift"], filename)
		case "mix.exs":
			indicators["Elixir"] = append(indicators["Elixir"], filename)
		case "pubspec.yaml":
			indicators["Dart"] = append(indicators["Dart"], filename)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert indicators to Language structs
	for lang, inds := range indicators {
		languages = append(languages, Language{
			Name:       lang,
			Indicators: dedupeStrings(inds),
		})
	}

	return languages, nil
}

// dedupeStrings removes duplicates from a string slice
func dedupeStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, str := range strs {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}

// GetPrimaryLanguage returns the language with most indicators, or empty string if none
func GetPrimaryLanguage(languages []Language) string {
	var primary Language

	for _, lang := range languages {
		if primary.Name == "" || len(lang.Indicators) > len(primary.Indicators) {
			primary = lang
		}
	}

	return primary.Name
}
