// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Package demo is an end-to-end proof that instrumentation hooks consume typed,
// baked runtime configuration and never parse YAML. See docs/rfc/0001 and
// docs/architecture.md for the mechanism.
//
// The hook models otelc's instrumentation/net/http/client/client_hook.go enabler
// pattern without the OpenTelemetry SDK: Client wraps an http.RoundTripper so
// requests are observed according to the baked configuration (captured headers,
// sensitive query redaction). cmd/demo imports the generated baked package and
// demonstrates the full supply chain.
//
// This package is a proof of mechanism. It is not part of otelc and does not
// integrate with otelc's -toolexec pipeline.
package demo
