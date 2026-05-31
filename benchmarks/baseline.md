# Benchmark Baseline

## Method

Native Go benchmarks cover the implemented data path. The admin HTTP k6 script
checks only `/healthz` and `/metrics` because cache commands are TCP.

Recommended full run:

```sh
go test -bench=. -run '^$' -benchmem ./...
```

CI smoke run:

```sh
go test -bench=. -run '^$' -benchtime=1x ./...
```

## Scenarios

| Scenario | Benchmark |
| --- | --- |
| SET path | `BenchmarkStoreSet1MKeys` |
| GET on 1M-key dataset | `BenchmarkStoreGetMillionKeyDataset` |
| Mixed 80/20 workload | `BenchmarkStoreMixed80Read20Write` |
| 100 concurrent logical clients | `BenchmarkStore100ConcurrentClients` |
| TTL cleanup | `BenchmarkTTLCleanup` |
| AOF replay | `BenchmarkAOFReplay100kRecords` |
| TCP SET/GET percentiles | `BenchmarkTCPSetGetLatencyPercentiles` |

## Local Targets

| Operation | Target | Notes |
| --- | --- | --- |
| GET p95 | < 1 ms | No broad `KEYS` scan competing for the lock |
| SET p95 | < 2 ms | `fsync=never` or `everysec` |
| Race detector | Clean | `go test -race ./...` |

## Current Evidence

Smoke benchmark run on 2026-05-31, Apple M1 Max, `darwin/arm64`,
`go test -bench=. -run '^$' -benchtime=1x ./...`:

| Metric | Result |
| --- | --- |
| TCP GET p50 | 50 us |
| TCP GET p95 | 155 us |
| TCP GET p99 | 776 us |
| TCP SET p50 | 50 us |
| TCP SET p95 | 145 us |
| TCP SET p99 | 1,619 us |
| AOF replay 100k records | 22.62 ms |
| TTL cleanup batch | 145.92 us |
| 100 concurrent logical clients | 104.58 us |

These numbers are smoke evidence, not capacity-planning numbers. Longer
benchmarks with representative value sizes should be stored under
`benchmarks/results/` before using the results for operational sizing.
