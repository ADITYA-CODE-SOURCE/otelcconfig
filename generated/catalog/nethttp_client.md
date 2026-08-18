# nethttp.client — Configuration Catalog

> Generated from the behavior manifest by otelcconfig. Do not edit.

HTTP client compile-time instrumentation built into otelc (net/http).

Library: https://pkg.go.dev/net/http

| Option | Declarative path | Type | Default | Stability | Env var | Description |
|--------|------------------|------|---------|-----------|---------|-------------|
| `enabled` | `go.nethttp.client.enabled` | `boolean` | `true` | `stable` | `OTEL_GO_ENABLED_INSTRUMENTATIONS` | Whether the net/http client instrumentation emits telemetry. Compatible with otelc's runtime instrumentation enable/disable lists. |
| `request_captured_headers` | `general.http.client.request_captured_headers` | `list_string` | `[]` | `experimental` | — | Outbound request headers to capture as span attributes. |
| `sensitive_query_parameters` | `general.sanitization.url.sensitive_query_parameters` | `list_string` | `[]` | `experimental` | — | Query parameter names whose values are redacted from URLs. |
