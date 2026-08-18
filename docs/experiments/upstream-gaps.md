# Upstream Gaps

> Independent prototype notes. Not an official OpenTelemetry proposal document.

This file records configuration needs discovered while building `otelcconfig`
that are **not** present (or not stable) in the official
[OpenTelemetry declarative configuration](https://github.com/open-telemetry/opentelemetry-configuration)
`instrumentation/development` schema.

Gaps are staged here instead of inventing unofficial keys that look official.

## Process

1. Prefer an existing `general.*` key when one exists.
2. If no `general.*` key exists, place the option under the open `go:` map and
   mark `stability: development`.
3. Document the gap here with:
   - Why it is needed
   - Where it lives temporarily (`go:`)
   - Whether a cross-language `general.*` proposal would make sense
4. Do not claim upstream acceptance until a proposal lands.

## Current gaps

### `go.nethttp.client.enabled` — per-instrumentation enable (Phase 1)

The net/http behavior manifest declares `enabled` under the prototype-owned `go:`
map because no cross-language `instrumentation/development` key expresses "this
instrumentation is enabled". The official schema models stability versions and
`general.*` runtime behavior, not per-instrumentation enablement.

- **Why it is needed:** otelc today enables instrumentations via the
  `OTEL_GO_ENABLED_INSTRUMENTATIONS` / `OTEL_GO_DISABLED_INSTRUMENTATIONS` env vars.
  A declarative equivalent belongs in the config file.
- **Where it lives:** `go.nethttp.client.enabled` (open `go:` map).
- **Stability:** `stable` — it mirrors the existing otelc runtime enable surface via
  `csv_contains` semantics, so it is compatibility-safe in the MVP.
- **Cross-language proposal:** possible as a `general.*` key, but out of scope for
  Phase 1; record here until upstream takes it up.
