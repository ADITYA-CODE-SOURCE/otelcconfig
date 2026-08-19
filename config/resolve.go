// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/generated/types"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

// Sources a resolved option value can come from.
const (
	SourceDefault = "default"
	SourceFile    = "file"
	SourceEnv     = "env"
)

const (
	envEnabledList  = "OTEL_GO_ENABLED_INSTRUMENTATIONS"
	envDisabledList = "OTEL_GO_DISABLED_INSTRUMENTATIONS"
)

// ResolvedOption is a single resolved option value and its source.
type ResolvedOption struct {
	// Path is the declarative path of the option.
	Path string
	// Value is the resolved value.
	Value any
	// Source is SourceDefault, SourceFile, or SourceEnv.
	Source string
	// EnvVar is the compatibility environment variable, when Source is env.
	EnvVar string
}

// Resolved is the resolved configuration for one instrumentation.
type Resolved struct {
	// Instrumentation is the instrumentation name (for example nethttp.client).
	Instrumentation string
	// Options lists every option of the instrumentation in manifest order.
	Options []ResolvedOption
	// NetHTTPClientConfig is the typed value for the nethttp.client manifest.
	// It is nil for other instrumentations.
	NetHTTPClientConfig *types.NetHTTPClientConfig
}

// Resolve computes final option values for every manifest with precedence
// environment variables > configuration file > defaults, including otelc
// enable/disable compatibility for csv_contains options.
func Resolve(manifests []*manifest.Manifest, development any, lookup LookupEnv) ([]*Resolved, error) {
	out := make([]*Resolved, 0, len(manifests))
	for _, m := range manifests {
		r, err := resolveOne(m, development, lookup)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func resolveOne(m *manifest.Manifest, development any, lookup LookupEnv) (*Resolved, error) {
	r := &Resolved{Instrumentation: m.Instrumentation.Name}
	if m.Instrumentation.Name == "nethttp.client" {
		r.NetHTTPClientConfig = &types.NetHTTPClientConfig{}
	}
	for _, o := range m.Options {
		val, present := NodeAtPath(development, o.DeclarativePath)
		src := SourceDefault
		if !present {
			dv, err := o.DefaultValue()
			if err != nil {
				return nil, err
			}
			val = dv
		} else {
			cv, err := coerce(&o, val)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", o.DeclarativePath, err)
			}
			val = cv
			src = SourceFile
		}
		if o.EnvVar != "" && o.EnvVarSemantics == string(manifest.SemanticsCSVContains) {
			if otelcVal, overridden := otelcEnableOverride(m, lookup); overridden {
				val = otelcVal
				src = SourceEnv
			}
		}
		r.Options = append(r.Options, ResolvedOption{
			Path:   o.DeclarativePath,
			Value:  val,
			Source: src,
			EnvVar: o.EnvVar,
		})
		if r.NetHTTPClientConfig != nil {
			setField(r.NetHTTPClientConfig, o.GoName(), val)
		}
	}
	return r, nil
}

// coerce converts a YAML-decoded value to the option's Go type.
func coerce(o *manifest.Option, v any) (any, error) {
	switch manifest.OptionType(o.Type) {
	case manifest.TypeBoolean:
		b, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("expected boolean, got %T", v)
		}
		return b, nil
	case manifest.TypeListString:
		switch items := v.(type) {
		case []string:
			return items, nil
		case []any:
			out := make([]string, 0, len(items))
			for _, it := range items {
				s, ok := it.(string)
				if !ok {
					return nil, fmt.Errorf("expected list of strings, got %T item", it)
				}
				out = append(out, s)
			}
			return out, nil
		default:
			return nil, fmt.Errorf("expected list of strings, got %T", v)
		}
	default:
		return nil, fmt.Errorf("unsupported type %q", o.Type)
	}
}

// otelcEnableOverride returns the effective enabled state from otelc's
// enable/disable environment variables and whether either variable is set.
// An empty enabled list does not restrict; the disabled list always wins.
func otelcEnableOverride(m *manifest.Manifest, lookup LookupEnv) (bool, bool) {
	id := strings.SplitN(m.Instrumentation.Name, ".", 2)[0]
	enabledList, hasEnabled := lookup(envEnabledList)
	disabledList, hasDisabled := lookup(envDisabledList)
	if (!hasEnabled || enabledList == "") && (!hasDisabled || disabledList == "") {
		return false, false
	}
	enabled := true
	if hasEnabled && enabledList != "" {
		enabled = csvContains(enabledList, id)
	}
	if hasDisabled && csvContains(disabledList, id) {
		enabled = false
	}
	return enabled, true
}

func csvContains(list, name string) bool {
	for _, part := range strings.Split(list, ",") {
		if strings.TrimSpace(part) == name {
			return true
		}
	}
	return false
}

func setField(c *types.NetHTTPClientConfig, field string, val any) {
	rv := reflect.ValueOf(c).Elem().FieldByName(field)
	if !rv.IsValid() {
		return
	}
	rv.Set(reflect.ValueOf(val))
}
