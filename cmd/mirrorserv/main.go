package main

import (
	"mirrorserv/pkg/entrypoint"
	"os"
)

func main() {
	// The real entry point is in the entrypoint package, which allows for efficient test integration.
	// Do not add more code to this file (it should also be excluded from coverage tracking).
	exitCode := entrypoint.Entrypoint(os.Stdout, os.Stderr)
	os.Exit(exitCode)
}
