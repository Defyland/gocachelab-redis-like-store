# Benchmark Methodology

Benchmarks are native Go benchmarks because the main data API is TCP and the
store hot path is inside the Go process. k6 is included only as an admin HTTP
smoke check for `/healthz` and `/metrics`.

## Scenarios

- `BenchmarkStoreSet1MKeys`: write-heavy SET path.
- `BenchmarkStoreGetMillionKeyDataset`: GET path with one million preloaded keys.
- `BenchmarkStoreMixed80Read20Write`: mixed workload with 80 percent reads.
- `BenchmarkStore100ConcurrentClients`: 100 concurrent logical clients against
  the store API.
- `BenchmarkTTLCleanup`: expired-key cleanup batches.
- `BenchmarkAOFReplay100kRecords`: replay throughput for durable records.

## Commands

```sh
go test -bench=. -run '^$' -benchmem ./...
go test -race ./...
```

For smoke verification in CI:

```sh
go test -bench=. -run '^$' -benchtime=1x ./...
```

## Local Targets

- GET p95 under 1 ms on a local machine without broad `KEYS` scans.
- SET p95 under 2 ms with `GOCACHELAB_AOF_FSYNC=never` or `everysec`.
- Race detector clean for command, TTL, and cleanup concurrency tests.

The repo records measured results in `benchmarks/results/` when full benchmark
runs are executed.

