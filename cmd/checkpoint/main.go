package main

import "github.com/dmoose/checkpoint/internal/app/checkpoint"

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	checkpoint.Execute(checkpoint.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
}
