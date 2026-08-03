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

_None yet — Phase 0 has no manifests._

Phase 1 will add the first net/http behavior manifest and update this file if
any temporary `go:`-only options are required beyond official `general.*` keys.
