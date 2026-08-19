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

func runCmd(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestRunValidateAcceptsExamples(t *testing.T) {
	for _, f := range []string{"minimal.yaml", "nethttp.yaml"} {
		path := filepath.Join("..", "..", "examples", f)
		if _, err := os.Stat(path); err != nil {
			t.Skipf("example %s not present", path)
		}
		code, stdout, stderr := runCmd(t, "validate", "--manifest", "../../manifest", path)
		if code != 0 {
			t.Fatalf("validate %s exit = %d, stderr=%s", f, code, stderr)
		}
		if !strings.Contains(stdout, "valid") {
			t.Errorf("validate %s stdout = %q", f, stdout)
		}
	}
}

func TestRunValidateRejectsUndeclaredGeneralKey(t *testing.T) {
	tmp := writeTempConfig(t, "instrumentation:\n  development:\n    general:\n      http:\n        server:\n          request_captured_headers: [\"x\"]\n")
	code, _, stderr := runCmd(t, "validate", "--manifest", "../../manifest", tmp)
	if code != 1 {
		t.Fatalf("validate exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "unexpected key") {
		t.Errorf("stderr = %q, want unexpected-key error", stderr)
	}
}

func TestRunValidateFailsOnMissingDevelopment(t *testing.T) {
	tmp := writeTempConfig(t, "instrumentation:\n  foo:\n    bar: 1\n")
	code, _, stderr := runCmd(t, "validate", "--manifest", "../../manifest", tmp)
	if code != 1 {
		t.Fatalf("validate exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "instrumentation.development") {
		t.Errorf("stderr = %q, want development error", stderr)
	}
}

func TestRunValidateFailsOnUnsetEnvReference(t *testing.T) {
	tmp := writeTempConfig(t, "instrumentation:\n  development:\n    general:\n      http:\n        client:\n          request_captured_headers: [\"${OTEL_UNSET_VAR_FOR_TEST}\"]\n")
	code, _, stderr := runCmd(t, "validate", "--manifest", "../../manifest", tmp)
	if code != 1 {
		t.Fatalf("validate exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "is not set") {
		t.Errorf("stderr = %q, want unset-var error", stderr)
	}
}

func TestRunResolveShowsSources(t *testing.T) {
	code, stdout, stderr := runCmd(t, "resolve", "--manifest", "../../manifest",
		filepath.Join("..", "..", "examples", "nethttp.yaml"))
	if code != 0 {
		t.Fatalf("resolve exit = %d, stderr=%s", code, stderr)
	}
	for _, want := range []string{
		"nethttp.client",
		"go.nethttp.client.enabled",
		"false",
		"file",
		"general.http.client.request_captured_headers",
		"user-agent",
		"typed: {Enabled:false",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("resolve stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestRunResolveYieldsEnvCompatSource(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "nethttp")
	code, stdout, stderr := runCmd(t, "resolve", "--manifest", "../../manifest",
		filepath.Join("..", "..", "examples", "minimal.yaml"))
	if code != 0 {
		t.Fatalf("resolve exit = %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "env:OTEL_GO_ENABLED_INSTRUMENTATIONS") {
		t.Errorf("resolve stdout missing env source:\n%s", stdout)
	}
	if !strings.Contains(stdout, "true") {
		t.Errorf("resolve stdout missing enabled value:\n%s", stdout)
	}
}

func TestRunExplain(t *testing.T) {
	code, stdout, stderr := runCmd(t, "explain", "--manifest", "../../manifest",
		"general.http.client.request_captured_headers")
	if code != 0 {
		t.Fatalf("explain exit = %d, stderr=%s", code, stderr)
	}
	for _, want := range []string{
		"request_captured_headers",
		"general.http.client.request_captured_headers",
		"list_string",
		"experimental",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("explain stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestRunExplainAcceptsShortName(t *testing.T) {
	code, stdout, _ := runCmd(t, "explain", "--manifest", "../../manifest", "enabled")
	if code != 0 {
		t.Fatalf("explain exit = %d", code)
	}
	if !strings.Contains(stdout, "go.nethttp.client.enabled") {
		t.Errorf("explain stdout missing path:\n%s", stdout)
	}
}

func TestRunExplainUnknownOption(t *testing.T) {
	code, _, stderr := runCmd(t, "explain", "--manifest", "../../manifest", "nope")
	if code != 1 {
		t.Fatalf("explain exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "no option matches") {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestRunCatalog(t *testing.T) {
	code, stdout, stderr := runCmd(t, "catalog", "--manifest", "../../manifest")
	if code != 0 {
		t.Fatalf("catalog exit = %d, stderr=%s", code, stderr)
	}
	for _, want := range []string{
		"nethttp.client",
		"enabled",
		"request_captured_headers",
		"sensitive_query_parameters",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("catalog stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestRunCatalogFiltered(t *testing.T) {
	code, stdout, stderr := runCmd(t, "catalog", "--manifest", "../../manifest", "nethttp.client")
	if code != 0 {
		t.Fatalf("catalog exit = %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "enabled") {
		t.Errorf("catalog stdout missing enabled:\n%s", stdout)
	}
}

func TestRunCatalogUnknownInstrumentation(t *testing.T) {
	code, _, stderr := runCmd(t, "catalog", "--manifest", "../../manifest", "grpc")
	if code != 1 {
		t.Fatalf("catalog exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "no manifest") {
		t.Errorf("stderr = %q", stderr)
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestCLIHelpListsConfigCommands(t *testing.T) {
	code, stdout, _ := runCmd(t, "help")
	if code != 0 {
		t.Fatalf("help exit = %d", code)
	}
	for _, want := range []string{"validate", "resolve", "explain", "catalog"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help stdout missing %q", want)
		}
	}
}
