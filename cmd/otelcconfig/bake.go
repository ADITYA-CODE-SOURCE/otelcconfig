// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"flag"
	"io"
	"os"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/bake"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/config"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

func runBake(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("bake", flag.ContinueOnError)
	fs.SetOutput(stderr)
	manifestDir := fs.String("manifest", "./manifest", "behavior manifests directory")
	outputDir := fs.String("output", "./baked", "directory where frozen configuration is written")
	check := fs.Bool("check", false, "verify committed baked output matches current resolution without writing")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		if err := writeErrf(stderr, "bake: expected exactly one configuration file\n"); err != nil {
			return 1
		}
		return 2
	}

	manifests, ok := discoverManifests(*manifestDir, stderr)
	if !ok {
		return 1
	}
	dev, ok := loadDevelopment("bake", fs.Arg(0), manifests, stderr)
	if !ok {
		return 1
	}
	resolved, err := config.Resolve(manifests, dev, os.LookupEnv)
	if err != nil {
		return reportError(stderr, "bake", err)
	}
	outputs, err := bake.Generate(resolved)
	if err != nil {
		return reportError(stderr, "bake", err)
	}

	if *check {
		return checkBakedOutputs(outputs, *outputDir, stdout, stderr)
	}
	// Environment overrides resolved above are frozen into the output, so the
	// baked binary is unaffected by later environment changes.
	if err := writeOutputs(outputs, *outputDir); err != nil {
		return reportError(stderr, "bake", err)
	}
	for _, p := range sortedPaths(outputs) {
		if err := writeErrf(stdout, "baked %s\n", p); err != nil {
			return 1
		}
	}
	if err := writeErrf(stdout, "baked %d artifact(s) under %s\n", len(outputs), *outputDir); err != nil {
		return 1
	}
	return 0
}

// loadDevelopment reads, parses, substitutes, and validates a configuration
// file, returning the validated instrumentation.development node.
func loadDevelopment(cmd, path string, manifests []*manifest.Manifest, stderr io.Writer) (any, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		reportError(stderr, cmd, err)
		return nil, false
	}
	doc, err := config.Parse(data)
	if err != nil {
		reportError(stderr, cmd, err)
		return nil, false
	}
	dev, err := config.Substitute(doc.Development, os.LookupEnv)
	if err != nil {
		reportError(stderr, cmd, err)
		return nil, false
	}
	validators, err := config.Validators(manifests)
	if err != nil {
		reportError(stderr, cmd, err)
		return nil, false
	}
	for _, m := range manifests {
		if verr := validators[m.Instrumentation.Name].Validate(dev); verr != nil {
			reportError(stderr, cmd, verr)
			return nil, false
		}
	}
	return dev, true
}

// checkBakedOutputs verifies committed baked output matches current resolution
// without writing, mirroring generate --check.
func checkBakedOutputs(outputs map[string][]byte, outputDir string, stdout, stderr io.Writer) int {
	var drifted []string
	for _, path := range sortedPaths(outputs) {
		full := outputDir + string(os.PathSeparator) + path
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
		if err := writeErrf(stderr, "bake --check: %d artifact(s) drifted:\n", len(drifted)); err != nil {
			return 1
		}
		for _, d := range drifted {
			if err := writeErrf(stderr, "  - %s\n", d); err != nil {
				return 1
			}
		}
		if err := writeErrf(stderr, "run `otelcconfig bake` and commit the regenerated files\n"); err != nil {
			return 1
		}
		return 1
	}
	if err := writeErrf(stdout, "bake --check: %d artifact(s) up to date\n", len(outputs)); err != nil {
		return 1
	}
	return 0
}
