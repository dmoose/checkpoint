package main

import (
	"github.com/dmoose/checkpoint/cmd"
)

// Build info - injected via ldflags at build time
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cmd.Execute(cmd.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
}
