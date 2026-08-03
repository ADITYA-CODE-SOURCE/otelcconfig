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
[![Status](https://img.shields.io/badge/status-phase%200%20foundation-yellow)](https://github.com/ADITYA-CODE-SOURCE/otelcconfig/releases)

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
typed runtime API consumed by instrumentation hooks
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
| 0 — Foundation | `v0.1.0` | **Current** |
| 1 — Manifest + codegen | `v0.2.0` | Planned |
| 2 — Validate + resolve | `v0.3.0` | Planned |
| 3 — Typed runtime demo + RFC | `v0.4.0` | Planned |
| 4 — Static-analysis guard | `v0.5.0` | Stretch |

Phase 0 ships a professional skeleton, ADRs, architecture docs, CI, and a CLI that reports
honest "not implemented" messages. No configuration or generation claims are made yet.

## Quick start (Phase 0)

```bash
git clone https://github.com/ADITYA-CODE-SOURCE/otelcconfig.git
cd otelcconfig
make check
go run ./cmd/otelcconfig version
go run ./cmd/otelcconfig --help
```

## Planned CLI (later phases)

```text
otelcconfig generate   # Phase 1 — generate types, defaults, schema, docs
otelcconfig validate   # Phase 2 — validate user YAML against generated schema
otelcconfig resolve    # Phase 2 — show final values and sources
otelcconfig explain    # Phase 2 — explain one option
otelcconfig catalog    # Phase 2 — list all options
otelcconfig bake       # Phase 3 — model build-time config embedding
otelcconfig guard      # Phase 4 — reject undeclared config access
otelcconfig diff       # Phase 4 — compare two configs
```

## Architecture overview

See [docs/architecture.md](docs/architecture.md).

Key decisions are recorded as ADRs under [docs/adr/](docs/adr/):

- [ADR-0001](docs/adr/0001-record-architecture-decisions.md) — Record architecture decisions
- [ADR-0002](docs/adr/0002-adopt-otel-declarative-config-node.md) — Adopt OTel declarative config node

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

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Small, well-tested pull requests are welcome.
Please keep the independent-prototype positioning clear in any public discussion.

## Security

See [SECURITY.md](SECURITY.md).

## License

[Apache License 2.0](LICENSE)
