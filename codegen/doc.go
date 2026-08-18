// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Package codegen generates deterministic artifacts from behavior manifests:
//
//   - typed Go configuration structs and option-path constants
//   - default-value constructors
//   - environment-variable compatibility mappings
//   - JSON Schema fragments for the instrumentation/development shape
//   - Markdown option catalogs
//
// Generated output is byte-stable for identical inputs so CI can fail on drift
// via `otelcconfig generate --check`. Behavior manifests are the single source
// of truth; generated files must never be hand-edited.
package codegen
