package main

import (
	"fmt"

	"github.com/cello-proj/cello/cli/cmd"
)

var (
	// Populated during build/release
	commit  string
	date    string
	version string
)

func main() {
	cmd.Execute(fmt.Sprintf("%s %s %s", version, commit, date))
}
