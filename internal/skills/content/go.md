# Go

Go programming language toolchain.

## Purpose

Build, test, and manage Go applications.

## Common Commands

```bash
go build .              # Build current package
go run .                # Build and run
go test ./...           # Run all tests
go test -v ./...        # Verbose tests
go test -cover ./...    # With coverage
go test -race ./...     # With race detector
go fmt ./...            # Format code
go vet ./...            # Static analysis
go mod init <name>      # Initialize module
go mod tidy             # Clean up go.mod
go mod download         # Download dependencies
go get <pkg>            # Add dependency
go get -u ./...         # Update all dependencies
```

## Project Structure

```
myproject/
├── main.go             # Entry point
├── go.mod              # Module definition
├── cmd/                # Command implementations
├── internal/           # Private packages
├── pkg/                # Public packages
└── *_test.go           # Test files
```

## Tips

- Run go fmt before committing
- Use go vet to catch common mistakes
- Use -race flag during development
- Keep go.mod tidy with go mod tidy
- Use internal/ for packages not meant for external use
