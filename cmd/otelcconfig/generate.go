// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/codegen"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

func runGenerate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	manifestDir := fs.String("manifest", "./manifest", "directory containing instrumentation behavior manifests")
	outputDir := fs.String("output", "./generated", "directory where generated artifacts are written")
	check := fs.Bool("check", false, "verify committed outputs match manifests without writing")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(stderr, "generate: unexpected arguments: %v\n", fs.Args())
		return 2
	}

	manifests, err := manifest.Discover(*manifestDir)
	if err != nil {
		fmt.Fprintf(stderr, "generate: %v\n", err)
		return 1
	}
	if len(manifests) == 0 {
		fmt.Fprintf(stderr, "generate: no manifests found under %s\n", *manifestDir)
		return 1
	}

	outputs, err := codegen.GenerateAll(manifests)
	if err != nil {
		fmt.Fprintf(stderr, "generate: %v\n", err)
		return 1
	}

	if *check {
		return checkOutputs(outputs, *outputDir, stdout, stderr)
	}
	if err := writeOutputs(outputs, *outputDir); err != nil {
		fmt.Fprintf(stderr, "generate: %v\n", err)
		return 1
	}
	for _, m := range manifests {
		fmt.Fprintf(stdout, "generated %s\n", m.Instrumentation.Name)
	}
	fmt.Fprintf(stdout, "generated %d artifacts under %s\n", len(outputs), *outputDir)
	return 0
}

func writeOutputs(outputs codegen.Outputs, outputDir string) error {
	for _, path := range sortedPaths(outputs) {
		full := filepath.Join(outputDir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, outputs[path], 0o644); err != nil {
			return fmt.Errorf("write %s: %w", full, err)
		}
	}
	return nil
}

func checkOutputs(outputs codegen.Outputs, outputDir string, stdout, stderr io.Writer) int {
	var drifted []string
	for _, path := range sortedPaths(outputs) {
		full := filepath.Join(outputDir, filepath.FromSlash(path))
		want, err := os.ReadFile(full)
		if err != nil {
			drifted = append(drifted, path+" (missing)")
			continue
		}
		if !bytes.Equal(want, outputs[path]) {
			drifted = append(drifted, path+" (out of date)")
		}
	}
	if len(drifted) > 0 {
		fmt.Fprintf(stderr, "generate --check: %d artifact(s) drifted:\n", len(drifted))
		for _, d := range drifted {
			fmt.Fprintf(stderr, "  - %s\n", d)
		}
		fmt.Fprintln(stderr, "run `otelcconfig generate` and commit the regenerated files")
		return 1
	}
	fmt.Fprintf(stdout, "generate --check: %d artifact(s) up to date\n", len(outputs))
	return 0
}

func sortedPaths(outputs codegen.Outputs) []string {
	paths := make([]string, 0, len(outputs))
	for p := range outputs {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}
