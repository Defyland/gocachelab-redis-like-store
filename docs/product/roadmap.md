# Roadmap

## v0.1

- Inline TCP protocol.
- String key-value entries.
- TTL with lazy and background expiration.
- AOF replay with corrupted and partial record handling.
- Snapshot `SAVE`.
- `INFO`, Prometheus metrics, pprof, tests, and benchmarks.

## v0.2

- RESP parser with compatibility tests.
- AOF compaction after snapshot.
- Connection quotas and max command rate per source address.
- Better latency histograms for command classes.

## v0.3

- Sharded map implementation behind the same store interface.
- Replica replay stream for warm standby nodes.
- ACLs for deployments beyond trusted local networks.

