// Copyright The otelcconfig Authors
// SPDX-License-Identifier: Apache-2.0

// Package runtime exposes the baked, resolved configuration that otelc
// instrumentation hooks consume at runtime.
//
// Hooks never open or parse YAML at runtime. The resolved configuration is
// frozen into the instrumented binary at build time by `otelcconfig bake`,
// which emits a package (see bake/) that registers a ConfigSnapshot here during
// init. See docs/architecture.md for the runtime contract.
package runtime

import (
	"sync/atomic"

	"github.com/ADITYA-CODE-SOURCE/otelcconfig/generated/types"
)

// ConfigSnapshot holds the frozen resolved configuration for every
// instrumentation. One snapshot is registered per binary.
type ConfigSnapshot struct {
	// NetHTTPClient is the resolved nethttp.client behavior configuration.
	NetHTTPClient types.NetHTTPClientConfig
}

// snapshot is the registered snapshot. The zero value means none is registered,
// which happens when a binary was built without importing the baked package.
var snapshot atomic.Value

// lookupSnapshot returns the registered snapshot. It is a variable so tests can
// emulate a binary without any baked configuration.
var lookupSnapshot = func() (ConfigSnapshot, bool) {
	v, ok := snapshot.Load().(ConfigSnapshot)
	return v, ok
}

// Register installs the baked configuration snapshot. It must be called from
// the generated baked package's init function, once per process.
func Register(c ConfigSnapshot) {
	snapshot.Store(c)
}

// NetHTTPClient returns the resolved behavior configuration for the
// nethttp.client instrumentation. The returned value is a copy; mutation is
// safe. It panics when no configuration snapshot has been baked into the binary.
func NetHTTPClient() types.NetHTTPClientConfig {
	c, ok := lookupSnapshot()
	if !ok {
		panic("otelcconfig runtime: no configuration baked; import the generated baked package")
	}
	return copyConfig(c.NetHTTPClient)
}

// copyConfig returns a deep copy so hooks can never mutate the shared snapshot.
func copyConfig(c types.NetHTTPClientConfig) types.NetHTTPClientConfig {
	if c.RequestCapturedHeaders != nil {
		c.RequestCapturedHeaders = append([]string(nil), c.RequestCapturedHeaders...)
	}
	if c.SensitiveQueryParameters != nil {
		c.SensitiveQueryParameters = append([]string(nil), c.SensitiveQueryParameters...)
	}
	return c
}
