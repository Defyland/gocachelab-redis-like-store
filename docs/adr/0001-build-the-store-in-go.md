# ADR 0001 - Build the Store in Go

## Status

Accepted

## Context

The product is a systems programming challenge: TCP sockets, goroutines,
shared-memory synchronization, file persistence, and pprof are central to the
evidence.

## Options Considered

1. Go.
2. Rust.
3. Elixir.

## Decision

Use Go and the standard library for the first implementation.

## Consequences

Positive:

- Direct fit for goroutine-per-connection TCP serving.
- `sync`, `net`, `pprof`, and native benchmarks are first-class.
- Simple binary deployment.

Negative:

- Memory safety depends on tests and race detector rather than ownership types.
- Manual protocol parsing and AOF handling need disciplined tests.

