# Senior Readiness Spec

## Product Bar

Gocachelab must read as a real cache-node product: a local Redis-like service
with clear users, workflows, constraints, and non-goals.

## Domain Bar

The repo must define the key-value aggregate, TTL state, persistence records,
and invariants that keep expired keys from being observed by clients.

## Architecture Bar

The architecture must justify a single-process Go TCP server, RWMutex-protected
map, AOF replay, background TTL cleanup, admin HTTP endpoints, and deferred RESP
compatibility.

## API Bar

The TCP command protocol must be documented with request, response, and failure
examples. The admin HTTP endpoints must have OpenAPI documentation.

## Data and Consistency Bar

The consistency model must explain in-memory source of truth, AOF durability
window, snapshot restore order, TTL clock assumptions, and recovery behavior for
corrupted or partial AOF records.

## Security Bar

The repo must document trusted-network assumptions, lack of auth, admin endpoint
risk, command size validation, value logging policy, filesystem permissions, and
abuse cases.

## Observability Bar

The service must expose `INFO`, Prometheus text metrics, structured logs,
health/readiness endpoints, and pprof.

## Performance Bar

Benchmarks must cover SET/GET with large key counts, mixed 80/20 workloads, 100
concurrent clients, TTL cleanup, and AOF replay. Local p95 targets are GET under
1 ms and SET under 2 ms when `fsync=never` or `everysec`.

## Scalability Bar

Docs must name hot paths, lock contention, map growth, AOF replay time, TTL
cleanup pressure, and the point where sharding or replication becomes necessary.

## Operational Cost Bar

Docs must compare one-process operation with the cost of adding Redis,
replication, sharding, external metrics systems, and managed storage.

## Maintainability Bar

Modules must have narrow ownership: protocol parsing, storage, persistence,
serving, and metrics. Tests must express domain behavior rather than only happy
paths.

## Readability Bar

Code and docs must use cache-domain language: key entry, expiration, replay,
snapshot, cleanup lag, durable command, and recovery report.

## Test and CI Bar

The repo must include native tests, race detector coverage, benchmark smoke
commands, vet, formatting checks, Docker build validation, and repository
structure validation.

## Evidence Matrix

| Criterion | Evidence | Status | Notes |
| --- | --- | --- | --- |
| Product problem and users are explicit | `README.md`, `docs/product/problem.md`, `docs/product/personas.md` | Done | Names product workflow, users, value, and non-goals. |
| TCP protocol commands are implemented | `internal/server/tcp.go`, `internal/server/tcp_test.go`, `docs/api/README.md` | Done | Includes command errors and bounded `KEYS`. |
| Store is thread-safe with documented invariants | `internal/store/store.go`, `internal/store/store_test.go`, `docs/domain/invariants.md` | Done | Race detector passed for concurrency tests. |
| TTL uses lazy and background expiration | `internal/store/store.go`, `cmd/gocachelab/main.go`, `docs/adr/0003-use-lazy-and-background-ttl-expiration.md` | Done | Cleanup lag trade-off documented. |
| AOF handles corrupted and partial records | `internal/aof/aof.go`, `internal/aof/aof_test.go` | Done | Replay report distinguishes applied, corrupted, and partial records. |
| Snapshot restore is tested | `internal/store/snapshot.go`, `internal/store/snapshot_test.go`, `internal/server/tcp_test.go` | Done | `SAVE` writes JSON snapshot atomically. |
| Metrics and pprof are exposed | `internal/server/admin.go`, `internal/metrics/metrics_test.go`, `deploy/grafana/gocachelab-dashboard.json`, `openapi.yaml` | Done | Admin listener exposes health, readiness, metrics, and pprof. |
| Benchmarks cover required scenarios | `internal/store/store_bench_test.go`, `internal/aof/aof_bench_test.go`, `internal/server/tcp_bench_test.go`, `benchmarks/results/2026-05-31-smoke.md` | Done | TCP GET p95 155 us and SET p95 145 us in smoke run. |
| Failure scenarios are tested | `internal/server/tcp_test.go`, `internal/aof/aof_test.go`, `internal/store/store_test.go` | Done | Includes invalid command, disconnect, corrupt AOF, partial AOF, and cleanup races. |
| CI runs quality gates | `.github/workflows/ci.yml`, `scripts/validate_repository.sh` | Done | CI includes format, vet, tests, race, coverage, benchmark smoke, Docker, and docs validation. |

## Out of Scope

- RESP compatibility in the first implementation.
- Cluster mode, replication, and sharding.
- Authentication and ACLs for public exposure.
- AOF compaction after snapshot.
- Eviction policies such as LRU/LFU.
