// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Command otelcconfig is the CLI entrypoint for the otelcconfig toolkit.
//
// Phase 4 implements every planned command: generate, validate, resolve,
// explain, catalog, bake, guard, and diff. See docs/architecture.md for the
// full roadmap.
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
		return writeUsage(stdout, stderr)
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "version", "--version", "-v":
		if _, err := fmt.Fprintf(stdout, "otelcconfig version %s\n", version); err != nil {
			reportWriteError(stderr, err)
			return 1
		}
		return 0
	case "help", "--help", "-h":
		return writeUsage(stdout, stderr)
	case "generate":
		return runGenerate(rest, stdout, stderr)
	case "validate":
		return runValidate(rest, stdout, stderr)
	case "resolve":
		return runResolve(rest, stdout, stderr)
	case "explain":
		return runExplain(rest, stdout, stderr)
	case "catalog":
		return runCatalog(rest, stdout, stderr)
	case "bake":
		return runBake(rest, stdout, stderr)
	case "diff":
		return runDiff(rest, stdout, stderr)
	case "guard":
		return runGuard(rest, stdout, stderr)
	default:
		if _, err := fmt.Fprintf(stderr, "unknown command %q\n\n", cmd); err != nil {
			return 1
		}
		if err := printUsage(stderr); err != nil {
			return 1
		}
		return 2
	}
}

func writeUsage(stdout, stderr io.Writer) int {
	if err := printUsage(stdout); err != nil {
		reportWriteError(stderr, err)
		return 1
	}
	return 0
}

func reportWriteError(stderr io.Writer, err error) {
	_, _ = fmt.Fprintf(stderr, "write output: %v\n", err)
}

func printUsage(w io.Writer) error {
	_, err := fmt.Fprint(w, strings.TrimSpace(`
otelcconfig — Declarative Configuration Toolkit for otelc

Independent prototype exploring OpenTelemetry Go Compile Instrumentation Issue #705.
Not an official OpenTelemetry component.

Usage:
  otelcconfig <command> [flags]

Commands:
  version     Print version and exit
  help        Show this help
  generate    Generate types, defaults, env map, schema, and docs from manifests
  validate    Validate a declarative configuration file       <file>
  resolve     Show resolved values and their sources          <file>
  explain     Explain one configuration option                <option-or-path>
  catalog     List all configurable options
  bake        Freeze resolved configuration into a Go package   <file>
  guard       Reject undeclared configuration access            <dir...>
  diff        Compare two configuration files                   <file> <file>
              exit 0 identical, 1 differences, 2 usage error

Generate flags:
  --manifest <dir>   Behavior manifests directory (default ./manifest)
  --output <dir>     Generated artifacts directory (default ./generated)
  --check            Verify committed outputs are up to date without writing

Config command flags:
  --manifest <dir>   Behavior manifests directory (default ./manifest)

Examples:
  otelcconfig version
  otelcconfig generate
  otelcconfig generate --check
  otelcconfig validate examples/nethttp.yaml
  otelcconfig resolve  examples/nethttp.yaml
  otelcconfig explain  request_captured_headers
  otelcconfig catalog
  otelcconfig bake    --output baked --check examples/demo.yaml
  otelcconfig guard   ./demo ./runtime ./baked
  otelcconfig diff    examples/minimal.yaml examples/nethttp.yaml

Docs: https://github.com/ADITYA-CODE-SOURCE/otelcconfig
`)+"\n")
	return err
}
