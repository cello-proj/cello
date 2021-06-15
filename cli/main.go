package main

import (
	"github.com/argoproj-labs/argo-cloudops/cli/cmd"
)

// TODO set this and align it to service version
const (
	version = "0.1.1"
)

func main() {
	cmd.Execute(version)
}
