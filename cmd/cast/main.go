package main

import (
	"fmt"
	"os"

	"github.com/ravistakumar/cast/internal/cli"
	"github.com/ravistakumar/cast/internal/version"
)

func main() {
	root := cli.New()
	root.Version = version.String()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "cast:", err)
		os.Exit(1)
	}
}
