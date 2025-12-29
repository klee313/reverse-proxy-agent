// Command rpa is the CLI entrypoint; it delegates argument handling to internal/cli.
// This binary is the main entry for running and managing the agent.

package main

import (
	"os"

	"reverse-proxy-agent/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
