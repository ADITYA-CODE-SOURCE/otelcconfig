// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Package manifest will parse and validate per-instrumentation behavior
// manifests (Java agent metadata.yaml style).
//
// Phase 0: package stub only.
// Phase 1: typed schema, net/http metadata.yaml, and loader tests.
//
// Behavior manifests are the single source of truth for configurable options.
// They are distinct from otelc's Weaver emission registry
// (schemas/otelc/groups/*.yaml), which describes emitted telemetry rather than
// runtime behavior controls. See docs/architecture.md and ADR-0002.
package manifest
