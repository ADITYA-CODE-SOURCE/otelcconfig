// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGenerateWritesArtifacts(t *testing.T) {
	tmp := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runGenerate([]string{"--manifest", "../../manifest", "--output", tmp}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runGenerate exit code = %d, stderr=%s", code, stderr.String())
	}
	for _, rel := range []string{
		"types/nethttp_client_gen.go",
		"defaults/nethttp_client_gen.go",
		"envmap/nethttp_client_gen.go",
		"schema/nethttp_client.json",
		"catalog/nethttp_client.md",
	} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); err != nil {
			t.Errorf("missing generated artifact %s: %v", rel, err)
		}
	}
	if !strings.Contains(stdout.String(), "nethttp.client") {
		t.Errorf("stdout missing instrumentation name: %s", stdout.String())
	}
}

func TestRunGenerateCheckPassesOnCommitted(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runGenerate([]string{"--manifest", "../../manifest", "--output", "../../generated", "--check"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runGenerate --check on committed output = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "up to date") {
		t.Errorf("stdout missing 'up to date': %s", stdout.String())
	}
}

func TestRunGenerateCheckFailsOnDrift(t *testing.T) {
	tmp := t.TempDir()
	var stdout bytes.Buffer
	code := runGenerate([]string{"--manifest", "../../manifest", "--output", tmp}, &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("generate into temp failed: %d", code)
	}
	path := filepath.Join(tmp, "types", "nethttp_client_gen.go")
	if err := os.WriteFile(path, []byte("// tampered\n"), 0o644); err != nil {
		t.Fatalf("tamper: %v", err)
	}

	var stderr bytes.Buffer
	code = runGenerate([]string{"--manifest", "../../manifest", "--output", tmp, "--check"}, &bytes.Buffer{}, &stderr)
	if code != 1 {
		t.Fatalf("generate --check on drifted output = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "out of date") {
		t.Errorf("stderr missing drift message: %s", stderr.String())
	}
}

func TestRunGenerateMissingManifestDir(t *testing.T) {
	var stderr bytes.Buffer
	code := runGenerate([]string{"--manifest", "./does-not-exist"}, &bytes.Buffer{}, &stderr)
	if code != 1 {
		t.Fatalf("runGenerate missing manifest dir exit code = %d, want 1", code)
	}
	if stderr.String() == "" {
		t.Error("expected error on stderr")
	}
}

func TestRunGenerateUnexpectedArgs(t *testing.T) {
	code := runGenerate([]string{"extra"}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 2 {
		t.Fatalf("runGenerate unexpected args exit code = %d, want 2", code)
	}
}
