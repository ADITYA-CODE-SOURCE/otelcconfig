// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"strings"
	"testing"
)

func runCommand(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestRunVersion(t *testing.T) {
	oldVersion := version
	version = "test-version"
	t.Cleanup(func() { version = oldVersion })

	code, stdout, stderr := runCommand("version")
	if code != 0 {
		t.Fatalf("version exit code = %d, want 0", code)
	}
	if stdout != "otelcconfig version test-version\n" {
		t.Fatalf("version stdout = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("version stderr = %q, want empty", stderr)
	}
}

func TestRunHelp(t *testing.T) {
	code, stdout, stderr := runCommand("help")
	if code != 0 {
		t.Fatalf("help exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "Declarative Configuration Toolkit") {
		t.Fatalf("help output missing description: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("help stderr = %q, want empty", stderr)
	}
}

func TestRunEmptyArgs(t *testing.T) {
	code, stdout, stderr := runCommand()
	if code != 0 {
		t.Fatalf("empty args exit code = %d, want 0", code)
	}
	if stdout == "" || stderr != "" {
		t.Fatalf("empty args stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestRunUnknown(t *testing.T) {
	code, stdout, stderr := runCommand("nope")
	if code != 2 {
		t.Fatalf("unknown exit code = %d, want 2", code)
	}
	if stdout != "" || !strings.Contains(stderr, `unknown command "nope"`) {
		t.Fatalf("unknown stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestRunNotImplemented(t *testing.T) {
	for _, cmd := range []string{"generate", "validate", "resolve", "explain", "catalog", "diff", "bake", "guard"} {
		code, stdout, stderr := runCommand(cmd)
		if code != 1 {
			t.Fatalf("%s exit code = %d, want 1", cmd, code)
		}
		if stdout != "" || !strings.Contains(stderr, "not implemented in Phase 0") {
			t.Fatalf("%s stdout=%q stderr=%q", cmd, stdout, stderr)
		}
	}
}

func TestPhaseFor(t *testing.T) {
	cases := map[string]string{
		"generate": "Phase 1 (v0.2.0)",
		"validate": "Phase 2 (v0.3.0)",
		"bake":     "Phase 3 (v0.4.0)",
		"guard":    "Phase 4 (v0.5.0)",
		"other":    "a later phase",
	}
	for cmd, want := range cases {
		if got := phaseFor(cmd); got != want {
			t.Errorf("phaseFor(%q) = %q, want %q", cmd, got, want)
		}
	}
}
