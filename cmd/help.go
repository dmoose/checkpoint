package cmd

import "fmt"

func Help() {
	fmt.Print(`checkpoint - LLM-assisted development checkpoint tracking

USAGE:
  checkpoint <command> [flags] [path]

COMMANDS:
  check       Generate input file for LLM
              Creates .checkpoint-input and .checkpoint-diff files
              Guards against concurrent checkpoints with lock files

  commit      Parse input, append to changelog, stage changes, and git commit
              Validates input, creates YAML document, stages files, commits
              Then backfills commit hash into the last changelog document
              Flags:
                -n, --dry-run       Show commit message and staged files without committing
                --changelog-only    Stage only changelog instead of all changes

  init        Create CHECKPOINT.md with usage instructions in project root
              Provides workflow guidance and schema documentation

  clean       Remove temporary checkpoint files to abort and restart
              Deletes .checkpoint-input and .checkpoint-diff files
              Use when you need to start over or resolve conflicts

  help        Display this help message
  version     Display version information

FILES:
  .checkpoint-input              Editable input file for LLM/user (temporary)
  .checkpoint-diff               Git diff context for reference (temporary)
  .checkpoint-changelog.yaml     Append-only YAML changelog (tracked in git)
  .checkpoint-status.yaml        Last commit metadata for discovery (not committed)
  .checkpoint-lock               Lock file to prevent concurrent operations (temporary)

ARGUMENTS:
  [path]      Path to git repository (optional; defaults to current directory)

EXAMPLES:
  checkpoint check                    # Generate input files in current directory
  checkpoint commit --dry-run         # Preview what would be committed
  checkpoint commit --changelog-only  # Only stage the changelog file
  checkpoint init ~/my-project        # Initialize checkpoint in specific directory
  checkpoint clean                    # Abort current checkpoint and clean up

For detailed workflow guidance, run: checkpoint init
`)
}
