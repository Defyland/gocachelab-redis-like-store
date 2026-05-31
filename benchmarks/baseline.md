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

## Local Targets

| Operation | Target | Notes |
| --- | --- | --- |
| GET p95 | < 1 ms | No broad `KEYS` scan competing for the lock |
| SET p95 | < 2 ms | `fsync=never` or `everysec` |
| Race detector | Clean | `go test -race ./...` |

## Current Evidence

The verification report records the commands run in this implementation pass.
Full p50/p95/p99 latency export is deferred to a dedicated benchmark run because
Go benchmark output reports ns/op rather than percentile histograms.

