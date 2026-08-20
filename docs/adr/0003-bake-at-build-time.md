# ADR-0003: Bake Configuration at Build Time

- **Status:** Accepted
- **Date:** 2026-08-19
- **Deciders:** Project maintainers
- **Related:** [RFC 0001](../rfc/0001-declarative-instrumentation-configuration.md),
  [ADR-0002](0002-adopt-otel-declarative-config-node.md),
  [otelc#705](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/issues/705)

## Context

Issue #705 states that otelc should "parse and validate the config at build time
and bake the resolved values into the instrumented binary. The hooks never touch
YAML." This is the design note that ends the dependency on runtime configuration
parsing and on per-machine environment precedence.

Without baking, hooks would either (a) parse YAML or JSON at runtime, (b) call
`os.Getenv` per option, or (c) rely on a global mutable config set by some
initializer. All three reintroduce the problems the issue wants to remove:
runtime failure modes, parser dependencies, and values that change per-deployment
after the binary is built.

## Decision

Configuration is **frozen into the binary at build time**:

- `otelcconfig bake <file>` resolves the validated configuration (precedence
  `env > file > defaults`) and emits a Go package that registers a
  `runtime.ConfigSnapshot` during `init`.
- Hooks import the `runtime` package and read typed accessors
  (e.g. `runtime.NetHTTPClient()`). Accessors return deep copies; a binary built
  without the baked package panics on access with a clear message.
- The local per-machine env flow exists only in validation/resolution tooling
  (`validate`, `resolve`); it cannot change a baked binary's behavior.

### Rules

1. Bake is deterministic for identical manifest + config + environment.
2. Committed `baked/` output is generated from an example config and drift-gated
   by `make check` and CI (`bake --check`), mirroring `generate --check`.
3. `bake` is standalone; it is **not** wired into otelc's `-toolexec` pipeline in
   this repository.
4. Baked values are captured at build time, including `OTEL_GO_*` overrides, so
   runtime environment changes have no effect.

## Consequences

- Hooks never parse YAML, never call `os.Getenv` for option values, and never
  read undeclared stringly-typed paths.
- The future `guard` (Phase 4) can statically verify hooks only access declared
  options.
- Trade-off: a redeploy requires a rebuild to change configuration. Open question
  recorded in RFC 0001 (env precedence at bake time vs. runtime).

## Alternatives considered

- **Runtime parsing by hooks** (rejected) — contradicts the #705 design note and
  adds a parser dependency to every binary.
- **Embedded JSON + `go:embed`** (considered) — viable, but a literal Go composite
  value is statically analyzable and needs no unmarshalling.
- **Global mutable config set at init** (rejected) — loses the build-time
  contract and cannot be guarded statically.