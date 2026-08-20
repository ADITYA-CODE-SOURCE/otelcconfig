# The end-to-end demo

`cmd/demo` proves the whole supply chain in one run: a declarative YAML file is
validated, resolved, frozen into a Go package at build time, and consumed by an
instrumentation hook that never touches YAML itself.

## Run it

```bash
make demo-run
```

`demo-run` first re-bakes `examples/demo.yaml` into `baked/` (`otelcconfig bake
--output baked examples/demo.yaml`) and then runs the demo binary. This matches
the real workflow: a hook author changes behavior by editing YAML and rebuilding,
not by editing hook code.

## What the demo does

The demo spins up a local HTTP server and fires one request through a transport
wrapper modeled on otelc's `client_hook.go` enabler pattern
(`demo/hook.go`). The request carries two query parameters —
`token=supersecret` and `keep=visible` — plus a `User-Agent` and `X-Request-Id`
header.

The behavior is defined in `examples/demo.yaml`, not in Go:

```yaml
instrumentation:
  development:
    general:
      http:
        client:
          request_captured_headers:
            - user-agent
            - x-request-id
      sanitization:
        url:
          sensitive_query_parameters:
            - token
    go:
      nethttp:
        client:
          enabled: true
```

Output:

```text
method=GET
url=http://127.0.0.1:50459/demo?keep=visible&token=[REDACTED]
header user-agent=otelcconfig-demo/1.0
header x-request-id=demo-request-42
enabled=true
```

## What it demonstrates, step by step

1. **Manifest-driven generation.** `request_captured_headers`,
   `sensitive_query_parameters`, and `enabled` all come from the behavior
   manifest (`manifest/nethttp/metadata.yaml`), which generated the schema the
   file is validated against.
2. **Build-time validation and resolution.** `examples/demo.yaml` is validated
   strictly against the generated schema and resolved with
   `env > file > defaults` precedence before it is baked.
3. **Baking freezes behavior.** `otelcconfig bake` writes `baked/` as a Go
   package that registers itself via `runtime.Register` at init. The demo
   imports it with a blank import (`_ "…/baked"`) — no runtime configuration
   loading.
4. **Hooks consume a typed API.** `demo.Client` reads values via the `runtime`
   accessors (`runtime.NetHTTPClient().RequestCapturedHeaders`, …). The demo
   code contains no YAML parsing and no reads of option environment variables,
   which is exactly what the guard check verifies:

   ```bash
   otelcconfig guard ./demo ./runtime ./baked ./cmd/demo
   # guard: no undeclared configuration access found
   ```

5. **The behavior is observable.** The captured headers appear in the output; the
   sensitive `token` parameter is redacted to `[REDACTED]`. Both effects come
   from YAML alone. Changing `sensitive_query_parameters`, re-running
   `make demo-run`, and watching `token` appear in the URL is a fast, convincing
   demonstration that configuration — not hook code — drives behavior.

The integration test in `demo/` repeats this end to end: it builds the demo
binary, runs it, and asserts the redaction and header-capture behavior.

## Reading the demo

```text
examples/demo.yaml      the configuration (user input)
manifest/nethttp/       the behavior manifest (source of truth)
generated/              committed, schema + code derived from the manifest
baked/                  committed, frozen configuration from demo.yaml
runtime/                typed accessor API consumed by hooks
demo/                   the net/http hook, modeled on otelc's enabler pattern
cmd/demo/               the binary that ties it together
```