# otelcconfig

> **otelcconfig is an independent Go prototype exploring the mechanism proposed in
> [OpenTelemetry Go Compile Instrumentation Issue #705](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/issues/705).
> It is intended to provide implementation evidence and study architectural trade-offs.
> It is not an official or maintainer-approved OpenTelemetry component.**

Declarative Configuration Toolkit for [otelc](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation)
(OpenTelemetry Go Compile Instrumentation).

[![CI](https://github.com/ADITYA-CODE-SOURCE/otelcconfig/actions/workflows/ci.yml/badge.svg)](https://github.com/ADITYA-CODE-SOURCE/otelcconfig/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/ADITYA-CODE-SOURCE/otelcconfig.svg)](https://pkg.go.dev/github.com/ADITYA-CODE-SOURCE/otelcconfig)
[![Status](https://img.shields.io/badge/status-phase%203%20runtime-blue)](https://github.com/ADITYA-CODE-SOURCE/otelcconfig/releases)

## What this project is

`otelcconfig` explores how per-instrumentation **behavior configuration** could work for otelc:

```text
behavior manifest (metadata.yaml)
        ↓
typed Go structs + defaults + JSON Schema + docs + env mappings
        ↓
user declarative YAML (instrumentation/development)
        ↓
env vars  >  config file  >  defaults
        ↓
`otelcconfig bake` freezes resolved values into a Go package (build time)
        ↓
typed runtime API consumed by instrumentation hooks (never parses YAML)
```

It is a **standalone design-exploration prototype** of the mechanism described in:

- Upstream issue: [opentelemetry-go-compile-instrumentation#705](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/issues/705)
- LFX proposal: [cncf/mentoring#1939](https://github.com/cncf/mentoring/issues/1939) — *Declarative instrumentation configuration for otelc* (2026 Term 3)

## What this project is not

- Not an official OpenTelemetry component
- Not affiliated with or endorsed by the OpenTelemetry Go Compile Instrumentation SIG
- Not a replacement for the RFC planned during the LFX mentorship
- Not integrated into otelc's `-toolexec` compile-time rewriting pipeline
- Not a selection mechanism (that belongs to `otel.instrumentation.go` / ADR-0005)

Module path: `github.com/ADITYA-CODE-SOURCE/otelcconfig`  
Do **not** import this under `go.opentelemetry.io/...`.

## Problem

Today otelc can turn instrumentations on or off with:

```bash
OTEL_GO_ENABLED_INSTRUMENTATIONS=nethttp,grpc
OTEL_GO_DISABLED_INSTRUMENTATIONS=nethttp
```

There is no way to configure *how* an instrumentation behaves — capture specific HTTP headers,
redact sensitive URL query parameters, or gate a semantic-convention migration.
Issue #705 proposes adopting the OpenTelemetry
[declarative configuration](https://github.com/open-telemetry/opentelemetry-configuration)
`instrumentation/development` node and a Java-agent-style per-instrumentation manifest.

## Status

| Phase | Release | Status |
|-------|---------|--------|
| 0 — Foundation | `v0.1.1` | Done |
| 1 — Manifest + codegen | `v0.2.0` | Done |
| 2 — Validate + resolve | `v0.3.0` | Done |
| 3 — Typed runtime demo + RFC | `v0.4.0` | **Current** |
| 4 — Static-analysis guard | `v0.5.0` | Stretch |

Phase 1 shipped the first behavior manifest (`manifest/nethttp/metadata.yaml`) and a
deterministic code-generation pipeline. `otelcconfig generate` derives typed structs,
defaults, env-var mappings, a JSON Schema fragment, and a Markdown catalog; committed
outputs are golden-tested and CI fails on drift via `generate --check`.

Phase 2 turned the generated schema into a working config pipeline.
`otelcconfig validate` loads and validates user declarative YAML
(`instrumentation/development`), `otelcconfig resolve` shows the final engine
values plus their sources, and `explain`/`catalog` document every option.
Validation is strict: unknown keys outside the declared `go:` subtree are rejected,
and `${ENV:-default}`-style references are resolved so unresolved environments fail
early at build time.

Phase 3 (current) proves the full supply chain. `otelcconfig bake` freezes resolved
configuration into a Go package (`baked/`) that registers it with the `runtime`
package at init; instrumentation hooks consume the typed runtime API and never
parse YAML. The `demo` package models otelc's net/http client enabler pattern, and
`cmd/demo` is an end-to-end demonstration: declarative YAML → bake → binary →
captured headers and redacted query parameters, asserted by an integration test.
The mechanism is documented in [RFC 0001](docs/rfc/0001-declarative-instrumentation-configuration.md).

> **Resume status:** Phase 3 is now a complete, tested end-to-end demonstration of
> the mechanism — configurable from YAML, validated, resolved, baked, and consumed
> by hooks without runtime parsing. `otelcconfig` can be described as a substantial
> LFX project. Remaining roadmap value sits in the Phase 4 static-analysis guard
> and any upstream RFC discussion.

## Quick start (Phase 3)

```bash
git clone https://github.com/ADITYA-CODE-SOURCE/otelcconfig.git
cd otelcconfig
make check
make build VERSION=v0.4.0
./otelcconfig version
go run ./cmd/otelcconfig validate examples/nethttp.yaml  # validate a config
go run ./cmd/otelcconfig resolve  examples/nethttp.yaml  # final values + sources
go run ./cmd/otelcconfig explain  request_captured_headers
go run ./cmd/otelcconfig catalog
go run ./cmd/otelcconfig bake --output baked --check examples/demo.yaml  # verify baked pkg current
make demo-run                                                     # bake + run the demo
go run ./cmd/otelcconfig generate       # regenerate derived artifacts
go run ./cmd/otelcconfig generate --check  # verify committed artifacts are current
```

`make check` is intentionally non-mutating: it fails on unformatted sources or
untidy module files instead of silently rewriting them. Run `make fmt` and
`make tidy` explicitly when needed. `make lint` requires `golangci-lint`.

## CLI

Implemented:

```text
otelcconfig generate   # Phase 1 — generate types, defaults, schema, docs (--check for drift)
otelcconfig validate   # Phase 2 — validate user YAML (+ ${ENV} substitution) against schema
otelcconfig resolve    # Phase 2 — show final engine values and their sources
otelcconfig explain    # Phase 2 — explain one option (path or short name)
otelcconfig catalog    # Phase 2 — list all options (optionally filtered by instrumentation)
otelcconfig bake       # Phase 3 — freeze resolved config into a Go package (--check for drift)
```

Planned (later phases):

```text
otelcconfig guard      # Phase 4 — reject undeclared config access
otelcconfig diff       # Phase 4 — compare two configs
```

## Architecture overview

See [docs/architecture.md](docs/architecture.md).

Key decisions are recorded as ADRs under [docs/adr/](docs/adr/):

- [ADR-0001](docs/adr/0001-record-architecture-decisions.md) — Record architecture decisions
- [ADR-0002](docs/adr/0002-adopt-otel-declarative-config-node.md) — Adopt OTel declarative config node
- [ADR-0003](docs/adr/0003-bake-at-build-time.md) — Bake configuration at build time

## Relationship to otelc

| Concern | Owned by | Notes |
|---------|----------|-------|
| What gets woven (selection) | otelc `otel.instrumentation.go` / ADR-0005 | Out of scope here |
| How it behaves (configuration) | Issue #705 / this prototype | Mechanism exploration |
| What telemetry is emitted | otelc Weaver registry `schemas/otelc/` | Separate from behavior manifests |

## Non-goals

- Frontend / UI
- Inventing config keys outside the OTel declarative configuration model
- Replacing otelc's Weaver emission registry
- Direct patches to otelc's compiler pipeline in this repository

## LFX application positioning

When this project reaches Phase 3, describe it accurately as:

> An independent Go prototype exploring the mechanism proposed in otelc Issue #705,
> with manifest-driven generation, declarative validation, backward-compatible
> resolution, and typed runtime hooks.

Do not claim that this repository implements Issue #705 upstream, is accepted by
OpenTelemetry maintainers, or is an official OpenTelemetry contribution. Upstream
contributor status comes from accepted participation in the upstream project; owning
this independent repository does not by itself make the author an OpenTelemetry contributor.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Small, well-tested pull requests are welcome.
Please keep the independent-prototype positioning clear in any public discussion.

## Security

See [SECURITY.md](SECURITY.md).

## License

[Apache License 2.0](LICENSE)
