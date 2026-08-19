// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Document is a parsed declarative configuration file.
type Document struct {
	// Root is the full parsed YAML document.
	Root any
	// Development is the content of instrumentation.development.
	Development any
}

// Parse parses declarative configuration YAML. The instrumentation.development
// node must be present; other top-level sections are tolerated.
func Parse(data []byte) (*Document, error) {
	var root any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse configuration: %w", err)
	}
	doc := &Document{Root: root}
	if err := doc.extractDevelopment(); err != nil {
		return nil, err
	}
	return doc, nil
}

func (d *Document) extractDevelopment() error {
	root, ok := d.Root.(map[string]any)
	if !ok {
		return fmt.Errorf("configuration must be a YAML mapping")
	}
	inst, ok := root["instrumentation"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing top-level \"instrumentation\" section")
	}
	dev, ok := inst["development"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing \"instrumentation.development\" section")
	}
	d.Development = dev
	return nil
}

// NodeAtPath returns the value at a dot-separated path within a YAML tree.
func NodeAtPath(v any, path string) (any, bool) {
	cur := v
	for _, seg := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[seg]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}
