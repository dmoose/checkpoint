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
