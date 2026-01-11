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

  init        Initialize checkpoint in a project with optional template
              Creates .checkpoint/ directory with config files
              Flags:
                --template <name>   Use a project template (go-cli, node-api, etc.)
                --list-templates    Show available templates

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

  prompt      Show LLM prompts from project library
              Display prompts with variable substitution
              Usage: checkpoint prompt [id] [--var key=value]
              Available prompts defined in .checkpoint/prompts/prompts.yaml

  explain     Get project context for LLMs and developers
              Display project info, tools, guidelines, and skills
              Usage: checkpoint explain [topic] [flags]
              Topics: project, tools, guidelines, skills, skill <name>, history, next
              Flags:
                --full   Complete context dump
                --md     Output as markdown
                --json   Output as JSON

  skill       Manage skills for LLM context
              Skills describe tools and capabilities available
              Actions:
                list              List available skills
                show <name>       Show skill details
                add <name>        Add global skill to project
                create <name>     Create new local skill

  search      Search checkpoint history
              Query changelog and context for patterns, decisions, failed approaches
              Usage: checkpoint search <query> [flags]
              Flags:
                --failed          Search failed approaches
                --pattern         Search established patterns
                --decision        Search decisions made
                --scope <scope>   Filter by scope/component
                --recent <n>      Limit to recent N checkpoints

  learn       Capture knowledge during development
              Low-friction way to add guidelines, tools, or insights
              Usage: checkpoint learn <content> [flags]
              Flags:
                --guideline       Add as a rule to follow
                --avoid           Add as an anti-pattern to avoid
                --principle       Add as a design principle
                --pattern         Add as an established pattern
                --tool [name]     Add as a tool command

  session     Manage session state for LLM handoff
              Capture and restore session context between LLM sessions
              Usage: checkpoint session [action] [summary] [flags]
              Actions:
                show              Show current session state (default)
                save <summary>    Save session state with summary
                clear             Clear session state
                handoff           Generate comprehensive handoff document
              Flags:
                --status <s>      Set status: in_progress, blocked, complete, handoff

  help        Display this help message
  version     Display version information

FILES:
  .checkpoint-input              Editable input file for LLM/user (temporary)
  .checkpoint-diff               Git diff context for reference (temporary)
  .checkpoint-changelog.yaml     Append-only YAML changelog (tracked in git)
  .checkpoint-status.yaml        Last commit metadata with project identity for discovery (not committed)
  .checkpoint-lock               Lock file to prevent concurrent operations (temporary)

ARGUMENTS:
  [path]      Path to git repository (optional; defaults to current directory)

EXAMPLES:
  checkpoint start                    # Check readiness and see what's next
  checkpoint summary                  # Show project overview
  checkpoint summary --json           # Get summary as JSON
  checkpoint explain                  # Get executive summary + options
  checkpoint explain project          # Detailed project architecture
  checkpoint explain tools            # All build/test/lint commands
  checkpoint explain guidelines       # Conventions and rules to follow
  checkpoint explain skills           # Available skills listing
  checkpoint explain skill ripgrep    # Specific skill details
  checkpoint explain history           # Show recent checkpoints, patterns, decisions
  checkpoint explain next              # Show all outstanding next steps by priority
  checkpoint explain --full            # Complete context dump
  checkpoint explain --json            # Machine-readable output
  checkpoint check                    # Generate input files in current directory
  checkpoint lint                     # Check input for issues before committing
  checkpoint examples                 # List available examples
  checkpoint examples feature         # Show feature example
  checkpoint examples anti-patterns   # Show common mistakes to avoid
  checkpoint guide                    # List available guides
  checkpoint guide first-time-user    # Show first-time user guide
  checkpoint guide llm-workflow       # Show LLM workflow guide
  checkpoint prompt                   # List available prompts
  checkpoint prompt fill-checkpoint   # Show checkpoint fill prompt
  checkpoint prompt implement-feature --var feature_name="Auth" # With variables
  checkpoint commit --dry-run         # Preview what would be committed
  checkpoint commit --changelog-only  # Only stage the changelog file
  checkpoint init                     # Initialize in current directory
  checkpoint init --list-templates    # Show available templates
  checkpoint init --template go-cli   # Initialize with Go CLI template
  checkpoint init --template node-api ~/my-project  # Use template in specific dir
  checkpoint clean                    # Abort current checkpoint and clean up
  checkpoint session                  # Show current session state
  checkpoint session save "Working on auth" # Save session with summary
  checkpoint session save "Blocked on API" --status blocked # Save with status
  checkpoint session handoff          # Generate handoff document for next LLM
  checkpoint session clear            # Clear session state

For detailed workflow guidance, run: checkpoint init
For project context, run: checkpoint explain
`)
}
