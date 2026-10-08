package main

import (
	"fmt"
	"os"

	"github.com/agentdock/agentdock/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "agentdock:", err)
		os.Exit(cli.ExitCode(err))
	}
}
