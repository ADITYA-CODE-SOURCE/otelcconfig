# Contributing to otelcconfig

Thank you for your interest in contributing.

## Positioning

This project is an **independent community prototype**. It is not an official
OpenTelemetry component. Contributions must preserve that positioning.

Do not:

- Present this project as official OpenTelemetry work
- Use `go.opentelemetry.io` import paths
- Invent configuration keys that conflict with the
  [OpenTelemetry declarative configuration](https://github.com/open-telemetry/opentelemetry-configuration)
  schema without documenting them under `docs/experiments/upstream-gaps.md`

## Development setup

```bash
git clone https://github.com/ADITYA-CODE-SOURCE/otelcconfig.git
cd otelcconfig
make check
```

Requirements:

- Go 1.25 or newer
- Optional: [golangci-lint](https://golangci-lint.run)

## Workflow

1. Open a GitHub issue describing the change.
2. Create a branch from `main` named `issue-<n>-short-description`.
3. Make small, focused commits.
4. Ensure `make check` passes.
5. Open a pull request against `main`.
6. Wait for CI to pass.

## Coding standards

- Apache-2.0 SPDX headers on every Go source file
- `gofmt` / `gofumpt` clean
- Package comments on every package
- Prefer table-driven tests
- No secrets in commits

## Phase roadmap

Work should align with the published phase plan in the README.
Prefer completing the current phase over starting later-phase features early.

## Code of conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
