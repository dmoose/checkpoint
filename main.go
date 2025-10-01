package main

import (
	"fmt"
	"os"
	"path/filepath"

	"go-llm/cmd"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		cmd.Help()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	projectPath := "."
	if len(os.Args) >= 3 {
		projectPath = os.Args[2]
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
		os.Exit(1)
	}

switch subcommand {
	case "check":
		cmd.Check(absPath)
	case "commit":
		cmd.Commit(absPath)
	case "init":
		cmd.Init(absPath)
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
