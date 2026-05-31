# Module Boundaries

| Module | Owns | Must not own |
| --- | --- | --- |
| `internal/protocol` | Inline parsing, quoting, response rendering | Store state or disk I/O |
| `internal/store` | Key entries, TTL, cleanup, snapshot materialization | TCP sockets or AOF headers |
| `internal/aof` | Record encoding, checksums, fsync policy, replay report | Command business rules beyond replay payload validity |
| `internal/server` | TCP connection lifecycle, command dispatch, admin HTTP | Raw map mutation without the store API |
| `internal/metrics` | Counters and text rendering for `INFO` and Prometheus | Command execution |
| `cmd/gocachelab` | Configuration, startup order, signal handling | Domain logic |

This boundary keeps future changes localized. RESP support belongs in
`internal/protocol`; sharded storage belongs behind `internal/store`; AOF
compaction belongs in `internal/aof`.

