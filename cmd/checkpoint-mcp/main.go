package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/dmoose/checkpoint/internal/mcpserver"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	projectPath := flag.String("project", ".", "Path to the project directory")
	flag.Parse()

	absPath, err := filepath.Abs(*projectPath)
	if err != nil {
		log.Fatalf("Failed to resolve project path: %v", err)
	}

	s := mcpserver.NewServer(absPath)
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
