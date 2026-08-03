// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Package codegen will generate deterministic artifacts from behavior
// manifests:
//
//   - typed Go configuration structs
//   - default-value constructors
//   - environment-variable compatibility mappings
//   - JSON Schema fragments
//   - Markdown option catalogs
//
// Phase 0: package stub only.
// Phase 1: generators, golden-file tests, and go generate integration.
//
// Generated output must be byte-stable for identical inputs so CI can fail on
// drift via `generate --check`.
package codegen
