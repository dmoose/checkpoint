package cmd

import "fmt"

func Help() {
	fmt.Print(`checkpoint - LLM-assisted development checkpoint tracking

USAGE:
  checkpoint <command> [path]

COMMANDS:
  check       Generate input file for LLM (creates .checkpoint-input and .checkpoint-diff)
commit      Parse input file, append to .checkpoint-changelog.yaml, stage all changes, and create git commit
             [flags] [-n|--dry-run] [--changelog-only]
  init        Create CHECKPOINT.md with usage instructions in the project root
  clean       Remove .checkpoint-input and .checkpoint-diff to abort and re-run
  help        Display this help message
  version     Display version information

FILES:
  .checkpoint-input              Editable input for LLM/user
  .checkpoint-diff               Git diff context
  .checkpoint-changelog.yaml     Append-only changelog (YAML, multi-change documents)
  .checkpoint-status.yaml        Last commit metadata (not committed)

OPTIONS:
  [path]      Path to git repository (optional; defaults to current directory)
`)
}