// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const validManifest = `
manifest_version: "0.1.0"

instrumentation:
  name: nethttp.client
  description: HTTP client instrumentation
  library_link: https://pkg.go.dev/net/http

options:
  - name: enabled
    declarative_path: go.nethttp.client.enabled
    type: boolean
    default: true
    env_var: OTEL_GO_ENABLED_INSTRUMENTATIONS
    env_var_semantics: csv_contains
    description: Whether the instrumentation is enabled.
    stability: stable
  - name: request_captured_headers
    declarative_path: general.http.client.request_captured_headers
    type: list_string
    default: []
    description: Headers to capture.
    stability: experimental
    upstream_ref: general.http.client.request_captured_headers
`

func TestLoadValidManifest(t *testing.T) {
	m, err := Load([]byte(validManifest))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if m.Instrumentation.Name != "nethttp.client" {
		t.Errorf("instrumentation name = %q, want nethttp.client", m.Instrumentation.Name)
	}
	if len(m.Options) != 2 {
		t.Fatalf("options = %d, want 2", len(m.Options))
	}
	v, err := m.Options[0].DefaultValue()
	if err != nil {
		t.Fatalf("DefaultValue(enabled) error: %v", err)
	}
	if v != true {
		t.Errorf("enabled default = %v, want true", v)
	}
}

func TestLoadRejectsInvalidManifests(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Manifest)
		wantErr string
	}{
		{
			name: "missing version",
			mutate: func(m *Manifest) {
				m.ManifestVersion = ""
			},
			wantErr: "manifest_version is required",
		},
		{
			name: "missing name",
			mutate: func(m *Manifest) {
				m.Instrumentation.Name = ""
			},
			wantErr: "instrumentation.name is required",
		},
		{
			name: "no options",
			mutate: func(m *Manifest) {
				m.Options = nil
			},
			wantErr: "declares no options",
		},
		{
			name: "bad type",
			mutate: func(m *Manifest) {
				m.Options[0].Type = "map"
			},
			wantErr: "unsupported type",
		},
		{
			name: "bad stability",
			mutate: func(m *Manifest) {
				m.Options[0].Stability = "development"
			},
			wantErr: "unsupported stability",
		},
		{
			name: "env var without semantics",
			mutate: func(m *Manifest) {
				m.Options[0].EnvVarSemantics = ""
			},
			wantErr: "env_var_semantics is required",
		},
		{
			name: "duplicate option name",
			mutate: func(m *Manifest) {
				m.Options[1].Name = m.Options[0].Name
			},
			wantErr: "more than once",
		},
		{
			name: "bad boolean default",
			mutate: func(m *Manifest) {
				m.Options[0].Default.SetString("not-a-bool")
			},
			wantErr: "default for boolean option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var base Manifest
			if err := yaml.Unmarshal([]byte(validManifest), &base); err != nil {
				t.Fatalf("setup: %v", err)
			}
			if tt.mutate != nil {
				tt.mutate(&base)
			}
			err := base.Validate()
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %q, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestDiscoverNetHTTP(t *testing.T) {
	manifests, err := Discover(".")
	if err != nil {
		t.Fatalf("Discover(.) error: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("Discover(.) found %d manifests, want 1", len(manifests))
	}
	if manifests[0].Instrumentation.Name != "nethttp.client" {
		t.Errorf("first manifest = %q, want nethttp.client", manifests[0].Instrumentation.Name)
	}
}

func TestLoadFileNetHTTP(t *testing.T) {
	p := filepath.Join("nethttp", "metadata.yaml")
	if _, err := os.Stat(p); err != nil {
		t.Skipf("manifest %s not present: %v", p, err)
	}
	m, err := LoadFile(p)
	if err != nil {
		t.Fatalf("LoadFile(%s) error: %v", p, err)
	}
	if m.StructName() != "NetHTTPClientConfig" {
		t.Errorf("StructName() = %q, want NetHTTPClientConfig", m.StructName())
	}
	if m.DefaultsFuncName() != "DefaultNetHTTPClientConfig" {
		t.Errorf("DefaultsFuncName() = %q, want DefaultNetHTTPClientConfig", m.DefaultsFuncName())
	}
	if m.InstrumentationFileName() != "nethttp_client" {
		t.Errorf("InstrumentationFileName() = %q, want nethttp_client", m.InstrumentationFileName())
	}
}

func TestGoIdent(t *testing.T) {
	tests := []struct{ in, want string }{
		{"enabled", "Enabled"},
		{"request_captured_headers", "RequestCapturedHeaders"},
		{"sensitive_query_parameters", "SensitiveQueryParameters"},
		{"nethttp", "NetHTTP"},
		{"semconv", "SemConv"},
	}
	for _, tt := range tests {
		if got := goIdent(tt.in); got != tt.want {
			t.Errorf("goIdent(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
