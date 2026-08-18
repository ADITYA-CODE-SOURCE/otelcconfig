// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"fmt"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

// Outputs maps a generated file path (relative to the generated/ root) to its
// exact byte content.
type Outputs map[string][]byte

// Generate produces every artifact for a single instrumentation manifest.
// The returned map is deterministic: identical manifests produce identical
// bytes for every path.
func Generate(m *manifest.Manifest) (Outputs, error) {
	outs := Outputs{}
	generators := []func(*manifest.Manifest) (string, []byte, error){
		genTypes,
		genDefaults,
		genEnvMap,
		genSchema,
		genDocs,
	}
	for _, gen := range generators {
		path, data, err := gen(m)
		if err != nil {
			return nil, fmt.Errorf("generate for %s: %w", m.Instrumentation.Name, err)
		}
		if _, exists := outs[path]; exists {
			return nil, fmt.Errorf("duplicate generated path %q", path)
		}
		outs[path] = data
	}
	return outs, nil
}

// GenerateAll produces artifacts for several manifests, failing on collisions.
func GenerateAll(manifests []*manifest.Manifest) (Outputs, error) {
	all := Outputs{}
	for _, m := range manifests {
		outs, err := Generate(m)
		if err != nil {
			return nil, err
		}
		for path, data := range outs {
			if _, exists := all[path]; exists {
				return nil, fmt.Errorf("generated path collision: %q", path)
			}
			all[path] = data
		}
	}
	return all, nil
}
