# Gocachelab Redis-like Store

## 1. What is this product?

Gocachelab is a Redis-like in-memory key-value store written in Go for backend
engineers who need to study systems programming without hiding behind Redis
itself. It exposes a TCP protocol, TTL semantics, append-only durability,
snapshotting, pprof, and metrics in one small codebase.

This repository is an R&D asset for engineers studying cache internals and
failure modes. It is intentionally runnable, but it is not positioned as a
hosted managed cache product.

## 2. Problem it solves

Many backend challenges only prove framework fluency. This project focuses on
the operational questions behind a cache service: how commands are parsed, how
concurrent clients share state, how expired keys disappear, how persistence
recovers after partial writes, and where latency is spent.

## 3. Target users

- Backend engineers preparing for systems-oriented interviews.
- Tech leads evaluating cache durability and latency trade-offs.
- Platform engineers who want a readable reference for TCP services in Go.

## 4. Main features

- Concurrent TCP clients with a simple inline protocol.
- Commands: `PING`, `SET`, `GET`, `DEL`, `EXISTS`, `EXPIRE`, `TTL`,
  `PERSIST`, bounded `KEYS`, `INFO`, `SAVE`, and `QUIT`.
- Thread-safe in-memory store with lazy expiration and background TTL cleanup.
- AOF persistence with replay reports for corrupted and partial records.
- JSON snapshot support through `SAVE`.
- Prometheus-compatible metrics and Go `pprof` on the admin HTTP listener.

## 5. Architecture overview

The runtime is a single Go process with clear module boundaries:

- `internal/protocol` parses client commands and renders Redis-like responses.
- `internal/store` owns key state, TTL invariants, snapshots, and cleanup.
- `internal/aof` encodes durable command records and replays them on startup.
- `internal/server` accepts TCP clients and exposes admin HTTP endpoints.
- `internal/metrics` records counters used by `INFO` and `/metrics`.

## 6. Tech stack

- Go `1.25.11`
- Standard-library networking, synchronization, JSON, HTTP, and pprof
- Native Go tests, benchmarks, race detector, and `go vet`
- Docker and GitHub Actions for build validation

## 7. Domain model

The core aggregate is a key entry: a string key, a string value, and an optional
absolute expiration timestamp. Expired entries may remain in memory until lazy
expiration or cleanup observes them, but public reads must not return them.

More detail: [domain invariants](docs/domain/invariants.md).

## 8. API documentation

The public data API is TCP, not HTTP. Command examples and response formats live
in [docs/api/README.md](docs/api/README.md). The HTTP admin surface is
documented in [openapi.yaml](openapi.yaml).

## 9. Async or event architecture

Gocachelab does not publish domain events. AOF records are internal durability
records and are documented in [docs/events/README.md](docs/events/README.md).

## 10. Database design

There is no external database. The source of live truth is the protected in-memory
map; durability is provided by a checksummed AOF and optional JSON snapshots.

## 11. Testing strategy

The test suite covers parser behavior, store invariants, TTL behavior, AOF
encoding and replay, snapshot restore, TCP command handling, concurrency races,
and expected failure modes such as invalid commands and client disconnects.

## 12. Performance benchmarks

Native Go benchmarks cover million-key datasets, mixed 80/20 workloads, 100
concurrent clients, TTL cleanup, and AOF replay. The methodology and current
results are in [benchmarks/baseline.md](benchmarks/baseline.md).

## 13. Observability

- `INFO` exposes runtime, store, TTL, and AOF counters over TCP.
- `/metrics` exposes Prometheus text metrics on the admin listener.
- `/debug/pprof/` is enabled on the admin listener.
- Structured logs include subsystem, operation, listener, and error fields.

## 14. Security considerations

The server intentionally ships without authentication because it models a
private cache node. The threat model documents the accepted boundary: bind to a
trusted network, restrict admin endpoints, validate command size, and avoid
logging values. See [docs/security/threat-model.md](docs/security/threat-model.md).

## 15. Trade-offs and decisions

Key ADRs:

- [ADR 0001 - Build the store in Go](docs/adr/0001-build-the-store-in-go.md)
- [ADR 0002 - Use RWMutex for v1 concurrency](docs/adr/0002-use-rwmutex-for-v1-concurrency.md)
- [ADR 0003 - Use lazy and background TTL expiration](docs/adr/0003-use-lazy-and-background-ttl-expiration.md)
- [ADR 0004 - Use AOF durability before replication](docs/adr/0004-use-aof-durability-before-replication.md)
- [ADR 0005 - Expose pprof and Prometheus metrics](docs/adr/0005-expose-pprof-and-prometheus-metrics.md)
- [ADR 0006 - Start with a simple inline protocol](docs/adr/0006-start-with-simple-inline-protocol.md)

## 16. How to run locally

```sh
go run ./cmd/gocachelab
```

Useful environment variables:

- `GOCACHELAB_TCP_ADDR`, default `127.0.0.1:7379`
- `GOCACHELAB_ADMIN_ADDR`, default `127.0.0.1:8080`
- `GOCACHELAB_DATA_DIR`, default `./data`
- `GOCACHELAB_AOF_FSYNC`, one of `never`, `everysec`, or `always`

Example:

```sh
printf 'SET user:1 Ada\r\nGET user:1\r\nQUIT\r\n' | nc 127.0.0.1 7379
```

## 16.1 Deployment truth

Gocachelab intentionally does not ship a Railway demo. As an R&D asset, the
product boundary is an unauthenticated TCP cache node plus a private admin HTTP
listener with pprof, and its durability story depends on local disk for the AOF
and snapshots.

A public single-node Railway-style deploy would misrepresent both the trust
boundary and the persistence contract. The truthful runnable surfaces in this
repository are the local binary and the private Docker Compose topology
documented in [docs/architecture/deployment-view.md](docs/architecture/deployment-view.md).

## 17. How to run tests

```sh
make test
make race
make bench
```

## 17.1 How to evaluate this repository in five minutes

Run the reviewer path:

```sh
make review
```

That validates the repository evidence pack, formatting, vet, tests, race
checks, benchmark compilation, build packaging, and the Go vulnerability scan.

## 18. Failure scenarios

- Corrupted AOF records are skipped and counted during replay.
- Partial trailing AOF records are treated as interrupted writes and ignored.
- Expired keys may exist physically until lazy read or cleanup removes them.
- Admin pprof must not be exposed on an untrusted interface.
- High connection counts are bounded by OS limits and goroutine scheduling.

Runbooks live under [docs/runbooks](docs/runbooks).

## 19. Roadmap

- RESP parser compatibility.
- AOF compaction after snapshots.
- Optional sharded map for write-heavy workloads.
- Connection limits and ACLs for exposed deployments.
- Replication protocol for warm standby nodes.
