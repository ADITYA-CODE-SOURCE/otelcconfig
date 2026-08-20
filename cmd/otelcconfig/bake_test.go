// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBakeWritesAndChecks(t *testing.T) {
	cfg := filepath.Join("..", "..", "examples", "nethttp.yaml")
	dir := t.TempDir()

	code, stdout, stderr := runCmd(t, "bake", "--manifest", "../../manifest",
		"--output", dir, cfg)
	if code != 0 {
		t.Fatalf("bake exit = %d, stderr=%s", code, stderr)
	}
	for _, f := range []string{"nethttp_client_gen.go", "nethttp_client.json"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatalf("bake did not write %s: %v", f, err)
		}
	}
	if !strings.Contains(stdout, "baked 2 artifact(s)") {
		t.Errorf("bake stdout = %q", stdout)
	}

	code, stdout, stderr = runCmd(t, "bake", "--check", "--manifest", "../../manifest",
		"--output", dir, cfg)
	if code != 0 {
		t.Fatalf("bake --check exit = %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "up to date") {
		t.Errorf("bake --check stdout = %q", stdout)
	}
}

func TestRunBakeDetectsDrift(t *testing.T) {
	cfg := filepath.Join("..", "..", "examples", "nethttp.yaml")
	dir := t.TempDir()
	if code, _, stderr := runCmd(t, "bake", "--manifest", "../../manifest",
		"--output", dir, cfg); code != 0 {
		t.Fatalf("bake exit = %d, stderr=%s", code, stderr)
	}
	gen := filepath.Join(dir, "nethttp_client_gen.go")
	data, err := os.ReadFile(gen)
	if err != nil {
		t.Fatal(err)
	}
	// Emulate committed drift.
	if err := os.WriteFile(gen, []byte("// drifted\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runCmd(t, "bake", "--check", "--manifest", "../../manifest",
		"--output", dir, cfg)
	if code != 1 {
		t.Fatalf("bake --check with drift exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "out of date") {
		t.Errorf("bake --check stderr = %q", stderr)
	}
	_ = data
}

func TestRunBakeFreezesFileValues(t *testing.T) {
	cfg := filepath.Join("..", "..", "examples", "nethttp.yaml")
	dir := t.TempDir()
	if code, _, stderr := runCmd(t, "bake", "--manifest", "../../manifest",
		"--output", dir, cfg); code != 0 {
		t.Fatalf("bake exit = %d, stderr=%s", code, stderr)
	}
	src, err := os.ReadFile(filepath.Join(dir, "nethttp_client_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Enabled: false", `"user-agent"`, `"x-request-id"`, `"token"`, `"password"`} {
		if !strings.Contains(string(src), want) {
			t.Errorf("baked source missing %q:\n%s", want, src)
		}
	}
}

func TestRunBakeRejectsUnsetEnv(t *testing.T) {
	tmp := writeTempConfig(t, "instrumentation:\n  development:\n    general:\n      http:\n        client:\n          request_captured_headers: [\"${OTEL_UNSET_VAR_FOR_TEST}\"]\n")
	code, _, stderr := runCmd(t, "bake", "--manifest", "../../manifest",
		"--output", t.TempDir(), tmp)
	if code != 1 {
		t.Fatalf("bake exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "is not set") {
		t.Errorf("stderr = %q, want unset-var error", stderr)
	}
}

func TestRunBakeMissingArgs(t *testing.T) {
	code, _, stderr := runCmd(t, "bake")
	if code != 2 {
		t.Fatalf("bake with no args exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected exactly one configuration file") {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestCLIHelpListsBake(t *testing.T) {
	code, stdout, _ := runCmd(t, "help")
	if code != 0 {
		t.Fatalf("help exit = %d", code)
	}
	if !strings.Contains(stdout, "bake") {
		t.Errorf("help stdout missing bake:\n%s", stdout)
	}
}
