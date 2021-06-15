package main

import (
	"github.com/argoproj-labs/argo-cloudops/cli/cmd"
)

// TODO Populate during build/release.
const (
	version = "0.2.1"
)

func main() {
	cmd.Execute(version)
}
