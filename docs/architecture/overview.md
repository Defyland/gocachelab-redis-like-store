# Architecture Overview

Gocachelab is a single-process Go service with two listeners:

- TCP cache listener for command traffic.
- HTTP admin listener for health, readiness, metrics, and pprof.

The central design keeps ownership narrow. The TCP server coordinates commands
but does not own map invariants. The store owns key state but does not parse
network input. The AOF package owns record integrity but does not know client
semantics beyond durable command payloads.

## Runtime Flow

1. Load snapshot if present.
2. Replay AOF records and record replay counters.
3. Open AOF appender.
4. Start TTL cleanup ticker.
5. Start admin HTTP listener.
6. Start TCP listener and accept clients concurrently.

## Rejected Complexity

- Early sharded maps: higher complexity before lock contention is measured.
- RESP-first parser: compatibility work would hide the core systems exercise.
- External Prometheus library: standard text output is enough for this scope.
- External WAL library: the AOF format is intentionally small and testable.

