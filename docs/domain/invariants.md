# Invariants

1. Public reads must not return expired keys.
2. An expired key may remain as a physical key until lazy expiration or cleanup.
3. A persistent key has no expiration timestamp and returns `TTL = -1`.
4. A missing or expired key returns `TTL = -2`.
5. `SET` replaces both value and TTL state.
6. `EXPIRE` only succeeds for a currently live key.
7. AOF TTL records use absolute timestamps so replay does not extend TTLs.
8. Snapshot files contain only keys live at snapshot time.
9. Store map access is protected by one `sync.RWMutex`.
10. Values are never written to structured logs.

Tests covering these invariants:

- `internal/store/store_test.go`
- `internal/aof/aof_test.go`
- `internal/server/tcp_test.go`

