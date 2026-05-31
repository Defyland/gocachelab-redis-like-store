# Implementation Plan

## Scope

Build a runnable Redis-like cache node in Go with TCP commands, TTL,
checksummed AOF durability, snapshot save/restore, admin observability,
concurrency tests, failure tests, native benchmarks, and the documentation pack
required by the shared project spec.

## Files to Create or Update

- Runtime: `cmd/gocachelab`, `internal/protocol`, `internal/store`,
  `internal/aof`, `internal/server`, `internal/metrics`.
- Verification: `*_test.go`, `*_bench_test.go`, `.github/workflows/ci.yml`,
  `scripts/validate_repository.sh`.
- Docs: `README.md`, `openapi.yaml`, `docs/**`, `benchmarks/**`,
  `deploy/grafana/gocachelab-dashboard.json`.

## Acceptance Criteria Mapping

| Acceptance criterion | Planned implementation | Verification |
| --- | --- | --- |
| Concurrent TCP clients | Goroutine per accepted connection, shared store guarded by `sync.RWMutex` | TCP concurrency test with 1000 clients and `go test -race ./...` |
| Command protocol | Inline parser with quoted arguments and Redis-like responses | Parser and server tests |
| TTL semantics | Lazy expiration on access plus background cleanup ticker | Store TTL tests and cleanup race tests |
| Durable recovery | AOF records with length and CRC, replay report for partial/corrupt records | AOF unit and failure tests |
| Snapshot `SAVE` | JSON snapshot written atomically through temp file rename | Snapshot restore tests |
| Observability | `INFO`, `/metrics`, `/healthz`, `/readyz`, `/debug/pprof` | Server tests and OpenAPI docs |
| Benchmarks | Native Go benchmarks for required scenarios | `go test -bench=. -run '^$' ./...` |
| Senior docs | Product, domain, architecture, ADRs, runbooks, security, case study | Repository validation script |

## Verification Commands

```sh
gofmt -w cmd internal
go test ./...
go test -race ./...
go test -bench=. -run '^$' -benchtime=1x ./...
go vet ./...
./scripts/validate_repository.sh
```

## Risks

- A single `RWMutex` is intentionally simple but can bottleneck mixed write
  workloads.
- AOF uses a write-ahead append before in-memory mutation so append failures do
  not acknowledge or apply writes. A process exit after append but before the
  response can replay a command the client did not observe; `fsync=always`
  narrows host-crash loss at higher SET latency.
- Background TTL cleanup is batch-based and may lag under large expiry storms.
- The simple protocol is not RESP-compatible, so existing Redis clients cannot
  be reused yet.

## Deferred Work

- RESP parser and compatibility tests.
- AOF compaction and snapshot-aware truncation.
- Authentication, ACLs, and connection quotas.
- Sharded map implementation for high write concurrency.
- Replication and promoted standby recovery.
