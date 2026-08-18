// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/manifest"
)

const manifestPath = "../manifest/nethttp/metadata.yaml"

func loadNetHTTP(t *testing.T) *manifest.Manifest {
	t.Helper()
	m, err := manifest.LoadFile(manifestPath)
	if err != nil {
		t.Fatalf("LoadFile(%s) error: %v", manifestPath, err)
	}
	return m
}

func TestGeneratedOutputsAreStable(t *testing.T) {
	m := loadNetHTTP(t)
	first, err := Generate(m)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	second, err := Generate(m)
	if err != nil {
		t.Fatalf("Generate() second run error: %v", err)
	}

	if len(first) != len(second) {
		t.Fatalf("first output paths = %d, second = %d", len(first), len(second))
	}
	for path := range first {
		want, ok := second[path]
		if !ok {
			t.Fatalf("second output missing path %q", path)
		}
		if !bytes.Equal(first[path], want) {
			t.Errorf("output for %q is not deterministic", path)
		}
	}
}

func TestGeneratedOutputsMatchCommitted(t *testing.T) {
	m := loadNetHTTP(t)
	outs, err := Generate(m)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if len(outs) != 5 {
		t.Fatalf("generated files = %d, want 5", len(outs))
	}
	for path, data := range outs {
		full := filepath.Join("../generated", filepath.FromSlash(path))
		want, err := os.ReadFile(full)
		if err != nil {
			t.Errorf("read committed %s: %v (run `otelcconfig generate`)", full, err)
			continue
		}
		if !bytes.Equal(want, data) {
			t.Errorf("generated output for %s drifted from committed file (run `otelcconfig generate`)", path)
		}
	}
}

func TestGeneratedTypesContainUpstreamPaths(t *testing.T) {
	m := loadNetHTTP(t)
	outs, err := Generate(m)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	typesOut := outs["types/nethttp_client_gen.go"]
	for _, path := range []string{
		"go.nethttp.client.enabled",
		"general.http.client.request_captured_headers",
		"general.sanitization.url.sensitive_query_parameters",
	} {
		if !bytes.Contains(typesOut, []byte(path)) {
			t.Errorf("generated types missing option path %q", path)
		}
	}
}
