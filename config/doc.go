// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Package config loads OpenTelemetry-style declarative configuration,
// performs ${ENV:-default} substitution, validates against generated JSON
// Schema, and resolves final values with precedence:
//
//	environment variables  >  configuration file  >  defaults
//
// Compatibility with otelc's existing controls is preserved:
//
//	OTEL_GO_ENABLED_INSTRUMENTATIONS
//	OTEL_GO_DISABLED_INSTRUMENTATIONS
//
// Phase 2: loader, substitution, validation, and resolver.
// Phase 3: bake-friendly resolved runtime package helpers.
//
// Hooks must consume typed accessors from this package (or generated runtime
// packages). They must never parse YAML directly.
package config
