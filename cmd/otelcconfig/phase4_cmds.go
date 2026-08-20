// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"io"
	"os"
	"reflect"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/config"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/internal/guard"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

func runGuard(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("guard", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	dirs := fs.Args()
	diags, err := guard.Run(dirs)
	if err != nil {
		return reportError(stderr, "guard", err)
	}
	if len(diags) == 0 {
		if err := writeErrf(stdout, "guard: no undeclared configuration access found\n"); err != nil {
			return 1
		}
		return 0
	}
	for _, d := range diags {
		if err := writeErrf(stdout, "%s: %s\n", d.Pos, d.Message); err != nil {
			return 1
		}
	}
	if err := writeErrf(stdout, "guard: %d violation(s)\n", len(diags)); err != nil {
		return 1
	}
	return 1
}

func runDiff(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	manifestDir := fs.String("manifest", "./manifest", "behavior manifests directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 2 {
		if err := writeErrf(stderr, "diff: expected exactly two configuration files\n"); err != nil {
			return 1
		}
		return 2
	}

	manifests, ok := discoverManifests(*manifestDir, stderr)
	if !ok {
		return 1
	}
	left, ok := resolveFile("diff", fs.Arg(0), manifests, stderr)
	if !ok {
		return 1
	}
	right, ok := resolveFile("diff", fs.Arg(1), manifests, stderr)
	if !ok {
		return 1
	}

	lines := compareResolved(manifests, left, right)
	if len(lines) == 0 {
		if err := writeErrf(stdout, "%s and %s: identical\n", fs.Arg(0), fs.Arg(1)); err != nil {
			return 1
		}
		return 0
	}
	for _, l := range lines {
		if err := writeErrf(stdout, "%s\n", l); err != nil {
			return 1
		}
	}
	return 1
}

// resolveFile resolves one configuration file into a per-option lookup keyed by
// instrumentation then option path.
func resolveFile(cmd, path string, manifests []*manifest.Manifest, stderr io.Writer) (map[string]map[string]config.ResolvedOption, bool) {
	dev, ok := loadDevelopment(cmd, path, manifests, stderr)
	if !ok {
		return nil, false
	}
	resolved, err := config.Resolve(manifests, dev, os.LookupEnv)
	if err != nil {
		reportError(stderr, cmd, err)
		return nil, false
	}
	out := make(map[string]map[string]config.ResolvedOption, len(resolved))
	for _, r := range resolved {
		opts := make(map[string]config.ResolvedOption, len(r.Options))
		for _, o := range r.Options {
			opts[o.Path] = o
		}
		out[r.Instrumentation] = opts
	}
	return out, true
}

// compareResolved produces one line per differing (or equal) option, in
// manifest order, with = for identical values and ! for changes.
func compareResolved(manifests []*manifest.Manifest, left, right map[string]map[string]config.ResolvedOption) []string {
	var lines []string
	for _, m := range manifests {
		lo := left[m.Instrumentation.Name]
		ro := right[m.Instrumentation.Name]
		for _, o := range m.Options {
			path := o.DeclarativePath
			lv := lo[path]
			rv := ro[path]
			if reflect.DeepEqual(lv.Value, rv.Value) {
				continue
			}
			lines = append(lines, "! "+path+" "+valueDisplay(lv.Value)+" -> "+valueDisplay(rv.Value))
		}
	}
	return lines
}
