// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Package demo will host an end-to-end net/http instrumentation demonstration
// that consumes typed runtime configuration produced by otelcconfig.
//
// Phase 0: package stub only.
// Phase 3: hook modelled on otelc's
// instrumentation/net/http/client/client_hook.go enabler pattern, with tests
// for header capture and query-parameter redaction.
//
// This package is a proof of mechanism. It is not part of otelc and does not
// integrate with otelc's -toolexec pipeline.
package demo
