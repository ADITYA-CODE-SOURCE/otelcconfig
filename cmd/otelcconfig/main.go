// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Command otelcconfig is the CLI entrypoint for the otelcconfig toolkit.
//
// Phase 0 provides version/help and honest stubs for later-phase commands.
// See docs/architecture.md for the full roadmap.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// version is set by -ldflags when building a release.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "otelcconfig version %s\n", version)
		return 0
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	case "generate", "validate", "resolve", "explain", "catalog", "diff", "bake", "guard":
		return notImplemented(cmd, rest, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", cmd)
		printUsage(stderr)
		return 2
	}
}

func notImplemented(cmd string, _ []string, stderr io.Writer) int {
	phase := phaseFor(cmd)
	fmt.Fprintf(stderr, "command %q is not implemented in Phase 0 (planned: %s)\n", cmd, phase)
	fmt.Fprintln(stderr, "See docs/architecture.md and the README roadmap.")
	return 1
}

func phaseFor(cmd string) string {
	switch cmd {
	case "generate":
		return "Phase 1 (v0.2.0)"
	case "validate", "resolve", "explain", "catalog":
		return "Phase 2 (v0.3.0)"
	case "bake":
		return "Phase 3 (v0.4.0)"
	case "guard", "diff":
		return "Phase 4 (v0.5.0)"
	default:
		return "a later phase"
	}
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, strings.TrimSpace(`
otelcconfig — Declarative Configuration Toolkit for otelc

Independent prototype exploring OpenTelemetry Go Compile Instrumentation Issue #705.
Not an official OpenTelemetry component.

Usage:
  otelcconfig <command> [flags]

Commands:
  version     Print version and exit
  help        Show this help

Planned (not implemented in Phase 0):
  generate    Generate types, defaults, schema, and docs from manifests   [Phase 1]
  validate    Validate a declarative configuration file                 [Phase 2]
  resolve     Show resolved values and their sources                    [Phase 2]
  explain     Explain one configuration option                          [Phase 2]
  catalog     List all configurable options                             [Phase 2]
  bake        Model build-time config embedding (not otelc-integrated)  [Phase 3]
  guard       Reject undeclared configuration access                    [Phase 4]
  diff        Compare two configuration files                           [Phase 4]

Examples:
  otelcconfig version
  otelcconfig help

Docs: https://github.com/ADITYA-CODE-SOURCE/otelcconfig
`)+"\n")
}
