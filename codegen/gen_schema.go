// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

func genSchema(m *manifest.Manifest) (string, []byte, error) {
	properties := map[string]any{}
	for _, o := range m.Options {
		insertSchemaPath(properties, &o)
	}

	doc := map[string]any{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"$id":         "https://github.com/ADITYA-CODE-SOURCE/otelcconfig/generated/schema/" + m.InstrumentationFileName() + ".json",
		"title":       m.Instrumentation.Name + " behavior configuration",
		"description": m.Instrumentation.Description,
		"type":        "object",
		"properties":  properties,
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("marshal generated schema for %s: %w", m.Instrumentation.Name, err)
	}
	data = append(data, '\n')
	return "schema/" + m.InstrumentationFileName() + ".json", data, nil
}

// insertSchemaPath builds nested instrumentation/development nodes for an
// option's dot-separated declarative path. General.* keys are official
// OpenTelemetry declarative configuration keys; go.* keys are prototype-owned.
func insertSchemaPath(properties map[string]any, o *manifest.Option) {
	segments := strings.Split(o.DeclarativePath, ".")
	cur := properties
	for i, seg := range segments {
		last := i == len(segments)-1
		if last {
			cur[seg] = schemaLeaf(o)
			return
		}
		node, ok := cur[seg].(map[string]any)
		if !ok {
			node = map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}
			cur[seg] = node
		}
		cur = node["properties"].(map[string]any)
	}
}

func schemaLeaf(o *manifest.Option) map[string]any {
	leaf := map[string]any{
		"description": o.Description,
		"x-stability": o.Stability,
	}
	if o.UpstreamRef != "" {
		leaf["x-upstream-ref"] = o.UpstreamRef
	}
	switch manifest.OptionType(o.Type) {
	case manifest.TypeBoolean:
		leaf["type"] = "boolean"
		if v, err := o.DefaultValue(); err == nil {
			leaf["default"] = v
		}
	case manifest.TypeListString:
		leaf["type"] = "array"
		leaf["items"] = map[string]any{"type": "string"}
		// Upstream general.http.client.request_captured_headers declares
		// minItems: 1. List defaults are intentionally not embedded so the
		// schema stays consistent with that constraint.
		leaf["minItems"] = 1
	}
	return leaf
}
