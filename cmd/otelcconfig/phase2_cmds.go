// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/config"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

func runValidate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	manifestDir := fs.String("manifest", "./manifest", "behavior manifests directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		if err := writeErrf(stderr, "validate: expected exactly one configuration file\n"); err != nil {
			return 1
		}
		return 2
	}

	manifests, ok := discoverManifests(*manifestDir, stderr)
	if !ok {
		return 1
	}
	validators, err := config.Validators(manifests)
	if err != nil {
		if writeErrf(stderr, "validate: %v\n", err) != nil {
			return 1
		}
		return 1
	}

	path := fs.Arg(0)
	data, err := os.ReadFile(path)
	if err != nil {
		if writeErrf(stderr, "validate: %v\n", err) != nil {
			return 1
		}
		return 1
	}
	doc, err := config.Parse(data)
	if err != nil {
		if writeErrf(stderr, "validate: %v\n", err) != nil {
			return 1
		}
		return 1
	}
	dev, err := config.Substitute(doc.Development, os.LookupEnv)
	if err != nil {
		if writeErrf(stderr, "validate: %v\n", err) != nil {
			return 1
		}
		return 1
	}

	for _, m := range manifests {
		if verr := validators[m.Instrumentation.Name].Validate(dev); verr != nil {
			if err := writeErrf(stderr, "%s: %v\n", path, verr); err != nil {
				return 1
			}
			return 1
		}
	}
	if err := writeErrf(stdout, "%s: valid\n", path); err != nil {
		return 1
	}
	return 0
}

func runResolve(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("resolve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	manifestDir := fs.String("manifest", "./manifest", "behavior manifests directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		if err := writeErrf(stderr, "resolve: expected exactly one configuration file\n"); err != nil {
			return 1
		}
		return 2
	}

	manifests, ok := discoverManifests(*manifestDir, stderr)
	if !ok {
		return 1
	}
	validators, err := config.Validators(manifests)
	if err != nil {
		return reportError(stderr, "resolve", err)
	}

	path := fs.Arg(0)
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return reportError(stderr, "resolve", readErr)
	}
	doc, err := config.Parse(data)
	if err != nil {
		return reportError(stderr, "resolve", err)
	}
	dev, err := config.Substitute(doc.Development, os.LookupEnv)
	if err != nil {
		return reportError(stderr, "resolve", err)
	}
	for _, m := range manifests {
		if verr := validators[m.Instrumentation.Name].Validate(dev); verr != nil {
			return reportError(stderr, "resolve", verr)
		}
	}

	resolved, err := config.Resolve(manifests, dev, os.LookupEnv)
	if err != nil {
		return reportError(stderr, "resolve", err)
	}

	tw := tabwriter.NewWriter(stdout, 2, 4, 2, ' ', 0)
	for _, r := range resolved {
		if err := writeErrf(tw, "%s\n\n", r.Instrumentation); err != nil {
			return 1
		}
		if err := writeErrf(tw, "  %-48s %-24s %s\n", "PATH", "VALUE", "SOURCE"); err != nil {
			return 1
		}
		for _, o := range r.Options {
			src := o.Source
			if o.Source == config.SourceEnv {
				src = "env:" + o.EnvVar
			}
			if err := writeErrf(tw, "  %-48s %-24s %s\n", o.Path, valueDisplay(o.Value), src); err != nil {
				return 1
			}
		}
		if r.NetHTTPClientConfig != nil {
			if err := writeErrf(tw, "\n  typed: %+v\n", *r.NetHTTPClientConfig); err != nil {
				return 1
			}
		}
		if err := writeErrf(tw, "\n"); err != nil {
			return 1
		}
	}
	if err := tw.Flush(); err != nil {
		return 1
	}
	return 0
}

func runExplain(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)
	fs.SetOutput(stderr)
	manifestDir := fs.String("manifest", "./manifest", "behavior manifests directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		if err := writeErrf(stderr, "explain: expected exactly one option name or path\n"); err != nil {
			return 1
		}
		return 2
	}

	manifests, ok := discoverManifests(*manifestDir, stderr)
	if !ok {
		return 1
	}
	query := fs.Arg(0)
	for _, m := range manifests {
		for _, o := range m.Options {
			if o.Name != query && o.DeclarativePath != query {
				continue
			}
			return printExplain(stdout, stderr, m, &o)
		}
	}
	if err := writeErrf(stderr, "explain: no option matches %q\n", query); err != nil {
		return 1
	}
	return 1
}

func printExplain(stdout, stderr io.Writer, m *manifest.Manifest, o *manifest.Option) int {
	env := "—"
	if o.EnvVar != "" {
		env = o.EnvVar
	}
	semantics := "—"
	if o.EnvVarSemantics != "" {
		semantics = o.EnvVarSemantics
	}
	upstream := "—"
	if o.UpstreamRef != "" {
		upstream = o.UpstreamRef
	}
	lines := []string{
		fmt.Sprintf("Option:          %s", o.Name),
		fmt.Sprintf("Instrumentation: %s", m.Instrumentation.Name),
		fmt.Sprintf("Path:            %s", o.DeclarativePath),
		fmt.Sprintf("Type:            %s", o.Type),
		fmt.Sprintf("Default:         %s", defaultDisplay(o)),
		fmt.Sprintf("Stability:       %s", o.Stability),
		fmt.Sprintf("Env var:         %s", env),
		fmt.Sprintf("Env semantics:   %s", semantics),
		fmt.Sprintf("Upstream ref:    %s", upstream),
		fmt.Sprintf("Description:     %s", o.Description),
	}
	for _, l := range lines {
		if err := writeErrf(stdout, "%s\n", l); err != nil {
			return 1
		}
	}
	return 0
}

func runCatalog(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("catalog", flag.ContinueOnError)
	fs.SetOutput(stderr)
	manifestDir := fs.String("manifest", "./manifest", "behavior manifests directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 1 {
		if err := writeErrf(stderr, "catalog: at most one instrumentation name\n"); err != nil {
			return 1
		}
		return 2
	}

	manifests, ok := discoverManifests(*manifestDir, stderr)
	if !ok {
		return 1
	}
	if fs.NArg() == 1 {
		filter := fs.Arg(0)
		var filtered []*manifest.Manifest
		for _, m := range manifests {
			if m.Instrumentation.Name == filter {
				filtered = append(filtered, m)
			}
		}
		if len(filtered) == 0 {
			if err := writeErrf(stderr, "catalog: no manifest for instrumentation %q\n", filter); err != nil {
				return 1
			}
			return 1
		}
		manifests = filtered
	}

	tw := tabwriter.NewWriter(stdout, 2, 4, 2, ' ', 0)
	for _, m := range manifests {
		if err := writeErrf(tw, "# %s\n\n", m.Instrumentation.Name); err != nil {
			return 1
		}
		if err := writeErrf(tw, "  %-24s %-52s %-12s %-10s %-12s %s\n",
			"OPTION", "DECLARATIVE PATH", "TYPE", "DEFAULT", "STABILITY", "ENV VAR"); err != nil {
			return 1
		}
		for _, o := range m.Options {
			env := "—"
			if o.EnvVar != "" {
				env = o.EnvVar
			}
			if err := writeErrf(tw, "  %-24s %-52s %-12s %-10s %-12s %s\n",
				o.Name, o.DeclarativePath, o.Type, defaultDisplay(&o), o.Stability, env); err != nil {
				return 1
			}
		}
		if err := writeErrf(tw, "\n"); err != nil {
			return 1
		}
	}
	if err := tw.Flush(); err != nil {
		return 1
	}
	return 0
}

func discoverManifests(dir string, stderr io.Writer) ([]*manifest.Manifest, bool) {
	manifests, err := manifest.Discover(dir)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return nil, false
	}
	if len(manifests) == 0 {
		_, _ = fmt.Fprintf(stderr, "no manifests found under %s\n", dir)
		return nil, false
	}
	return manifests, true
}

func reportError(stderr io.Writer, cmd string, err error) int {
	if writeErr := writeErrf(stderr, "%s: %v\n", cmd, err); writeErr != nil {
		return 1
	}
	return 1
}

func defaultDisplay(o *manifest.Option) string {
	v, err := o.DefaultValue()
	if err != nil {
		return ""
	}
	switch manifest.OptionType(o.Type) {
	case manifest.TypeBoolean:
		b, _ := v.(bool)
		return strconv.FormatBool(b)
	case manifest.TypeListString:
		items, _ := v.([]string)
		if len(items) == 0 {
			return "[]"
		}
		quoted := make([]string, 0, len(items))
		for _, it := range items {
			quoted = append(quoted, strconv.Quote(it))
		}
		return "[" + strings.Join(quoted, ", ") + "]"
	default:
		return ""
	}
}

func valueDisplay(v any) string {
	switch x := v.(type) {
	case []string:
		return "[" + strings.Join(x, ", ") + "]"
	default:
		return fmt.Sprintf("%v", v)
	}
}
