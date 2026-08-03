// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestRunVersion(t *testing.T) {
	if code := run([]string{"version"}); code != 0 {
		t.Fatalf("version exit code = %d, want 0", code)
	}
}

func TestRunHelp(t *testing.T) {
	if code := run([]string{"help"}); code != 0 {
		t.Fatalf("help exit code = %d, want 0", code)
	}
}

func TestRunEmptyArgs(t *testing.T) {
	if code := run(nil); code != 0 {
		t.Fatalf("empty args exit code = %d, want 0", code)
	}
}

func TestRunUnknown(t *testing.T) {
	if code := run([]string{"nope"}); code != 2 {
		t.Fatalf("unknown exit code = %d, want 2", code)
	}
}

func TestRunNotImplemented(t *testing.T) {
	for _, cmd := range []string{"generate", "validate", "resolve", "explain", "catalog", "diff", "bake", "guard"} {
		if code := run([]string{cmd}); code != 1 {
			t.Fatalf("%s exit code = %d, want 1", cmd, code)
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
