# RFC 0001: Declarative Instrumentation Configuration for otelc

- **Status:** Draft (mechanism demonstrated in the `otelcconfig` prototype)
- **Date:** 2026-08-19
- **Related:** [otelc#705](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/issues/705),
  [opentelemetry-configuration](https://github.com/open-telemetry/opentelemetry-configuration),
  [ADR-0002](../adr/0002-adopt-otel-declarative-config-node.md),
  [ADR-0003](../adr/0003-bake-at-build-time.md)

> This is an independent prototype document exploring otelc#705. It is not an
> official OpenTelemetry proposal.

## Problem

otelc can enable or disable instrumentations via
`OTEL_GO_ENABLED_INSTRUMENTATIONS` / `OTEL_GO_DISABLED_INSTRUMENTATIONS`, but there
is no declarative way to configure *how* an instrumentation behaves: which HTTP
headers to capture, which URL query parameters to redact, or which semantic
convention generation to emit. Issue #705 asks for the OpenTelemetry declarative
configuration `instrumentation/development` node plus a Java-agent-style
per-instrumentation manifest.

## Proposed design

### 1. Behavior manifest (single source of truth)

One `metadata.yaml` per instrumentation declares its configurable options:
the declarative path (`general.*` official, `go.*` prototype-owned), type,
default, compatibility env var, stability, and documentation. Everything a user
can configure about an instrumentation is derived from this manifest.

**Generated from each manifest (build time, deterministic):**

- typed Go config struct with defaults
- env-var compatibility mapping
- JSON Schema fragment for the `instrumentation/development` shape
- human-readable option catalog

### 2. Validation and resolution (build time)

`otelcconfig` loads the user's declarative YAML, substitutes
`${ENV:-default}` references, validates it strictly against the generated schema
(unknown keys outside the open `go:` map are rejected), and resolves the final
engine values with precedence `env  >  file  >  defaults`. An unresolved
environment reference fails the build instead of failing silently at runtime.

### 3. Bake (freeze resolved values into the binary)

The resolved config is frozen into a generated Go package at build time. A
`runtime` package exposes typed accessors; instrumentation hooks import the
runtime package and call those accessors. Hooks never:

- open or parse YAML at runtime
- call `os.Getenv` for option values
- read stringly-typed undeclared paths

This matches the Issue #705 design note: "otelc parses and validates the config
at build time and bakes the resolved values into the instrumented binary. The
hooks never touch YAML."

### 4. Runtime consumption (enabler pattern)

Hooks mirror otelc's existing enabler pattern (see otelc
`instrumentation/net/http/client/client_hook.go`): an `Enable()` method reads the
frozen value (e.g. `runtime.NetHTTPClient().Enabled`) instead of a global toggle,
and request/response hooks read the remaining options.

## Alternatives considered

- **Environment variables only** (rejected) — cannot express list options (197
  headers, redaction lists, semconv migration) and gives no validation or
  documentation surface.
- **Runtime YAML/JSON parsing by hooks** (rejected) — contradicts #705's baked
  design note; adds latency, error paths, and a parser dependency to every binary.
- **Embedded JSON + `go:embed` read at init** (considered) — works, but a literal
  Go composite value is statically analyzable, needs no unmarshalling, and can be
  checked by the implemented `guard` analyzer.
- **Tool-file selection mixing selection and behavior** (rejected) — otelc ADR-0005
  owns selection (what gets woven); this mechanism owns behavior (how it behaves).
  The two must not be merged.

## Open questions

1. **Where does `bake` run inside otelc?** A standalone command (this prototype),
   a `-toolexec` stage, or part of the existing selection toolchain? The prototype
   deliberately stays standalone.
2. **Versioning and schema evolution** — how do generated types and defaults
   migrate when the manifest adds or changes options (semver for manifests)?
3. **`.otel.yml` integration** — if otelc already parses a per-module tool file,
   can the declarative section live there, or should it stay a separate file?
4. **Env precedence at bake time** — freezing env overrides into the binary is the
   demonstrated mechanism; otelc must decide whether env should still win at
   runtime (which the frozen model cannot, by construction).

## Out of scope

- Weaver emission registry (`schemas/otelc/groups/*.yaml`) — telemetry schema vs.
  behavior configuration stays separate.
- Selection mechanisms. See [ADR-0002](../adr/0002-adopt-otel-declarative-config-node.md).