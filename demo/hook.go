// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Package demo is an end-to-end proof that instrumentation hooks consume typed,
// baked runtime configuration and never parse YAML.
//
// It models the net/http client hook on otelc's
// instrumentation/net/http/client/client_hook.go enabler pattern, without the
// OpenTelemetry SDK: a minimal span observation records the configured behavior
// so tests can assert it. The binary under cmd/demo imports the generated baked
// package and demonstrates the full supply chain: declarative YAML -> bake ->
// runtime -> hook.
//
// This package is a proof of mechanism. It is not part of otelc and does not
// integrate with otelc's -toolexec pipeline.
package demo

import (
	"net/http"
	"net/url"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/runtime"
)

// Observation describes what the nethttp.client hook recorded for one request,
// mirroring the span attributes a real hook would emit.
type Observation struct {
	// Method is the HTTP method of the observed request.
	Method string
	// URL is the request URL with sensitive query parameter values redacted.
	URL string
	// CapturedHeaders maps each configured captured header name to its values.
	CapturedHeaders map[string][]string
}

// Observe receives one Observation per enabled, observed request.
type Observe func(Observation)

// Enabler controls whether the nethttp.client instrumentation emits telemetry,
// mirroring otelc's client enabler pattern. It reads the baked runtime
// configuration rather than a global toggle.
type Enabler struct{}

// Enable reports whether the nethttp.client instrumentation is enabled.
func (Enabler) Enable() bool {
	return runtime.NetHTTPClient().Enabled
}

// Client wraps base so that requests are observed according to the baked
// configuration. A nil base uses http.DefaultTransport; a nil observe discards
// observations.
func Client(base http.RoundTripper, observe Observe) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if observe == nil {
		observe = func(Observation) {}
	}
	return &client{base: base, observe: observe}
}

type client struct {
	base    http.RoundTripper
	observe Observe
}

func (c *client) RoundTrip(req *http.Request) (*http.Response, error) {
	cfg := runtime.NetHTTPClient()
	if !cfg.Enabled {
		return c.base.RoundTrip(req)
	}
	obs := Observation{
		Method:          req.Method,
		URL:             redactedURL(req.URL, cfg.SensitiveQueryParameters),
		CapturedHeaders: capturedHeaders(req.Header, cfg.RequestCapturedHeaders),
	}
	resp, err := c.base.RoundTrip(req)
	c.observe(obs)
	return resp, err
}

// capturedHeaders copies the values of the configured header names.
func capturedHeaders(h http.Header, names []string) map[string][]string {
	out := make(map[string][]string, len(names))
	for _, name := range names {
		canonical := http.CanonicalHeaderKey(name)
		if vals, ok := h[canonical]; ok {
			out[name] = append([]string(nil), vals...)
		}
	}
	return out
}

// redactedURL returns the URL string with the value of every sensitive query
// parameter replaced by [REDACTED]. The original URL is never mutated.
func redactedURL(u *url.URL, sensitive []string) string {
	if len(sensitive) == 0 {
		return u.String()
	}
	clone := *u
	query := clone.Query()
	changed := false
	for _, name := range sensitive {
		if _, ok := query[name]; ok {
			query.Set(name, "[REDACTED]")
			changed = true
		}
	}
	if !changed {
		return u.String()
	}
	clone.RawQuery = query.Encode()
	return clone.String()
}
