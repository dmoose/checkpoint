package templates

// Python CLI template
const pythonCliProjectYml = `schema_version: "1"
name: "{{project_name}}"
type: python-cli
purpose: |
  Python command-line application.
  (Update this with your project's specific purpose)

architecture:
  overview: |
    Python CLI with entry point in __main__.py or cli.py.
    Core logic in src/ or package directory.

  key_paths:
    entry_point: src/__main__.py
    core: src/
    tests: tests/

languages:
  primary: python
  version: "3.10+"

dependencies:
  external: []
`

const pythonCliToolsYml = `schema_version: "1"

build:
  default:
    command: pip install -e .
    notes: Install in development mode

  dist:
    command: python -m build
    notes: Build distribution packages

test:
  default:
    command: pytest
    notes: Run all tests

  coverage:
    command: pytest --cov=src
    notes: Run tests with coverage

  verbose:
    command: pytest -v
    notes: Verbose test output

lint:
  default:
    command: ruff check .
    notes: Run Ruff linter

  fix:
    command: ruff check --fix .
    notes: Auto-fix linting issues

  format:
    command: ruff format .
    notes: Format code

  type:
    command: mypy src
    notes: Run type checker

check:
  default:
    command: ruff check . && pytest
    notes: Lint and test

run:
  default:
    command: python -m src
    notes: Run the application

maintenance:
  deps:
    command: pip install -r requirements.txt
    notes: Install dependencies

  update:
    command: pip install -U -r requirements.txt
    notes: Update dependencies

  venv:
    command: python -m venv .venv && source .venv/bin/activate
    notes: Create virtual environment
`

const pythonCliGuidelinesYml = `schema_version: "1"

naming:
  files:
    pattern: snake_case.py
    examples:
      - user_service.py
      - cli.py

  modules:
    pattern: snake_case
    examples:
      - src/utils
      - src/models

  functions:
    pattern: snake_case
    examples:
      - get_user_by_id
      - validate_input

  classes:
    pattern: PascalCase
    examples:
      - UserService
      - ConfigManager

structure:
  new_module: |
    1. Create src/{name}.py or src/{name}/__init__.py
    2. Add tests in tests/test_{name}.py

errors:
  style: |
    Use specific exception types.
    Create custom exceptions for domain errors.
    Use try-except at boundaries.

testing:
  framework: pytest
  pattern: "tests/test_{name}.py"
  style: Use descriptive test function names with test_ prefix

commits:
  tool: Use checkpoint commit for all commits
  pre_commit: Run pytest before committing

rules:
  - Use type hints for function signatures
  - Document public functions with docstrings
  - Use virtual environments
  - Pin dependency versions

avoid:
  - Bare except clauses
  - Mutable default arguments
  - Import * in production code
  - Global mutable state

principles:
  - "Explicit is better than implicit"
  - "Readability counts"
  - "Errors should never pass silently"
`

const pythonCliSkillsYml = `schema_version: "1"

global:
  - git
  - python

local: []

config:
  python:
    version: "3.10+"
`
