// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/codegen"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

// Validator validates instrumentation/development content against a generated
// JSON Schema fragment.
type Validator struct {
	schema *jsonschema.Schema
}

// NewValidator compiles a generated JSON Schema fragment (the decoded
// development-node document produced by the codegen package).
func NewValidator(schemaDoc any) (*Validator, error) {
	const loc = "file:///otelcconfig/generated/schema/development.json"
	c := jsonschema.NewCompiler()
	if err := c.AddResource(loc, schemaDoc); err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	sch, err := c.Compile(loc)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return &Validator{schema: sch}, nil
}

// Validate reports whether the instrumentation/development content satisfies
// the generated schema. The general: node is closed (undeclared keys are
// rejected); the go: node is open.
func (v *Validator) Validate(development any) error {
	if err := v.schema.Validate(development); err != nil {
		return formatValidationError(err)
	}
	return nil
}

// Validators builds a JSON Schema validator for each instrumentation manifest
// from the deterministic codegen output, keyed by instrumentation name.
func Validators(manifests []*manifest.Manifest) (map[string]*Validator, error) {
	outs, err := codegen.GenerateAll(manifests)
	if err != nil {
		return nil, err
	}
	vs := make(map[string]*Validator, len(manifests))
	for _, m := range manifests {
		schemaBytes, ok := outs["schema/"+m.InstrumentationFileName()+".json"]
		if !ok {
			return nil, fmt.Errorf("no generated schema for %s", m.Instrumentation.Name)
		}
		var doc any
		if err := json.Unmarshal(schemaBytes, &doc); err != nil {
			return nil, fmt.Errorf("decode schema for %s: %w", m.Instrumentation.Name, err)
		}
		v, err := NewValidator(doc)
		if err != nil {
			return nil, fmt.Errorf("schema for %s: %w", m.Instrumentation.Name, err)
		}
		vs[m.Instrumentation.Name] = v
	}
	return vs, nil
}

func formatValidationError(err error) error {
	var ve *jsonschema.ValidationError
	if !errors.As(err, &ve) {
		return err
	}
	var msgs []string
	var walk func(*jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) == 0 {
			msgs = append(msgs, fmt.Sprintf("%s: %s", instancePath(e), describeKeyword(e)))
			return
		}
		for _, c := range e.Causes {
			walk(c)
		}
	}
	walk(ve)
	if len(msgs) == 0 {
		msgs = append(msgs, ve.Error())
	}
	return fmt.Errorf("validation failed:\n  %s", strings.Join(msgs, "\n  "))
}

func instancePath(e *jsonschema.ValidationError) string {
	var parts []string
	for _, p := range e.InstanceLocation {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return "instrumentation.development"
	}
	return strings.Join(parts, ".")
}

func describeKeyword(e *jsonschema.ValidationError) string {
	for _, k := range e.ErrorKind.KeywordPath() {
		switch k {
		case "additionalProperties":
			return "unexpected key (undeclared option)"
		case "type":
			return "wrong type"
		case "required":
			return "missing required key"
		case "minItems":
			return "too few items"
		case "items":
			return "invalid list item"
		}
	}
	kw := e.ErrorKind.KeywordPath()
	if len(kw) > 0 {
		return strings.Join(kw, ".")
	}
	return "invalid value"
}
