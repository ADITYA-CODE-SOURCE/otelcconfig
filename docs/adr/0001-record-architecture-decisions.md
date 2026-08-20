# ADR-0001: Record Architecture Decisions

- **Status:** Accepted
- **Date:** 2026-08-03
- **Deciders:** Project maintainers

## Context

`otelcconfig` explores a non-trivial configuration mechanism related to
OpenTelemetry Go Compile Instrumentation Issue #705. Design choices around
manifest format, code generation, precedence, and runtime APIs will evolve.
Without a written record, later contributors cannot tell
*why* a decision was made or which alternatives were rejected.

The upstream otelc project records significant decisions as ADRs under
`docs/adr/` (see otelc ADR-0001 through ADR-0005). Matching that practice
keeps this prototype legible to otelc maintainers.

## Decision

Significant architectural decisions for `otelcconfig` are recorded as
numbered Architecture Decision Records in `docs/adr/`.

Each ADR includes:

- Status (Proposed / Accepted / Superseded / Deprecated)
- Context
- Decision
- Consequences
- Alternatives considered (when relevant)

Superseded ADRs are marked, never deleted.

## Consequences

- Design rationale is reviewable without reading git history.
- Reviewers can challenge a decision by opening a PR against the ADR.
- Small implementation details do not require ADRs; only decisions that
  constrain future work do.

## Alternatives considered

- **Only README prose** — insufficient structure; hard to supersede cleanly.
- **Git commit messages only** — not discoverable as a design corpus.
