package templates

// Generic project template
const genericProjectYml = `schema_version: "1"
name: "{{project_name}}"
type: generic
purpose: |
  (Describe your project's purpose here)

architecture:
  overview: |
    (Describe your project's architecture here)

  key_paths:
    # Add your project's key directories/files
    src: src/

languages:
  primary: unknown
  version: ""

dependencies:
  external: []

integrations: []

# AI collaboration boundaries - what AI can do without asking
ai_authority:
  autonomous:
    - "Bug fixes with clear reproduction steps"
    - "Test additions for existing code"
    - "Documentation updates"
    - "Code formatting and style fixes"
  requires_approval:
    - "New features"
    - "Architecture changes"
    - "Dependency additions"
    - "API or interface changes"
    - "Database schema changes"
  notes: ""

# Project-wide lessons from failed approaches (accumulates over time)
lessons_learned: []

# Project roadmap and deferred items
roadmap:
  planned: []
  deferred: []
`

const genericToolsYml = `schema_version: "1"

# Add your build commands
build:
  default:
    command: echo "No build configured"
    notes: Configure your build command

# Add your test commands
test:
  default:
    command: echo "No tests configured"
    notes: Configure your test command

# Add your lint/check commands
lint:
  default:
    command: echo "No linter configured"
    notes: Configure your linter

# Pre-commit checks
check:
  default:
    command: echo "No checks configured"
    notes: Add your pre-commit checks

# Run commands
run:
  default:
    command: echo "No run command configured"
    notes: Configure how to run your project

# Maintenance commands
maintenance:
  deps:
    command: echo "No dependency command configured"
    notes: Configure dependency installation

# Pre-commit verification (runs automatically before checkpoint commit)
verify:
  pre_commit: []
  # Example:
  # pre_commit:
  #   - command: make check
  #     description: Run all quality checks
  #     required: true
  #   - command: go build ./...
  #     description: Ensure compilation
  #     required: true
`

const genericGuidelinesYml = `schema_version: "1"

# Naming conventions for your project
naming:
  files:
    pattern: "(define your file naming pattern)"
    examples: []

  functions:
    pattern: "(define your function naming pattern)"
    examples: []

# Project structure patterns
structure:
  new_feature: |
    (Describe how to add new features)

# Error handling approach
errors:
  style: |
    (Describe your error handling approach)

# Testing approach
testing:
  pattern: "(describe test file location)"
  style: "(describe testing style)"

# Commit practices
commits:
  tool: Use checkpoint commit for all commits
  pre_commit: "(describe pre-commit checks)"

# Rules to follow
rules:
  - (Add your project rules)

# Things to avoid
avoid:
  - (Add anti-patterns to avoid)

# Guiding principles
principles:
  - (Add your guiding principles)

# Human-AI collaboration protocol
collaboration:
  protocol: "propose → approve → implement → summarize"
  principles:
    - "Ask for clarification rather than assuming"
    - "Human approves architecture and features, AI proposes implementation"
    - "Summarize completed work before moving to next task"
    - "Reference project patterns before introducing new ones"
    - "Check ai_authority in project.yaml for what requires approval"
`

const genericSkillsYml = `schema_version: "1"

global:
  - git

local: []

config: {}
`
