// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestEndToEnd builds and runs the demo binary and asserts it behaves exactly
// as the baked examples/demo.yaml configuration prescribes: the client is
// enabled, both configured headers are captured, and the sensitive query
// parameter value is redacted.
func TestEndToEnd(t *testing.T) {
	repo := filepath.Clean(filepath.Join("..", ".."))
	if _, err := os.Stat(filepath.Join(repo, "baked", "nethttp_client_gen.go")); err != nil {
		t.Skipf("baked package not present: %v", err)
	}

	bin := filepath.Join(t.TempDir(), "demo")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build demo: %v\n%s", err, out)
	}

	run := exec.Command(bin)
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("run demo: %v\n%s", err, out)
	}
	stdout := string(out)

	for _, want := range []string{
		"enabled=true",
		"method=GET",
		"url=",
		"token=%5BREDACTED%5D",
		"keep=visible",
		"header user-agent=otelcconfig-demo/1.0",
		"header x-request-id=demo-request-42",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("demo output missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "token=supersecret") {
		t.Errorf("demo output leaked sensitive token:\n%s", stdout)
	}
	if strings.Contains(stdout, "enabled=false") {
		t.Errorf("demo output reported disabled:\n%s", stdout)
	}
}
