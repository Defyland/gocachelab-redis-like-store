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
| Product problem and users are explicit | `README.md`, `docs/product/problem.md`, `docs/product/personas.md` | Planned | Filled during documentation pass. |
| TCP protocol commands are implemented | `internal/server`, `docs/api/README.md` | Planned | Must include errors and bounded `KEYS`. |
| Store is thread-safe with documented invariants | `internal/store`, `docs/domain/invariants.md` | Planned | Must include concurrency tests. |
| TTL uses lazy and background expiration | `internal/store`, `docs/adr/0003-use-lazy-and-background-ttl-expiration.md` | Planned | Must document cleanup lag trade-off. |
| AOF handles corrupted and partial records | `internal/aof`, `internal/aof/*_test.go` | Planned | Replay report must distinguish both. |
| Snapshot restore is tested | `internal/store`, `internal/server` tests | Planned | Optional feature included because `SAVE` is in scope. |
| Metrics and pprof are exposed | `internal/server/admin.go`, `deploy/grafana/gocachelab-dashboard.json` | Planned | Admin listener only. |
| Benchmarks cover required scenarios | `internal/*_bench_test.go`, `benchmarks/baseline.md` | Planned | Full numbers depend on local run. |
| Failure scenarios are tested | `internal/server`, `internal/aof` tests | Planned | Includes invalid command and disconnect. |
| CI runs quality gates | `.github/workflows/ci.yml` | Planned | Uses Go native tooling. |

## Out of Scope

- RESP compatibility in the first implementation.
- Cluster mode, replication, and sharding.
- Authentication and ACLs for public exposure.
- AOF compaction after snapshot.
- Eviction policies such as LRU/LFU.

