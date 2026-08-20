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
[![Status](https://img.shields.io/badge/status-roadmap%20complete-blue)](https://github.com/ADITYA-CODE-SOURCE/otelcconfig/releases)

## What this project is

Per-instrumentation **behavior configuration** for otelc, end to end:

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

Today otelc can only turn instrumentations on or off:

```bash
OTEL_GO_ENABLED_INSTRUMENTATIONS=nethttp,grpc
OTEL_GO_DISABLED_INSTRUMENTATIONS=nethttp
```

There is no way to configure *how* an instrumentation behaves — which HTTP headers
to capture, which sensitive URL parameters to redact, or how to gate a
semantic-convention migration. Issue #705 proposes adopting the OpenTelemetry
[declarative configuration](https://github.com/open-telemetry/opentelemetry-configuration)
`instrumentation/development` node plus a Java-agent-style per-instrumentation
manifest. `otelcconfig` implements and demonstrates that mechanism.

The design is described in
[RFC 0001](docs/rfc/0001-declarative-instrumentation-configuration.md) and the
roadmap-level details in [docs/architecture.md](docs/architecture.md).

## What it proves

- A **single behavior manifest** drives typed structs, defaults, a strict JSON
  Schema, a Markdown catalog, and env-var mappings — one source of truth, no drift.
- **Declarative YAML is validated and resolved at build time**: unknown keys are
  rejected, `${ENV:-default}` references resolve so missing environments fail
  early, and precedence is `env > file > defaults`.
- **Configuration is baked into the binary**: `otelcconfig bake` freezes the
  resolved values into a Go package. Hooks consume a typed `runtime` API and
  never parse YAML — configuration cannot change a built binary's behavior.
- **The contract is enforced, not assumed**: `otelcconfig guard` is a
  dependency-free static analyzer that rejects `runtime.Register` outside
  generated code, option reads from the environment, and YAML imports in hooks.
- **It works**: an end-to-end demo captures request headers and redacts sensitive
  query parameters, driven entirely by the baked configuration. See
  [docs/demo.md](docs/demo.md).

## Try it in 60 seconds

```bash
git clone https://github.com/ADITYA-CODE-SOURCE/otelcconfig.git
cd otelcconfig
make check        # fmt, tidy, vet, tests, drift + guard checks — nothing is modified
make demo-run     # bake examples/demo.yaml, then run the end-to-end demo
```

The demo fires an HTTP request through a hook modeled on otelc's net/http client
enabler pattern. The URL query parameter `token` is redacted and the configured
request headers are captured, purely from YAML frozen at build time:

```text
method=GET
url=http://127.0.0.1:50459/demo?keep=visible&token=[REDACTED]
header user-agent=otelcconfig-demo/1.0
header x-request-id=demo-request-42
enabled=true
```

## CLI

```text
otelcconfig generate   generate types, defaults, schema, docs (--check for drift)
otelcconfig validate   validate user YAML (+ ${ENV} substitution) against the schema
otelcconfig resolve    show final engine values and their sources
otelcconfig explain    explain one option (path or short name)
otelcconfig catalog    list all options (optionally filtered by instrumentation)
otelcconfig bake       freeze resolved config into a Go package (--check for drift)
otelcconfig guard      reject undeclared config access in hook directories
otelcconfig diff       compare two configs by resolved value (0 same, 1 different, 2 usage)
```

## Quick start

```bash
make check
make build VERSION=v0.5.0
./otelcconfig version
./otelcconfig validate examples/nethttp.yaml          # validate a config
./otelcconfig resolve  examples/nethttp.yaml          # final values + sources
./otelcconfig explain  request_captured_headers
./otelcconfig catalog
./otelcconfig bake --output baked --check examples/demo.yaml   # verify baked pkg current
./otelcconfig guard ./demo ./runtime ./baked          # reject undeclared access
./otelcconfig diff examples/minimal.yaml examples/nethttp.yaml
go run ./cmd/otelcconfig generate                     # regenerate derived artifacts
go run ./cmd/otelcconfig generate --check             # verify committed artifacts are current
```

`make check` is intentionally non-mutating: it fails on unformatted sources or
untidy module files instead of silently rewriting them (`make fmt`, `make tidy`,
and `make lint` are available). CI runs the same checks plus the race detector
on Linux, macOS, and Windows.

## Why the pipeline is shaped this way

- **Manifest → generated schema.** The manifest for the net/http instrumentation
  (`manifest/nethttp/metadata.yaml`) is the single source of truth; everything the
  pipeline validates, documents, and resolves derives from it.
- **Bake at build time (ADR-0003).** Hooks never handle YAML. Freezing resolved
  values in code keeps the dependency-light, enables static analysis, and matches
  the "baked config" design note in Issue #705.
- **Guard as the enforcement layer (ADR-0004).** Import- and name-based analysis
  (go/ast) keeps the module dependency-free while mechanically enforcing that the
  typed runtime API is the only legitimate access path — on this repo via
  `make guard-check` and CI.
- **Diff as an audit tool.** `otelcconfig diff` compares two files by resolved
  engine value, so reviewers see the effective behavior change, not raw YAML.

The full data flow, package layout, and the runtime contract live in
[docs/architecture.md](docs/architecture.md); every major decision has a recorded
ADR under [docs/adr/](docs/adr/).

## Roadmap status

| Phase | Release | Status |
|-------|---------|--------|
| 0 — Foundation | `v0.1.1` | Done |
| 1 — Manifest + codegen | `v0.2.0` | Done |
| 2 — Validate + resolve | `v0.3.0` | Done |
| 3 — Typed runtime demo + RFC | `v0.4.0` | Done |
| 4 — Static-analysis guard + diff | `v0.5.0` | Done |

The roadmap is complete. Every command above is implemented, tested (including
race conditions and drift checks), and enforced in CI.

## What this project is not

- Not an official OpenTelemetry component
- Not affiliated with or endorsed by the OpenTelemetry Go Compile Instrumentation SIG
- Not the upstream RFC itself, nor a replacement for one
- Not integrated into otelc's `-toolexec` compile-time rewriting pipeline
- Not a selection mechanism (that belongs to `otel.instrumentation.go` / ADR-0005)

Module path: `github.com/ADITYA-CODE-SOURCE/otelcconfig`
Do **not** import this under `go.opentelemetry.io/...`.

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