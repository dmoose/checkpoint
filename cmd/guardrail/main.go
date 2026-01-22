package main

import "github.com/dmoose/checkpoint/internal/app/guardrail"

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	guardrail.Execute(guardrail.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
}
