# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] — 2026-08-18

### Added

- Behavior manifest model (`manifest` package): typed schema for Java-agent-style
  per-instrumentation `metadata.yaml`, with validation, default decoding, and discovery
- First behavior manifest: `manifest/nethttp/metadata.yaml` with three upstream-aligned
  options (`enabled`, `request_captured_headers`, `sensitive_query_parameters`)
- Deterministic code generation (`codegen` package): typed Go structs and option-path
  constants, defaults constructors, env-var compatibility mappings, a JSON Schema
  fragment for the `instrumentation/development` shape, and a Markdown catalog
- `otelcconfig generate` command with `--manifest`, `--output`, and `--check` (drift detection)
- Committed generated artifacts under `generated/` with golden-file and determinism tests
- `make generate` and `make generate-check`; CI now fails when committed artifacts drift

### Changed

- `generate` is no longer a stub; remaining planned commands report honest "not implemented"
- `yaml.v3` promoted from indirect to direct dependency

### Notes

- Phase 1 only. Supports the net/http client instrumentation via existing otelc env surface.
- Next: Phase 2 — validate + resolve user declarative YAML (`v0.3.0`).

## [0.1.1] — 2026-08-03

### Changed

- Hardened CI with Go 1.25/current stable and Linux/macOS/Windows coverage
- Added race-detector and pinned golangci-lint checks
- Made `make check` non-mutating and added module-tidiness verification
- Strengthened CLI tests to assert stdout, stderr, and exit status
- Clarified honest LFX resume wording and independent contributor status

## [0.1.0] — 2026-08-03

### Added

- Project foundation: Go module `github.com/ADITYA-CODE-SOURCE/otelcconfig`
- Apache-2.0 license and SPDX headers
- README with independent-prototype disclaimer and LFX Issue #705 positioning
- ADR-0001 (record architecture decisions) and ADR-0002 (adopt OTel declarative config node)
- Architecture overview document
- CLI skeleton (`version`, `help`) with honest "not implemented" stubs for later commands
- Package stubs: `manifest`, `codegen`, `config`, `demo`, `internal/yamlutil`
- Makefile (`build`, `test`, `vet`, `fmt`, `lint`, `tidy`, `generate`, `check`)
- GitHub Actions CI, Dependabot, issue and PR templates
- Contributing, Code of Conduct, and Security policies

### Notes

- Phase 0 only. No configuration loading, code generation, or resolution yet.
- Next: Phase 1 — behavior manifest and deterministic code generation (`v0.2.0`).

[Unreleased]: https://github.com/ADITYA-CODE-SOURCE/otelcconfig/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/ADITYA-CODE-SOURCE/otelcconfig/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/ADITYA-CODE-SOURCE/otelcconfig/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/ADITYA-CODE-SOURCE/otelcconfig/releases/tag/v0.1.0
