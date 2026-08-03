# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/ADITYA-CODE-SOURCE/otelcconfig/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ADITYA-CODE-SOURCE/otelcconfig/releases/tag/v0.1.0
