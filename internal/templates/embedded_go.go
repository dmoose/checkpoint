package templates

// Go CLI template
const goCliProjectYml = `schema_version: "1"
name: "{{project_name}}"
type: go-cli
purpose: |
  Go command-line application.
  (Update this with your project's specific purpose)

architecture:
  overview: |
    Standard Go CLI structure with main.go entry point.
    Commands in cmd/, core logic in internal/, config in pkg/.

  key_paths:
    entry_point: main.go
    commands: cmd/
    core_logic: internal/
    config: pkg/

languages:
  primary: go
  version: "1.21"

dependencies:
  external: []
`

const goCliToolsYml = `schema_version: "1"

build:
  default:
    command: go build -o bin/{{project_name}} .
    output: bin/{{project_name}}
    notes: Build for current platform

  install:
    command: go install .
    notes: Install to GOPATH/bin

test:
  default:
    command: go test ./...
    notes: Run all tests

  coverage:
    command: go test -cover ./...
    notes: Run tests with coverage

  verbose:
    command: go test -v ./...
    notes: Verbose test output

  race:
    command: go test -race ./...
    notes: Run with race detector

lint:
  default:
    command: go vet ./...
    notes: Run go vet

  fmt:
    command: go fmt ./...
    notes: Format code

check:
  default:
    command: go fmt ./... && go vet ./... && go test ./...
    notes: Format, vet, and test

run:
  default:
    command: go run .
    notes: Run the application

maintenance:
  deps:
    command: go mod download
    notes: Download dependencies

  tidy:
    command: go mod tidy
    notes: Clean up go.mod

  update:
    command: go get -u ./...
    notes: Update dependencies
`

const goCliGuidelinesYml = `schema_version: "1"

naming:
  files:
    pattern: snake_case.go
    examples:
      - main.go
      - user_service.go

  packages:
    pattern: lowercase, single word
    examples:
      - internal/config
      - pkg/utils

  functions:
    exported: PascalCase
    internal: camelCase

structure:
  new_package: |
    1. Create internal/{name}/{name}.go or pkg/{name}/{name}.go
    2. Keep package focused on single responsibility
    3. Add {name}_test.go for tests

errors:
  style: |
    Wrap errors with context using fmt.Errorf:
    return fmt.Errorf("operation failed: %w", err)

testing:
  pattern: "{file}_test.go in same package"
  naming: "Test{Function}_{Scenario}"
  style: Use table-driven tests for multiple scenarios

commits:
  tool: Use checkpoint commit for all commits
  pre_commit: Run go test ./... before committing

rules:
  - Run tests before committing
  - Use go fmt for formatting
  - Handle all errors explicitly
  - Keep functions focused and small

avoid:
  - Global mutable state
  - Panic for recoverable errors
  - Ignoring errors with _

principles:
  - "Explicit is better than implicit"
  - "Return errors, don't panic"
  - "Accept interfaces, return structs"
`

const goCliSkillsYml = `schema_version: "1"

global:
  - go
  - git

local: []

config:
  go:
    version: "1.21"
`

// Go Library template
const goLibProjectYml = `schema_version: "1"
name: "{{project_name}}"
type: go-lib
purpose: |
  Go library package.
  (Update this with your library's specific purpose)

architecture:
  overview: |
    Go library with public API in root package.
    Internal implementation in internal/.

  key_paths:
    public_api: "*.go (root)"
    internal: internal/

languages:
  primary: go
  version: "1.21"

dependencies:
  external: []
`

const goLibToolsYml = `schema_version: "1"

test:
  default:
    command: go test ./...
    notes: Run all tests

  coverage:
    command: go test -cover ./...
    notes: Run tests with coverage

  bench:
    command: go test -bench=. ./...
    notes: Run benchmarks

lint:
  default:
    command: go vet ./...
    notes: Run go vet

  fmt:
    command: go fmt ./...
    notes: Format code

check:
  default:
    command: go fmt ./... && go vet ./... && go test ./...
    notes: Format, vet, and test

maintenance:
  deps:
    command: go mod download
    notes: Download dependencies

  tidy:
    command: go mod tidy
    notes: Clean up go.mod
`
