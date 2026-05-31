# ADR 0004 - Use AOF Durability Before Replication

## Status

Accepted

## Context

The project needs durable recovery evidence, not distributed-system breadth. An
append-only file shows the core trade-off between latency, fsync policy, and
replay time.

## Options Considered

1. No persistence.
2. JSON snapshot only.
3. AOF with checksummed records.
4. Replication stream.

## Decision

Use an AOF with length and CRC32 per record, plus optional JSON snapshots.

## Consequences

Positive:

- Mutating command history is replayable.
- Partial trailing records can be detected.
- Corrupted records are counted rather than silently accepted.

Negative:

- AOF grows until compaction is added.
- `fsync=always` increases write latency.
- `fsync=everysec` accepts a small crash-loss window.

