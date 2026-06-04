package main

import (
	"fmt"
	"os"

	"github.com/ravistakumar/cast/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("cast", version.String())
		return
	}
	fmt.Fprintln(os.Stderr, "cast: no command (cli wired in a later task)")
	os.Exit(1)
}
