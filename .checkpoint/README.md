# Checkpoint Support Files

This directory contains supporting materials for the checkpoint workflow.

## Directory Structure

- **examples/** - Example checkpoint entries showing best practices
  - Good examples of features, bug fixes, refactorings
  - Context examples showing effective decision capture
  - Anti-patterns to avoid

- **guides/** - Detailed documentation for checkpoint users
  - First-time user walkthrough
  - LLM integration patterns
  - Context writing guidelines
  - Best practices

- **prompts/** - LLM prompt templates
  - Session start prompts
  - Checkpoint filling prompts
  - Feature implementation and bug fix prompts

- **skills/** - Skill definitions for LLM context
  - Local project-specific skills
  - References to global skills

## Usage

These files are referenced by checkpoint commands and can be read directly:

- Run `checkpoint examples` to view examples
- Run `checkpoint guide [topic]` to view guides
- LLMs can read these files directly when filling checkpoint entries

## Maintenance

- This directory is tracked in git
- Add new examples as you develop useful patterns
- Update guides as the workflow evolves
- Customize for your project's specific needs
