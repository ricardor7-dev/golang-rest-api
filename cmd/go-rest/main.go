package main

import (
	"fmt"
	"golang-rest-api/cmd"
	"os"
)

func main() {
	if err := cmd.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error from commands.Run(): %s\n", err)
		os.Exit(1)
		}
	}