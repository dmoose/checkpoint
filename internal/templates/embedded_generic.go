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
`

const genericSkillsYml = `schema_version: "1"

global:
  - git

local: []

config: {}
`
