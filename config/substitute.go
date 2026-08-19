// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"regexp"
	"strings"
)

// LookupEnv fetches an environment variable value. It returns false when the
// variable is not set.
type LookupEnv func(string) (string, bool)

var subPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:-[^}]*)?\}`)

// Substitute replaces ${ENV} and ${ENV:-default} references in every string
// scalar of the tree, following the OTel configuration ABNF. A reference
// without a default fails when the variable is unset. The returned tree is a
// copy; the input is not modified.
func Substitute(v any, lookup LookupEnv) (any, error) {
	return substitute(v, lookup)
}

func substitute(v any, lookup LookupEnv) (any, error) {
	switch x := v.(type) {
	case string:
		return substituteString(x, lookup)
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			nv, err := substitute(val, lookup)
			if err != nil {
				return nil, err
			}
			out[k] = nv
		}
		return out, nil
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			nv, err := substitute(val, lookup)
			if err != nil {
				return nil, err
			}
			out[i] = nv
		}
		return out, nil
	default:
		return v, nil
	}
}

func substituteString(s string, lookup LookupEnv) (string, error) {
	if !strings.Contains(s, "${") {
		return s, nil
	}
	var err error
	out := subPattern.ReplaceAllStringFunc(s, func(m string) string {
		if err != nil {
			return m
		}
		sub := subPattern.FindStringSubmatch(m)
		name := sub[1]
		if val, ok := lookup(name); ok {
			return val
		}
		if sub[2] != "" {
			return sub[2][2:] // strip the ":-" prefix
		}
		err = fmt.Errorf("environment variable %s is not set", name)
		return m
	})
	return out, err
}
