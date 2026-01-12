package cmd

import (
	"fmt"
	"os"
)

// CompletionOptions holds flags for the completion command
type CompletionOptions struct {
	Shell string // bash, zsh, or fish
}

// Completion generates shell completion scripts
func Completion(opts CompletionOptions) {
	switch opts.Shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	case "":
		fmt.Fprintf(os.Stderr, "Usage: checkpoint completion <shell>\n\n")
		fmt.Fprintf(os.Stderr, "Supported shells: bash, zsh, fish\n\n")
		fmt.Fprintf(os.Stderr, "Installation:\n")
		fmt.Fprintf(os.Stderr, "  Bash:  checkpoint completion bash >> ~/.bashrc\n")
		fmt.Fprintf(os.Stderr, "  Zsh:   checkpoint completion zsh >> ~/.zshrc\n")
		fmt.Fprintf(os.Stderr, "  Fish:  checkpoint completion fish > ~/.config/fish/completions/checkpoint.fish\n")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown shell '%s'\n", opts.Shell)
		fmt.Fprintf(os.Stderr, "supported: bash, zsh, fish\n")
		os.Exit(1)
	}
}

const bashCompletion = `# checkpoint bash completion
_checkpoint_completions() {
    local cur prev words cword
    _init_completion || return

    local commands="start summary check commit init clean doctor lint examples guide prompt explain skill search learn session help version completion"

    local examples_opts="feature bugfix refactor context anti-patterns"
    local guide_opts="first-time-user llm-workflow best-practices"
    local explain_opts="project tools guidelines skills learnings skill history next"
    local skill_opts="list show add create"
    local session_opts="show save clear handoff"

    case "${prev}" in
        checkpoint)
            COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
            return 0
            ;;
        examples)
            COMPREPLY=($(compgen -W "${examples_opts}" -- "${cur}"))
            return 0
            ;;
        guide)
            COMPREPLY=($(compgen -W "${guide_opts}" -- "${cur}"))
            return 0
            ;;
        explain)
            COMPREPLY=($(compgen -W "${explain_opts} --full --md --json" -- "${cur}"))
            return 0
            ;;
        skill)
            COMPREPLY=($(compgen -W "${skill_opts}" -- "${cur}"))
            return 0
            ;;
        session)
            COMPREPLY=($(compgen -W "${session_opts}" -- "${cur}"))
            return 0
            ;;
        commit)
            COMPREPLY=($(compgen -W "-n --dry-run --changelog-only" -- "${cur}"))
            return 0
            ;;
        init)
            COMPREPLY=($(compgen -W "--template --list-templates" -- "${cur}"))
            return 0
            ;;
        doctor)
            COMPREPLY=($(compgen -W "-v --verbose --fix" -- "${cur}"))
            return 0
            ;;
        summary)
            COMPREPLY=($(compgen -W "--json" -- "${cur}"))
            return 0
            ;;
        search)
            COMPREPLY=($(compgen -W "--failed --pattern --decision --scope --recent" -- "${cur}"))
            return 0
            ;;
        learn)
            COMPREPLY=($(compgen -W "--guideline --avoid --principle --pattern --tool" -- "${cur}"))
            return 0
            ;;
        prompt)
            COMPREPLY=($(compgen -W "session-start fill-checkpoint implement-feature fix-bug code-review --var" -- "${cur}"))
            return 0
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
            return 0
            ;;
    esac

    # Handle flags with dashes
    if [[ "${cur}" == -* ]]; then
        case "${words[1]}" in
            commit)
                COMPREPLY=($(compgen -W "-n --dry-run --changelog-only" -- "${cur}"))
                ;;
            init)
                COMPREPLY=($(compgen -W "--template --list-templates" -- "${cur}"))
                ;;
            doctor)
                COMPREPLY=($(compgen -W "-v --verbose --fix" -- "${cur}"))
                ;;
            explain)
                COMPREPLY=($(compgen -W "--full --md --json" -- "${cur}"))
                ;;
            search)
                COMPREPLY=($(compgen -W "--failed --pattern --decision --scope --recent" -- "${cur}"))
                ;;
            learn)
                COMPREPLY=($(compgen -W "--guideline --avoid --principle --pattern --tool" -- "${cur}"))
                ;;
            session)
                COMPREPLY=($(compgen -W "--status" -- "${cur}"))
                ;;
            summary)
                COMPREPLY=($(compgen -W "--json" -- "${cur}"))
                ;;
        esac
        return 0
    fi
}

complete -F _checkpoint_completions checkpoint
`

const zshCompletion = `#compdef checkpoint

_checkpoint() {
    local -a commands
    commands=(
        'start:Validate readiness and show next steps'
        'summary:Show project overview and recent activity'
        'check:Generate input file for LLM'
        'commit:Parse input, append to changelog, and git commit'
        'init:Initialize checkpoint in a project'
        'clean:Remove temporary checkpoint files'
        'doctor:Check project setup and suggest fixes'
        'lint:Check checkpoint input for issues'
        'examples:Show example checkpoint entries'
        'guide:Show detailed guides and documentation'
        'prompt:Show LLM prompts from project library'
        'explain:Get project context for LLMs and developers'
        'skill:Manage skills for LLM context'
        'search:Search checkpoint history'
        'learn:Capture knowledge during development'
        'session:Manage session state for LLM handoff'
        'help:Display help message'
        'version:Display version information'
        'completion:Generate shell completion script'
    )

    local -a examples_opts guide_opts explain_opts skill_opts session_opts prompt_opts completion_opts
    examples_opts=('feature' 'bugfix' 'refactor' 'context' 'anti-patterns')
    guide_opts=('first-time-user' 'llm-workflow' 'best-practices')
    explain_opts=('project' 'tools' 'guidelines' 'skills' 'learnings' 'skill' 'history' 'next')
    skill_opts=('list' 'show' 'add' 'create')
    session_opts=('show' 'save' 'clear' 'handoff')
    prompt_opts=('session-start' 'fill-checkpoint' 'implement-feature' 'fix-bug' 'code-review')
    completion_opts=('bash' 'zsh' 'fish')

    if (( CURRENT == 2 )); then
        _describe 'command' commands
        return
    fi

    case $words[2] in
        examples)
            if (( CURRENT == 3 )); then
                _describe 'example' examples_opts
            fi
            ;;
        guide)
            if (( CURRENT == 3 )); then
                _describe 'guide' guide_opts
            fi
            ;;
        explain)
            if (( CURRENT == 3 )); then
                _describe 'topic' explain_opts
            fi
            ;;
        skill)
            if (( CURRENT == 3 )); then
                _describe 'action' skill_opts
            fi
            ;;
        session)
            if (( CURRENT == 3 )); then
                _describe 'action' session_opts
            fi
            ;;
        prompt)
            if (( CURRENT == 3 )); then
                _describe 'prompt' prompt_opts
            fi
            ;;
        completion)
            if (( CURRENT == 3 )); then
                _describe 'shell' completion_opts
            fi
            ;;
        commit)
            _values 'flags' \
                '-n[Dry run]' \
                '--dry-run[Dry run]' \
                '--changelog-only[Stage only changelog]'
            ;;
        init)
            _values 'flags' \
                '--template[Use template]' \
                '--list-templates[List available templates]'
            ;;
        doctor)
            _values 'flags' \
                '-v[Verbose output]' \
                '--verbose[Verbose output]' \
                '--fix[Auto-fix issues]'
            ;;
        search)
            _values 'flags' \
                '--failed[Search failed approaches]' \
                '--pattern[Search patterns]' \
                '--decision[Search decisions]' \
                '--scope[Filter by scope]' \
                '--recent[Limit to recent]'
            ;;
        learn)
            _values 'flags' \
                '--guideline[Add as rule]' \
                '--avoid[Add as anti-pattern]' \
                '--principle[Add as principle]' \
                '--pattern[Add as pattern]' \
                '--tool[Add as tool]'
            ;;
        summary)
            _values 'flags' '--json[Output as JSON]'
            ;;
    esac
}

compdef _checkpoint checkpoint
`

const fishCompletion = `# checkpoint fish completion

# Disable file completion by default
complete -c checkpoint -f

# Main commands
complete -c checkpoint -n '__fish_use_subcommand' -a 'start' -d 'Validate readiness and show next steps'
complete -c checkpoint -n '__fish_use_subcommand' -a 'summary' -d 'Show project overview and recent activity'
complete -c checkpoint -n '__fish_use_subcommand' -a 'check' -d 'Generate input file for LLM'
complete -c checkpoint -n '__fish_use_subcommand' -a 'commit' -d 'Parse input, append to changelog, and git commit'
complete -c checkpoint -n '__fish_use_subcommand' -a 'init' -d 'Initialize checkpoint in a project'
complete -c checkpoint -n '__fish_use_subcommand' -a 'clean' -d 'Remove temporary checkpoint files'
complete -c checkpoint -n '__fish_use_subcommand' -a 'doctor' -d 'Check project setup and suggest fixes'
complete -c checkpoint -n '__fish_use_subcommand' -a 'lint' -d 'Check checkpoint input for issues'
complete -c checkpoint -n '__fish_use_subcommand' -a 'examples' -d 'Show example checkpoint entries'
complete -c checkpoint -n '__fish_use_subcommand' -a 'guide' -d 'Show detailed guides'
complete -c checkpoint -n '__fish_use_subcommand' -a 'prompt' -d 'Show LLM prompts'
complete -c checkpoint -n '__fish_use_subcommand' -a 'explain' -d 'Get project context'
complete -c checkpoint -n '__fish_use_subcommand' -a 'skill' -d 'Manage skills'
complete -c checkpoint -n '__fish_use_subcommand' -a 'search' -d 'Search checkpoint history'
complete -c checkpoint -n '__fish_use_subcommand' -a 'learn' -d 'Capture knowledge'
complete -c checkpoint -n '__fish_use_subcommand' -a 'session' -d 'Manage session state'
complete -c checkpoint -n '__fish_use_subcommand' -a 'help' -d 'Display help message'
complete -c checkpoint -n '__fish_use_subcommand' -a 'version' -d 'Display version'
complete -c checkpoint -n '__fish_use_subcommand' -a 'completion' -d 'Generate shell completion'

# examples subcommands
complete -c checkpoint -n '__fish_seen_subcommand_from examples' -a 'feature bugfix refactor context anti-patterns'

# guide subcommands
complete -c checkpoint -n '__fish_seen_subcommand_from guide' -a 'first-time-user llm-workflow best-practices'

# explain subcommands
complete -c checkpoint -n '__fish_seen_subcommand_from explain' -a 'project tools guidelines skills learnings skill history next'
complete -c checkpoint -n '__fish_seen_subcommand_from explain' -l full -d 'Complete context dump'
complete -c checkpoint -n '__fish_seen_subcommand_from explain' -l md -d 'Output as markdown'
complete -c checkpoint -n '__fish_seen_subcommand_from explain' -l json -d 'Output as JSON'

# skill subcommands
complete -c checkpoint -n '__fish_seen_subcommand_from skill' -a 'list show add create'

# session subcommands
complete -c checkpoint -n '__fish_seen_subcommand_from session' -a 'show save clear handoff'
complete -c checkpoint -n '__fish_seen_subcommand_from session' -l status -d 'Set status'

# prompt subcommands
complete -c checkpoint -n '__fish_seen_subcommand_from prompt' -a 'session-start fill-checkpoint implement-feature fix-bug code-review'
complete -c checkpoint -n '__fish_seen_subcommand_from prompt' -l var -d 'Set variable'

# commit flags
complete -c checkpoint -n '__fish_seen_subcommand_from commit' -s n -d 'Dry run'
complete -c checkpoint -n '__fish_seen_subcommand_from commit' -l dry-run -d 'Dry run'
complete -c checkpoint -n '__fish_seen_subcommand_from commit' -l changelog-only -d 'Stage only changelog'

# init flags
complete -c checkpoint -n '__fish_seen_subcommand_from init' -l template -d 'Use template'
complete -c checkpoint -n '__fish_seen_subcommand_from init' -l list-templates -d 'List templates'

# doctor flags
complete -c checkpoint -n '__fish_seen_subcommand_from doctor' -s v -d 'Verbose'
complete -c checkpoint -n '__fish_seen_subcommand_from doctor' -l verbose -d 'Verbose'
complete -c checkpoint -n '__fish_seen_subcommand_from doctor' -l fix -d 'Auto-fix'

# search flags
complete -c checkpoint -n '__fish_seen_subcommand_from search' -l failed -d 'Search failed approaches'
complete -c checkpoint -n '__fish_seen_subcommand_from search' -l pattern -d 'Search patterns'
complete -c checkpoint -n '__fish_seen_subcommand_from search' -l decision -d 'Search decisions'
complete -c checkpoint -n '__fish_seen_subcommand_from search' -l scope -d 'Filter by scope'
complete -c checkpoint -n '__fish_seen_subcommand_from search' -l recent -d 'Limit to recent'

# learn flags
complete -c checkpoint -n '__fish_seen_subcommand_from learn' -l guideline -d 'Add as rule'
complete -c checkpoint -n '__fish_seen_subcommand_from learn' -l avoid -d 'Add as anti-pattern'
complete -c checkpoint -n '__fish_seen_subcommand_from learn' -l principle -d 'Add as principle'
complete -c checkpoint -n '__fish_seen_subcommand_from learn' -l pattern -d 'Add as pattern'
complete -c checkpoint -n '__fish_seen_subcommand_from learn' -l tool -d 'Add as tool'

# summary flags
complete -c checkpoint -n '__fish_seen_subcommand_from summary' -l json -d 'Output as JSON'

# completion subcommands
complete -c checkpoint -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'
`
