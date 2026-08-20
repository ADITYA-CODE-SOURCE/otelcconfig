// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("test write failure")
}

type failOnWrite struct {
	call int
	fail int
}

func (w *failOnWrite) Write(p []byte) (int, error) {
	w.call++
	if w.call == w.fail {
		return 0, errors.New("test write failure")
	}
	return len(p), nil
}

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

func TestRunReportsOutputFailure(t *testing.T) {
	var stderr bytes.Buffer
	if code := run([]string{"version"}, failingWriter{}, &stderr); code != 1 {
		t.Fatalf("version with failing output exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "test write failure") {
		t.Fatalf("stderr = %q, want write error", stderr.String())
	}
}

func TestRunReportsHelpOutputFailure(t *testing.T) {
	var stderr bytes.Buffer
	if code := run([]string{"help"}, failingWriter{}, &stderr); code != 1 {
		t.Fatalf("help with failing output exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "test write failure") {
		t.Fatalf("stderr = %q, want write error", stderr.String())
	}
}

func TestRunHandlesUnknownCommandOutputFailure(t *testing.T) {
	if code := run([]string{"nope"}, &bytes.Buffer{}, failingWriter{}); code != 1 {
		t.Fatalf("unknown with failing stderr exit code = %d, want 1", code)
	}
}

func TestRunHandlesUnknownCommandUsageFailure(t *testing.T) {
	stderr := &failOnWrite{fail: 2}
	if code := run([]string{"nope"}, &bytes.Buffer{}, stderr); code != 1 {
		t.Fatalf("unknown with failing usage exit code = %d, want 1", code)
	}
}
