# LogSage Architecture Context

CodeRabbit should review changes with these invariants in mind.

## CLI

`cmd/logsage` is a thin adapter layer.

Responsibilities:
- parse CLI flags
- collect input
- call the engine

It must NOT implement detection logic.

## Engine

All analysis logic lives in:

internal/engine

The pipeline is:

Normalization → Signal Extraction → Detection → Scoring → Explanation → Recommendations

## Determinism

The engine must produce identical results for identical inputs.

Avoid:

- map iteration ordering
- non-deterministic sorting
- randomization

## Evidence Requirement

Every hypothesis must contain at least one evidence entry.

Never generate hypotheses without signal matches.

## Detectors

Detectors must:

- never panic
- tolerate missing signals
- operate only on SignalSet
- return deterministic results

## Test Policy

Modified packages must maintain ≥90% coverage.

All new logic must include tests.

Coverage verification commands:

go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
