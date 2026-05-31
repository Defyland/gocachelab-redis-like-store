# Engineering Case Study

## 1. Product Context

Gocachelab is a Redis-like cache node for demonstrating backend systems work:
TCP protocol handling, shared-memory concurrency, TTL, persistence, recovery,
observability, and performance analysis.

## 2. Domain Model

The core aggregate is a key entry with a string key, string value, and optional
absolute expiration timestamp. AOF records are internal durability records for
replaying mutating commands.

## 3. Architecture

The service is a single Go process with a TCP data listener and an HTTP admin
listener. Modules separate protocol parsing, keyspace state, AOF, serving, and
metrics.

## 4. Key Trade-offs

- `RWMutex` before sharding: simpler correctness, known write contention.
- Lazy plus background TTL: bounded implementation, accepted cleanup lag.
- AOF before replication: local durability evidence, accepted file growth.
- Inline protocol before RESP: easier audit, no existing Redis client support.

## 5. Data Model

Live state is a Go map protected by one lock. Durable state is an AOF file with
length and CRC32 per record. Snapshot state is JSON containing only live entries
at save time.

## 6. Consistency Model

Within one process, each command observes the result of prior completed commands.
AOF replay preserves record order. `fsync=everysec` may lose up to roughly one
second of acknowledged writes during a host crash.

## 7. Failure Scenarios

The implementation tests invalid commands, partial client disconnects,
corrupted AOF records, partial trailing AOF records, TTL cleanup races, and
concurrent clients.

## 8. Performance Strategy

Benchmarks focus on hot paths: SET, GET on a million-key dataset, mixed 80/20
load, 100 concurrent logical clients, TTL cleanup, and AOF replay. pprof is
available for CPU and heap evidence during runs.

## 9. Scalability Strategy

The first limit is lock contention and memory growth in one process. The next
step is a sharded store behind the same API, followed later by key partitioning
across nodes if the product goal changes from demonstration to service.

## 10. Security Model

The service assumes trusted local or private network access. It documents the
lack of auth, keeps pprof on the admin listener, bounds command size and `KEYS`,
avoids value logging, and protects AOF files with process-user permissions.

## 11. Observability

`INFO` exposes server counters over TCP. `/metrics` exposes Prometheus text
metrics. `/debug/pprof` exposes runtime profiles. Logs are structured through
Go `slog`.

## 12. Operational Cost

The low infrastructure cost is one binary and local disk. The accepted operating
cost is monitoring memory, AOF growth, pprof exposure, replay counters, and
cleanup lag.

## 13. Maintainability

The code is intentionally modular. RESP support can be added in protocol,
sharding in store, compaction in AOF, and new admin endpoints in server without
cross-cutting rewrites.

## 14. Product Decisions

The product favors transparent engineering evidence over feature breadth. That
is why multi-data-type Redis compatibility, replication, ACLs, and eviction are
excluded from v1.

## 15. What I Would Do Next

Add RESP compatibility, AOF compaction, command latency histograms, connection
quotas, and a sharded store experiment with before-and-after benchmark results.

