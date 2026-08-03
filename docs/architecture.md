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
| `config` | Load, substitute, resolve, typed runtime | 2–3 |
| `cmd/otelcconfig` | CLI entrypoint | 0+ |
| `demo` | net/http end-to-end proof | 3 |
| `internal/yamlutil` | Shared YAML helpers | 1+ |
| `internal/guard` | Static analysis (stretch) | 4 |

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

## Runtime contract (planned)

Hooks read configuration only through a typed API. They never:

- Open or parse YAML at runtime
- Call `os.Getenv` for option values directly (compat env vars are resolved once)
- Access stringly-typed undeclared paths

This matches the Issue #705 design note:

> otelc parses and validates the config at build time and bakes the resolved
> values into the instrumented binary. The hooks never touch YAML.

Phase 3 models bake-in as a standalone `otelcconfig bake` command. It is **not**
wired into otelc's `-toolexec` pipeline.

## Official keys used in the first instrumentation (planned)

For net/http client demonstration:

| Option | Declarative path | Origin |
|--------|------------------|--------|
| Enabled | `go.nethttp.enabled` | Prototype-owned (`go:` open map) |
| Request captured headers | `general.http.client.request_captured_headers` | Official OTel schema |
| Sensitive query parameters | `general.sanitization.url.sensitive_query_parameters` | Official OTel schema |

Compatibility env vars (existing otelc surface only in MVP):

- `OTEL_GO_ENABLED_INSTRUMENTATIONS`
- `OTEL_GO_DISABLED_INSTRUMENTATIONS`

## Phase status

Phase 0 (this release) provides the skeleton and documentation only.
See the README roadmap for subsequent phases.
