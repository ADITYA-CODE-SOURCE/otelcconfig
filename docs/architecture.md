# Architecture Overview

> Independent prototype. Not an official OpenTelemetry component.
> Related upstream: [otelc#705](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/issues/705).

## Goal

Demonstrate a maintainable pipeline where each instrumentation declares its
configurable **behavior** options once, and every consumer artifact is generated
from that declaration:

```text
┌──────────────────────────┐
│  Behavior manifest       │  manifest/<inst>/metadata.yaml
│  (Java metadata.yaml     │  single source of truth
│   style)                 │
└────────────┬─────────────┘
             │ codegen (Phase 1)
             ▼
┌──────────────────────────┐
│  Generated artifacts     │  types · defaults · env map
│                          │  JSON Schema · docs catalog
└────────────┬─────────────┘
             │ load + validate (Phase 2)
             ▼
┌──────────────────────────┐
│  User YAML               │  instrumentation/development
│  + ${ENV:-default}       │  general: + go:
└────────────┬─────────────┘
             │ resolve (Phase 2)
             ▼
┌──────────────────────────┐
│  Precedence              │  env  >  file  >  defaults
│  + otelc enable/disable  │  OTEL_GO_ENABLED/DISABLED_...
└────────────┬─────────────┘
             │ typed API (Phase 3)
             ▼
┌──────────────────────────┐
│  Runtime config package  │  hooks never parse YAML
│  consumed by hooks       │  enabler pattern (otelc-style)
└────────────┬─────────────┘
             │ guard (Phase 4)
             ▼
┌──────────────────────────┐
│  go/analysis analyzer    │  reject undeclared option access
└──────────────────────────┘
```

## Package layout

| Package | Role | Phase |
|---------|------|-------|
| `manifest` | Parse and validate behavior manifests | 1 |
| `codegen` | Deterministic generators | 1 |
| `generated` | Committed, golden-tested outputs | 1 |
| `config` | Load, substitute, validate, resolve | 2 |
| `runtime` | Baked configuration accessors for hooks | 3 |
| `bake` | Freeze resolved config into a Go package | 3 |
| `baked` | Committed, drift-gated frozen configuration | 3 |
| `demo` | net/http end-to-end proof | 3 |
| `cmd/otelcconfig` | CLI entrypoint | 0+ |
| `cmd/demo` | End-to-end demo binary | 3 |
| `internal/yamlutil` | Shared YAML helpers | 1+ |
| `internal/guard` | Static analysis: reject undeclared config access | 4 |

## Two-manifest discipline

otelc already has a **Weaver emission registry** under `schemas/otelc/groups/*.yaml`.
That registry answers: *"what telemetry does otelc emit?"*

This project introduces a separate **behavior manifest**. It answers:
*"what can the user configure about how instrumentation behaves?"*

These must not be mixed. See [ADR-0002](adr/0002-adopt-otel-declarative-config-node.md).

## Split with otelc selection

Per otelc ADR-0005 and Issue #705:

- `otel.instrumentation.go` decides **what gets woven** (selection)
- Declarative behavior config decides **how it behaves** (this prototype)

Selection mechanisms (`.otel.yml`, tool files) are out of scope.

## Runtime contract (implemented in Phase 3)

Hooks read configuration only through a typed API. They never:

- Open or parse YAML at runtime
- Call `os.Getenv` for option values directly (compat env vars are resolved once at bake time)
- Access stringly-typed undeclared paths

This matches the Issue #705 design note:

> otelc parses and validates the config at build time and bakes the resolved
> values into the instrumented binary. The hooks never touch YAML.

The `runtime` package implements the accessor side: `otelcconfig bake` freezes the
resolved configuration into a generated package that registers a `ConfigSnapshot`
at init; hooks such as the `demo` enabler read frozen values through
`runtime.NetHTTPClient()`. Accessors return deep copies so hooks cannot mutate the
shared snapshot, and they panic with a clear message when a binary was built
without importing the baked package.

`otelcconfig bake` is standalone; it is **not** wired into otelc's `-toolexec`
pipeline.

## Official keys used in the first instrumentation

For net/http client demonstration (`manifest/nethttp/metadata.yaml`, Phase 1):

| Option | Declarative path | Origin |
|--------|------------------|--------|
| Enabled | `go.nethttp.client.enabled` | Prototype-owned (`go:` open map) |
| Request captured headers | `general.http.client.request_captured_headers` | Official OTel schema |
| Sensitive query parameters | `general.sanitization.url.sensitive_query_parameters` | Official OTel schema |

Compatibility env vars (existing otelc surface only in MVP):

- `OTEL_GO_ENABLED_INSTRUMENTATIONS`

## Phase status

### Phase 4 (v0.5.0, current)

Closes the roadmap with the final two commands:

- **`internal/guard` + `otelcconfig guard <dir...>`** — a dependency-free static
  analyzer (go/ast) that rejects undeclared configuration access in hook
  packages. Three rules:
  1. `runtime.Register` is only allowed in generated baked code (files carrying
     the `// Code generated by otelcconfig` header); hooks must not replace the
     snapshot.
  2. Option values never come from the environment: `os.Getenv` /
     `os.LookupEnv` calls whose argument matches a declared option env var
     (`OTEL_GO_ENABLED_INSTRUMENTATIONS`, `OTEL_GO_DISABLED_INSTRUMENTATIONS`)
     are flagged.
  3. Hooks never parse YAML at runtime: imports of `gopkg.in/yaml.v3` are
     flagged outside tooling packages (`config`, `cmd`).
- **`otelcconfig diff a.yaml b.yaml`** — resolves both files with the same
  pipeline and prints one `!` line per changed option value; exit 0 when
  identical, 1 when differences exist, 2 on usage errors (mirrors `diff`(1)).

`make guard-check` runs the analyzer over `./demo ./runtime ./baked ./cmd/demo`
and is part of `make check` and CI, so the runtime contract is enforced on the
repository itself.

### Phase 3 (v0.4.0, done)

Completes the supply chain from declarative YAML to typed runtime hooks:

- **`bake` package + `otelcconfig bake`** — freezes resolved values (including
  `OTEL_GO_*` env overrides) into a Go package (`baked/`) plus a JSON audit file.
  Deterministic for identical manifest + config + env; `--check` fails on drift.
  Committed `baked/` is regenerated from `examples/demo.yaml` and drift-gated in
  `make check` and CI.
- **`runtime` package** — the typed accessor layer hooks consume; hooks import it
  and never parse YAML. Accessors return deep copies; they panic if no snapshot
  was baked in.
- **`demo` package + `cmd/demo`** — an end-to-end net/http demonstration modeled on
  otelc's `client_hook.go` enabler pattern (no SDK dependency): captured request
  headers become observation attributes and sensitive query parameters are
  redacted, driven entirely by the baked configuration. An integration test builds
  and runs the binary and asserts the configured behavior.
- **RFC 0001** — `docs/rfc/0001-declarative-instrumentation-configuration.md`
  records the mechanism, alternatives, and open questions.
- **ADR-0003** — bake-at-build-time with the local (per-machine) env flow reserved
  for tooling only.

### Phase 2 (v0.3.0, done)

The `config` package implements the load/substitute/validate/resolve half of the
pipeline over the generated schema:

- **Load** — read and decode `instrumentation/development` YAML
- **Substitute** — resolve `${ENV:-default}` references before validation so an
  unresolved environment fails early at build time
- **Validate** — strict JSON Schema validation against the generated schema; the
  `general:` subtree is closed (`additionalProperties: false`) so undeclared official
  keys are rejected, while the prototype-owned `go:` map stays open per
  [docs/experiments/upstream-gaps.md](experiments/upstream-gaps.md)
- **Resolve** — compute the final engine values with precedence
  `env  >  file  >  defaults`, including the existing
  `OTEL_GO_ENABLED_INSTRUMENTATIONS` compatibility surface, and report each value's
  source for transparency

CLI: `validate`, `resolve`, `explain`, `catalog`. The per-machine local env flow in
validation/resolution tooling remains; production hooks consume only frozen values
via `runtime`, so local env cannot change a baked binary's behavior.

### Phase 1 (v0.2.0, done)

Added the first behavior manifest and a deterministic code-generation pipeline
(`otelcconfig generate`): typed structs, defaults, env-var mappings, a JSON Schema
fragment for the `instrumentation/development` shape, and a Markdown catalog.
Committed outputs under `generated/` are golden-tested and CI fails on drift via
`generate --check`.

Every roadmap phase is delivered; see the README for the complete status table.
