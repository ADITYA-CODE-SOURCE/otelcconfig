# ADR-0002: Adopt the OpenTelemetry Declarative Configuration `instrumentation/development` Node

- **Status:** Accepted
- **Date:** 2026-08-03
- **Deciders:** Project maintainers
- **Related:** [otelc#705](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/issues/705),
  [opentelemetry-configuration](https://github.com/open-telemetry/opentelemetry-configuration)

## Context

Issue #705 proposes per-instrumentation behavior configuration for otelc.
The issue body states explicitly:

> Adopt the OTel declarative configuration model **instead of inventing our own format**.
> Its `instrumentation/development` node defines exactly this surface.

The upstream schema (`opentelemetry-configuration/schema/instrumentation.yaml`) defines:

- A closed `general:` node for cross-language options (`additionalProperties: false`)
- An open `go:` language-specific node (`ExperimentalLanguageSpecificInstrumentation`)
- Environment variable substitution `${ENV:-default}` (Stable in the OTel data model)

Inventing parallel keys (for example a custom `semconv_mode` enum or
`capture_request_headers`) would conflict with the mentors' stated requirement
and teach the wrong vocabulary to anyone reading this prototype.

## Decision

`otelcconfig` adopts the official OpenTelemetry declarative configuration
`instrumentation/development` node as its user-facing configuration shape.

### Rules

1. **Cross-language keys** use the official `general.*` paths when they exist, including:
   - `general.http.client.request_captured_headers`
   - `general.http.client.response_captured_headers`
   - `general.http.server.request_captured_headers`
   - `general.http.server.response_captured_headers`
   - `general.http.*.known_methods`
   - `general.<domain>.semconv.{version,experimental,dual_emit}`
   - `general.sanitization.url.sensitive_query_parameters`
   - `general.stability_opt_in_list`

2. **Language-specific keys** live under the open `go:` map and are owned by this
   prototype (for example `go.nethttp.enabled`). Experimental options are labelled
   `stability: development`.

3. **Gaps** (keys needed but missing upstream) are recorded in
   `docs/experiments/upstream-gaps.md` as proposal candidates — never silently
   invented as if they were official.

4. **Environment substitution** follows the OTel specification ABNF for
   `${ENV:-default}`.

5. **Precedence** is environment variables > configuration file > defaults,
   preserving compatibility with otelc's existing
   `OTEL_GO_ENABLED_INSTRUMENTATIONS` / `OTEL_GO_DISABLED_INSTRUMENTATIONS`.

## Consequences

- The prototype stays aligned with the LFX mentorship design surface.
- Maintainers reviewing the repo see familiar key names.
- Some desired options may not exist upstream yet; those must be staged under
  `go:` or filed as upstream gaps.

## Alternatives considered

- **Invent a custom top-level format** (rejected) — contradicts #705.
- **Only environment variables** (rejected) — does not exercise the declarative
  file + codegen path that #705 targets.
- **Reuse the Weaver emission registry as the behavior manifest** (rejected) —
  otelc's `schemas/otelc/groups/*.yaml` describes *what telemetry is emitted*,
  not *how instrumentation behaves*. Those concerns must stay separate.
