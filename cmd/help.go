package cmd

import "fmt"

func Help() {
	fmt.Print(`checkpoint - LLM-assisted development checkpoint tracking

USAGE:
  checkpoint <command> [flags] [path]

COMMANDS:
  start       Validate readiness and show next steps
              Checks git status, checkpoint initialization, and displays
              planned work from last checkpoint

  summary     Show project overview and recent activity
              Display checkpoints, status, next steps, and patterns
              Flags: --json (machine-readable output)

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

  lint        Check checkpoint input for obvious mistakes and issues
              Validates input file and suggests improvements before commit
              Catches placeholder text, vague summaries, and common errors

  examples    Show example checkpoint entries and best practices
              Display examples of well-structured checkpoints
              Available categories: feature, bugfix, refactor, context, anti-patterns

  guide       Show detailed guides and documentation
              Display comprehensive guides for checkpoint usage
              Available topics: first-time-user, llm-workflow, best-practices

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
  checkpoint start                    # Check readiness and see what's next
  checkpoint summary                  # Show project overview
  checkpoint summary --json           # Get summary as JSON
  checkpoint check                    # Generate input files in current directory
  checkpoint lint                     # Check input for issues before committing
  checkpoint examples                 # List available examples
  checkpoint examples feature         # Show feature example
  checkpoint examples anti-patterns   # Show common mistakes to avoid
  checkpoint guide                    # List available guides
  checkpoint guide first-time-user    # Show first-time user guide
  checkpoint guide llm-workflow       # Show LLM workflow guide
  checkpoint commit --dry-run         # Preview what would be committed
  checkpoint commit --changelog-only  # Only stage the changelog file
  checkpoint init ~/my-project        # Initialize checkpoint in specific directory
  checkpoint clean                    # Abort current checkpoint and clean up

For detailed workflow guidance, run: checkpoint init
`)
}
