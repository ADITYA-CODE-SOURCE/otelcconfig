// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

package demo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/generated/types"
	"github.com/ADITYA-CODE-SOURCE/otelcconfig/runtime"
)

// registerForTest installs a snapshot for the duration of the test.
func registerForTest(t *testing.T, cfg types.NetHTTPClientConfig) {
	t.Helper()
	runtime.Register(runtime.ConfigSnapshot{NetHTTPClient: cfg})
}

func newStubServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestClientCapturesHeadersAndRedactsURL(t *testing.T) {
	registerForTest(t, types.NetHTTPClientConfig{
		Enabled:                  true,
		RequestCapturedHeaders:   []string{"user-agent", "x-request-id"},
		SensitiveQueryParameters: []string{"token"},
	})
	srv := newStubServer(t)

	var got *Observation
	rt := Client(nil, func(o Observation) { got = &o })
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/path?token=secret&keep=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "demo/1.0")
	req.Header.Set("X-Request-Id", "abc-123")

	resp, err := (&http.Client{Transport: rt}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if got == nil {
		t.Fatal("no observation recorded")
	}
	if got.Method != http.MethodGet {
		t.Errorf("Method = %q", got.Method)
	}
	if got.CapturedHeaders["user-agent"][0] != "demo/1.0" {
		t.Errorf("CapturedHeaders user-agent = %v", got.CapturedHeaders["user-agent"])
	}
	if got.CapturedHeaders["x-request-id"][0] != "abc-123" {
		t.Errorf("CapturedHeaders x-request-id = %v", got.CapturedHeaders["x-request-id"])
	}
	if got.URL == "" || strings.Contains(got.URL, "token=secret") {
		t.Errorf("URL not redacted: %q", got.URL)
	}
	if !strings.Contains(got.URL, "keep=1") {
		t.Errorf("non-sensitive param lost: %q", got.URL)
	}
}

func TestClientDisabledEmitsNothing(t *testing.T) {
	registerForTest(t, types.NetHTTPClientConfig{Enabled: false})
	srv := newStubServer(t)

	calls := 0
	rt := Client(nil, func(Observation) { calls++ })
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/?token=secret", nil)
	resp, err := (&http.Client{Transport: rt}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if calls != 0 {
		t.Fatalf("observe called %d times, want 0", calls)
	}
}

func TestEnablerReadsBakedConfig(t *testing.T) {
	var enabler Enabler

	registerForTest(t, types.NetHTTPClientConfig{Enabled: false})
	if enabler.Enable() {
		t.Fatal("Enable = true, want false")
	}
	registerForTest(t, types.NetHTTPClientConfig{Enabled: true})
	if !enabler.Enable() {
		t.Fatal("Enable = false, want true")
	}
}

func TestRedactedURLUnchangedWithoutSensitiveParams(t *testing.T) {
	u, _ := url.Parse("http://example.com/path?keep=1")
	got := redactedURL(u, []string{"token"})
	if got != "http://example.com/path?keep=1" {
		t.Fatalf("redactedURL = %q, want unchanged", got)
	}
	if u.String() != "http://example.com/path?keep=1" {
		t.Fatalf("original URL mutated: %q", u.String())
	}
}

func TestRedactedURLNoSensitiveConfig(t *testing.T) {
	u, _ := url.Parse("http://example.com/?token=secret")
	if got := redactedURL(u, nil); got != "http://example.com/?token=secret" {
		t.Fatalf("redactedURL = %q, want unchanged", got)
	}
}
