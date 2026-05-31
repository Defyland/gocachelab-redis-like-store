# ADR 0002 - Use RWMutex for v1 Concurrency

## Status

Accepted

## Context

The store must be correct under concurrent clients before it is clever. A single
map guarded by `sync.RWMutex` is easy to reason about and easy to test with the
race detector.

## Options Considered

1. Single `sync.RWMutex`.
2. Sharded map with many locks.
3. `sync.Map`.

## Decision

Use one `sync.RWMutex` around the keyspace map.

## Consequences

Positive:

- Clear invariant: all map access is lock-protected.
- Small implementation with deterministic tests.
- Future sharding can preserve the public store API.

Negative:

- Write-heavy workloads contend on one lock.
- Lazy expiration can upgrade from read observation to write deletion.
- Long `KEYS` scans hold the write lock because they may delete expired keys.

