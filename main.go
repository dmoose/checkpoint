package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-llm/cmd"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		cmd.Help()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]
	projectPath := "."
	// crude flag parsing: collect known flags, last non-flag is path
	dryRun := false
	changelogOnly := false
	var positional []string
	for _, a := range args {
		if a == "-n" || a == "--dry-run" {
			dryRun = true
			continue
		}
		if a == "--changelog-only" {
			changelogOnly = true
			continue
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		positional = append(positional, a)
	}
	if len(positional) > 0 {
		projectPath = positional[len(positional)-1]
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
		os.Exit(1)
	}

	switch subcommand {
	case "start":
		cmd.Start(absPath)
	case "check":
		cmd.Check(absPath)
	case "commit":
		cmd.CommitWithOptions(absPath, cmd.CommitOptions{DryRun: dryRun, ChangelogOnly: changelogOnly}, version)
	case "init":
		cmd.Init(absPath, version)
	case "clean":
		cmd.Clean(absPath)
	case "lint":
		cmd.Lint(absPath)
	case "examples":
		// Examples takes optional category as first positional arg
		// The path handling should use "." as default, not interpret category as path
		category := ""
		examplesPath := "."

		if len(positional) > 0 {
			// First arg could be category or path
			// Check if it looks like a path (has / or . or is a directory)
			firstArg := positional[0]
			if strings.Contains(firstArg, "/") || strings.Contains(firstArg, ".") ||
				(len(positional) > 1) {
				// Looks like we have both category and path
				if len(positional) == 1 {
					// Just a path
					examplesPath = firstArg
				} else {
					// category and path
					category = firstArg
					examplesPath = positional[1]
				}
			} else {
				// Single word with no path separators - treat as category
				category = firstArg
			}
		}

		// Resolve the path
		examplesAbsPath, err := filepath.Abs(examplesPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.Examples(examplesAbsPath, category)
	case "guide":
		// Guide takes optional topic as first positional arg
		// Same logic as examples command
		topic := ""
		guidePath := "."

		if len(positional) > 0 {
			firstArg := positional[0]
			if strings.Contains(firstArg, "/") || strings.Contains(firstArg, ".") ||
				(len(positional) > 1) {
				if len(positional) == 1 {
					guidePath = firstArg
				} else {
					topic = firstArg
					guidePath = positional[1]
				}
			} else {
				topic = firstArg
			}
		}

		guideAbsPath, err := filepath.Abs(guidePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
			os.Exit(1)
		}
		cmd.Guide(guideAbsPath, topic)
	case "help", "-h", "--help":
		cmd.Help()
	case "version", "-v", "--version":
		fmt.Printf("checkpoint version %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", subcommand)
		cmd.Help()
		os.Exit(1)
	}
}
