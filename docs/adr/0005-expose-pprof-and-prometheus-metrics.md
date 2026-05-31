# ADR 0005 - Expose pprof and Prometheus Metrics

## Status

Accepted

## Context

The interesting failure modes are latency, memory growth, lock contention, AOF
errors, and cleanup lag. Operators need runtime evidence without attaching a
debugger.

## Options Considered

1. Logs only.
2. `INFO` only.
3. Admin HTTP with Prometheus metrics and pprof.

## Decision

Expose `INFO` over TCP and `/metrics`, `/healthz`, `/readyz`, and pprof over an
admin HTTP listener.

## Consequences

Positive:

- Works with Prometheus and Grafana without client protocol changes.
- pprof gives direct CPU and heap evidence during benchmarks.
- Health and readiness are easy for Compose and supervisors.

Negative:

- pprof is sensitive and must stay on a trusted interface.
- Text metrics do not provide command latency histograms yet.

