// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Package config will load OpenTelemetry-style declarative configuration,
// perform ${ENV:-default} substitution, validate against generated JSON
// Schema, and resolve final values with precedence:
//
//	environment variables  >  configuration file  >  defaults
//
// Compatibility with otelc's existing controls is required:
//
//	OTEL_GO_ENABLED_INSTRUMENTATIONS
//	OTEL_GO_DISABLED_INSTRUMENTATIONS
//
// Phase 0: package stub only.
// Phase 2: loader, substitution, resolver, and typed runtime accessors.
// Phase 3: bake-friendly resolved runtime package helpers.
//
// Hooks must consume typed accessors from this package (or generated runtime
// packages). They must never parse YAML directly.
package config
