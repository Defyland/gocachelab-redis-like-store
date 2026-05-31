# Bounded Contexts

## Command Protocol

Owns parsing inline commands, quoting rules, and Redis-like response rendering.
It does not know how keys are stored or persisted.

## Keyspace

Owns key entries, TTL semantics, lazy expiration, background cleanup, and
snapshot materialization. The keyspace is the live source of truth.

## Durability

Owns AOF record encoding, fsync policy, replay reports, and corrupted or partial
record handling. It does not decide client command semantics.

## Serving and Operations

Owns TCP connection lifecycle, admin HTTP endpoints, metrics, pprof, and
structured logs. It coordinates command execution across the other contexts.

