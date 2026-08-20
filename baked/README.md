# baked

Committed, generated package produced by `otelcconfig bake`.

`nethttp_client_gen.go` is a Go literal of the resolved `examples/demo.yaml`
configuration, registered with the `runtime` package at init. Hooks (see
`demo/`) read frozen values through `runtime` and never parse YAML.
`nethttp_client.json` is a JSON audit of the resolved values and their sources.

Regenerate with:

```bash
make bake
```

Verify it is current with `make bake-check` (also part of `make check` and CI).

Do not edit the generated files by hand.