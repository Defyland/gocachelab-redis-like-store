# Runbook - High Memory Usage

## Signals

- `gocachelab_physical_keys` grows faster than `gocachelab_keys`.
- `gocachelab_alloc_bytes` keeps increasing.
- pprof heap profile shows key/value retention.

## Triage

1. Run `INFO` and compare `keys`, `physical_keys`, and `expired_keys_total`.
2. Check whether many keys have short TTLs and cleanup is lagging.
3. Capture `/debug/pprof/heap` from the trusted admin listener.
4. Inspect client behavior for broad `KEYS` usage or write bursts.

## Mitigation

- Lower `GOCACHELAB_TTL_CLEANUP_INTERVAL`.
- Raise `GOCACHELAB_TTL_CLEANUP_BATCH` if CPU headroom exists.
- Restart from snapshot and AOF only after preserving evidence.
- Add memory alerts before the process approaches host limits.

## Follow-up

Consider maxmemory eviction or a min-heap TTL index only after measuring cleanup
lag and lock contention.

