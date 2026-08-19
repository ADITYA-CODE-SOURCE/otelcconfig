// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/codegen"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

func testValidator(t *testing.T) *Validator {
	t.Helper()
	m, err := manifest.LoadFile("../manifest/nethttp/metadata.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	outs, err := codegen.Generate(m)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	var doc any
	if err := json.Unmarshal(outs["schema/nethttp_client.json"], &doc); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	v, err := NewValidator(doc)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	return v
}

func envMap(env map[string]string) LookupEnv {
	return func(k string) (string, bool) {
		v, ok := env[k]
		return v, ok
	}
}

func decode(t *testing.T, yml string) any {
	t.Helper()
	doc, err := Parse([]byte(yml))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return doc.Development
}

func TestParseRequiresDevelopment(t *testing.T) {
	for name, yml := range map[string]string{
		"empty":      ``,
		"no section": "instrumentation:\n  foo:\n    bar: 1\n",
		"no dev":     "service:\n  name: x\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(yml)); err == nil {
				t.Fatalf("Parse() = nil error, want error")
			}
		})
	}
}

func TestSubstitute(t *testing.T) {
	env := envMap(map[string]string{"HOST": "example.com", "EMPTY": ""})
	tree := map[string]any{
		"general": map[string]any{"s": "${HOST}"},
		"empty":   map[string]any{"e": "${EMPTY}"},
		"def":     "x-${MISSING:-fallback}",
		"list":    []any{"${HOST}/a", "static"},
		"num":     42,
	}
	got, err := Substitute(tree, env)
	if err != nil {
		t.Fatalf("Substitute: %v", err)
	}
	m := got.(map[string]any)
	if m["general"].(map[string]any)["s"] != "example.com" {
		t.Errorf("general.s = %v", m["general"])
	}
	if m["empty"].(map[string]any)["e"] != "" {
		t.Errorf("empty substitution got %q", m["empty"])
	}
	if m["def"] != "x-fallback" {
		t.Errorf("default substitution got %v", m["def"])
	}
	if got := m["list"].([]any); got[0] != "example.com/a" || got[1] != "static" {
		t.Errorf("list substitution got %v", got)
	}
}

func TestSubstituteMissingWithoutDefaultFails(t *testing.T) {
	_, err := Substitute(map[string]any{"a": "${MISSING}"}, envMap(nil))
	if err == nil || !strings.Contains(err.Error(), "MISSING is not set") {
		t.Fatalf("Substitute() error = %v, want missing-var error", err)
	}
}

func TestSubstituteLeavesUnrelatedStrings(t *testing.T) {
	in := map[string]any{"a": "plain string", "b": "${SET:-x}"}
	got, err := Substitute(in, envMap(nil))
	if err != nil {
		t.Fatalf("Substitute: %v", err)
	}
	if got.(map[string]any)["a"] != "plain string" {
		t.Errorf("plain string changed")
	}
}

func TestValidateAcceptsMinimal(t *testing.T) {
	v := testValidator(t)
	if err := v.Validate(decode(t, "instrumentation:\n  development:\n    go:\n      nethttp:\n        client:\n          enabled: true\n")); err != nil {
		t.Fatalf("minimal validate: %v", err)
	}
}

func TestValidateAcceptsGoAndGeneralValues(t *testing.T) {
	v := testValidator(t)
	dev := decode(t, `instrumentation:
  development:
    general:
      http:
        client:
          request_captured_headers: ["user-agent"]
      sanitization:
        url:
          sensitive_query_parameters: ["token"]
    go:
      nethttp:
        client:
          enabled: false
`)
	if err := v.Validate(dev); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestValidateAcceptsOpenGoKey(t *testing.T) {
	v := testValidator(t)
	dev := decode(t, "instrumentation:\n  development:\n    go:\n      futureinst:\n        future_option: 1\n")
	if err := v.Validate(dev); err != nil {
		t.Fatalf("open go: validate: %v", err)
	}
}

func TestValidateRejectsUndeclaredGeneralKey(t *testing.T) {
	v := testValidator(t)
	dev := decode(t, "instrumentation:\n  development:\n    general:\n      http:\n        server:\n          request_captured_headers: [\"x\"]\n")
	err := v.Validate(dev)
	if err == nil || !strings.Contains(err.Error(), "unexpected key") {
		t.Fatalf("Validate() error = %v, want unexpected-key error", err)
	}
	if !strings.Contains(err.Error(), "general.http") {
		t.Errorf("error should mention the path, got: %v", err)
	}
}

func TestValidateRejectsWrongType(t *testing.T) {
	v := testValidator(t)
	dev := decode(t, "instrumentation:\n  development:\n    go:\n      nethttp:\n        client:\n          enabled: \"yes\"\n")
	err := v.Validate(dev)
	if err == nil || !strings.Contains(err.Error(), "wrong type") {
		t.Fatalf("Validate() error = %v, want wrong-type error", err)
	}
}

func TestValidateRejectsWrongListItem(t *testing.T) {
	v := testValidator(t)
	dev := decode(t, "instrumentation:\n  development:\n    general:\n      http:\n        client:\n          request_captured_headers: [1, 2]\n")
	err := v.Validate(dev)
	if err == nil || !strings.Contains(err.Error(), "wrong type") {
		t.Fatalf("Validate() error = %v, want wrong-type error", err)
	}
}

func TestResolveDefaults(t *testing.T) {
	m, err := manifest.LoadFile("../manifest/nethttp/metadata.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	res, err := Resolve([]*manifest.Manifest{m}, map[string]any{}, envMap(nil))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("resolved = %d, want 1", len(res))
	}
	got := res[0]
	if got.NetHTTPClientConfig == nil {
		t.Fatal("typed config is nil")
	}
	if got.NetHTTPClientConfig.Enabled != true {
		t.Errorf("Enabled default = %v, want true", got.NetHTTPClientConfig.Enabled)
	}
	if len(got.NetHTTPClientConfig.RequestCapturedHeaders) != 0 {
		t.Errorf("RequestCapturedHeaders default = %v", got.NetHTTPClientConfig.RequestCapturedHeaders)
	}
	if got.Options[0].Source != SourceDefault {
		t.Errorf("enabled source = %q, want default", got.Options[0].Source)
	}
}

func TestResolveFileValues(t *testing.T) {
	m, err := manifest.LoadFile("../manifest/nethttp/metadata.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	dev := decode(t, `instrumentation:
  development:
    general:
      http:
        client:
          request_captured_headers: ["user-agent"]
    go:
      nethttp:
        client:
          enabled: false
`)
	res, err := Resolve([]*manifest.Manifest{m}, dev, envMap(nil))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	c := res[0].NetHTTPClientConfig
	if c.Enabled != false {
		t.Errorf("Enabled = %v, want false from file", c.Enabled)
	}
	if c.RequestCapturedHeaders[0] != "user-agent" {
		t.Errorf("headers = %v", c.RequestCapturedHeaders)
	}
	if res[0].Options[0].Source != SourceFile {
		t.Errorf("enabled source = %q, want file", res[0].Options[0].Source)
	}
	if res[0].Options[1].Source != SourceFile {
		t.Errorf("headers source = %q, want file", res[0].Options[1].Source)
	}
}

func TestResolveCSVCompat(t *testing.T) {
	m, err := manifest.LoadFile("../manifest/nethttp/metadata.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	dev := map[string]any{}
	tests := []struct {
		name    string
		env     map[string]string
		wantVal bool
		wantSrc string
	}{
		{"enabled list contains", map[string]string{"OTEL_GO_ENABLED_INSTRUMENTATIONS": "nethttp,foo"}, true, SourceEnv},
		{"enabled list excludes", map[string]string{"OTEL_GO_ENABLED_INSTRUMENTATIONS": "foo,bar"}, false, SourceEnv},
		{"disabled list contains", map[string]string{"OTEL_GO_DISABLED_INSTRUMENTATIONS": "nethttp"}, false, SourceEnv},
		{"disabled wins", map[string]string{
			"OTEL_GO_ENABLED_INSTRUMENTATIONS":  "nethttp",
			"OTEL_GO_DISABLED_INSTRUMENTATIONS": "nethttp",
		}, false, SourceEnv},
		{"unset", nil, true, SourceDefault},
		{"empty enabled list", map[string]string{"OTEL_GO_ENABLED_INSTRUMENTATIONS": ""}, true, SourceDefault},
		{"enabled excludes but empty disabled", map[string]string{
			"OTEL_GO_ENABLED_INSTRUMENTATIONS":  "foo",
			"OTEL_GO_DISABLED_INSTRUMENTATIONS": "",
		}, false, SourceEnv},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Resolve([]*manifest.Manifest{m}, dev, envMap(tt.env))
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			got := res[0].Options[0]
			if got.Value != tt.wantVal {
				t.Errorf("enabled = %v, want %v", got.Value, tt.wantVal)
			}
			if got.Source != tt.wantSrc {
				t.Errorf("source = %q, want %q", got.Source, tt.wantSrc)
			}
		})
	}
}

func TestNodeAtPath(t *testing.T) {
	tree := map[string]any{
		"general": map[string]any{
			"http": map[string]any{"client": map[string]any{"request_captured_headers": []any{"x"}}},
		},
	}
	if _, ok := NodeAtPath(tree, "general.http.client.request_captured_headers"); !ok {
		t.Error("expected value at known path")
	}
	if _, ok := NodeAtPath(tree, "general.http.missing"); ok {
		t.Error("unexpected value at unknown path")
	}
}
