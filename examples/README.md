# Examples

Example declarative configuration files for the OpenTelemetry
`instrumentation/development` shape consumed by the nethttp.client behavior
manifest (Phase 2) and, for `demo.yaml`, frozen into the demo binary (Phase 3).

- `minimal.yaml` — smallest valid example; everything falls back to defaults
- `nethttp.yaml` — sets `general.*` values and disables the net/http client
- `demo.yaml` — enables the client and is baked into `cmd/demo` (captured
  headers + redacted query values)

Try them:

```bash
go run ./cmd/otelcconfig validate examples/minimal.yaml
go run ./cmd/otelcconfig validate examples/nethttp.yaml
go run ./cmd/otelcconfig resolve  examples/nethttp.yaml
go run ./cmd/otelcconfig bake --output baked --check examples/demo.yaml
go run ./cmd/demo
```

See [docs/architecture.md](../docs/architecture.md) for the intended YAML shape
under the official OpenTelemetry declarative configuration model.