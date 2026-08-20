// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGuardCleanRepoDirs(t *testing.T) {
	code, stdout, stderr := runCmd(t, "guard", "../../demo", "../../runtime", "../../baked")
	if code != 0 {
		t.Fatalf("guard exit = %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "no undeclared configuration access") {
		t.Errorf("guard stdout = %q", stdout)
	}
}

func TestRunGuardFlagsViolations(t *testing.T) {
	dir := t.TempDir()
	writeViolatingHook(t, dir)
	code, stdout, stderr := runCmd(t, "guard", dir)
	if code != 1 {
		t.Fatalf("guard exit = %d, want 1 (stderr=%s)", code, stderr)
	}
	if !strings.Contains(stdout, "runtime.Register is reserved") {
		t.Errorf("guard stdout missing register diagnostic:\n%s", stdout)
	}
	if !strings.Contains(stdout, "must come from the baked runtime configuration") {
		t.Errorf("guard stdout missing env diagnostic:\n%s", stdout)
	}
	if !strings.Contains(stdout, "must not parse YAML at runtime") {
		t.Errorf("guard stdout missing yaml diagnostic:\n%s", stdout)
	}
	if !strings.Contains(stdout, "violation(s)") {
		t.Errorf("guard stdout missing summary = %q", stdout)
	}
}

func TestRunGuardMissingDir(t *testing.T) {
	code, _, stderr := runCmd(t, "guard", filepath.Join(t.TempDir(), "nope"))
	if code != 1 {
		t.Fatalf("guard exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "read directory") {
		t.Errorf("stderr = %q", stderr)
	}
}

func writeViolatingHook(t *testing.T, dir string) {
	t.Helper()
	content := `package hook

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/runtime"
)

func init() {
	runtime.Register(runtime.ConfigSnapshot{})
}

func F() string {
	_ = os.Getenv("OTEL_GO_ENABLED_INSTRUMENTATIONS")
	_ = yaml.Unmarshal
	return ""
}
`
	if err := os.WriteFile(filepath.Join(dir, "hook.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunDiffIdentical(t *testing.T) {
	code, stdout, stderr := runCmd(t, "diff", "--manifest", "../../manifest",
		"../../examples/minimal.yaml", "../../examples/minimal.yaml")
	if code != 0 {
		t.Fatalf("diff exit = %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "identical") {
		t.Errorf("diff stdout = %q", stdout)
	}
}

func TestRunDiffDifferences(t *testing.T) {
	code, stdout, stderr := runCmd(t, "diff", "--manifest", "../../manifest",
		"../../examples/minimal.yaml", "../../examples/nethttp.yaml")
	if code != 1 {
		t.Fatalf("diff exit = %d, want 1 (stderr=%s)", code, stderr)
	}
	for _, want := range []string{
		"! go.nethttp.client.enabled",
		"! general.http.client.request_captured_headers",
		"! general.sanitization.url.sensitive_query_parameters",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("diff stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestRunDiffUsage(t *testing.T) {
	code, _, stderr := runCmd(t, "diff")
	if code != 2 {
		t.Fatalf("diff with no args exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "expected exactly two configuration files") {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestRunDiffInvalidFile(t *testing.T) {
	code, _, stderr := runCmd(t, "diff", "--manifest", "../../manifest",
		filepath.Join(t.TempDir(), "nope.yaml"), "../../examples/minimal.yaml")
	if code != 1 {
		t.Fatalf("diff exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "diff:") {
		t.Errorf("stderr = %q", stderr)
	}
}
