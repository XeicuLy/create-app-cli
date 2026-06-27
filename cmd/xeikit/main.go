package main

import (
	"os"

	"github.com/XeicuLy/create-app-cli/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.NewRootCmd(version).Execute(); err != nil {
		os.Exit(1)
	}
}
