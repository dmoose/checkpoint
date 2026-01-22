# Getting Started

Complete setup guide for checkpoint and guardrail.

## Prerequisites

- **Go 1.25+** -- [install](https://go.dev/doc/install)
- **git** -- any recent version

## Install both binaries

### From source

```bash
git clone https://github.com/dmoose/checkpoint.git
cd checkpoint
make install-user    # Installs to ~/.local/bin
```

### With go install

```bash
go install github.com/dmoose/checkpoint/cmd/checkpoint@latest
go install github.com/dmoose/checkpoint/cmd/guardrail@latest
```

Verify both are available:

```bash
checkpoint version
guardrail --help
```

## Initialize a project

Run both init commands in your project root:

```bash
cd your-project
checkpoint init
guardrail init
```

### What `checkpoint init` creates

- `.checkpoint-changelog.yaml` -- Empty changelog, ready for entries
- `.checkpoint-context.yaml` -- Empty context file
- `.gitignore` entries for temporary files (`checkpoint-input`, `.checkpoint-diff`, `.checkpoint-status.yaml`)

### What `guardrail init` creates

- `.checkpoint/project.yaml` -- Project overview, architecture, AI authority boundaries
- `.checkpoint/tools.yaml` -- Build, test, lint, and run commands
- `.checkpoint/guidelines.yaml` -- Coding standards, naming conventions, rules
- `.checkpoint/skills.yaml` -- LLM skill references (ripgrep, git, language toolchain)
- `.checkpoint/skills/` -- Directory for local skill definitions

Guardrail auto-detects your project language, build system, and common tools during init.

### Configure your project

Edit the generated files to match your project:

```bash
# Add project purpose, architecture overview, AI authority rules
$EDITOR .checkpoint/project.yaml

# Set build, test, lint commands
$EDITOR .checkpoint/tools.yaml

# Define coding standards, patterns to follow, anti-patterns to avoid
$EDITOR .checkpoint/guidelines.yaml
```

## Your first checkpoint

1. **Make a change** to your code -- fix a bug, add a feature, refactor something.

2. **Generate the input file:**

   ```bash
   checkpoint check
   ```

   This creates `checkpoint-input` (a YAML template) and `.checkpoint-diff` (the git diff for reference).

3. **Fill in `checkpoint-input`:**

   Open the file and describe your changes. If you are working with an LLM, it can fill this in for you -- but you review it before committing.

   At minimum, fill in:
   - `changes[].summary` -- What changed (present tense, under 80 chars)
   - `changes[].change_type` -- One of: `feature`, `fix`, `refactor`, `docs`, `perf`, `other`
   - `context.problem_statement` -- What problem this solves

4. **Validate (optional):**

   ```bash
   checkpoint lint
   ```

   Catches placeholder text, vague summaries, and schema issues.

5. **Commit:**

   ```bash
   checkpoint commit
   ```

   This validates the input, appends to `.checkpoint-changelog.yaml`, extracts context to `.checkpoint-context.yaml`, and creates a git commit.

## Verify your setup

```bash
# Check that everything is wired up correctly
guardrail doctor

# See a summary of your changelog
checkpoint summary

# View the project context an LLM would see
guardrail explain
```

## Next steps

- [Workflow](workflow.md) -- The daily workflow for humans and LLMs
- [LLM Integration](llm-integration.md) -- Setting up Claude Code, Cursor, Aider, and other tools
