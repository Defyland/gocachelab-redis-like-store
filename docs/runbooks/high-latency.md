# Runbook - High Latency

## Signals

- Local benchmark p95 exceeds GET 1 ms or SET 2 ms with `fsync=never/everysec`.
- Clients report slow responses under mixed workloads.
- CPU profile shows lock contention, AOF sync, or broad key scans.

## Triage

1. Confirm `GOCACHELAB_AOF_FSYNC`; `always` is expected to raise SET latency.
2. Run `INFO` and check active clients, command errors, and key counts.
3. Capture `/debug/pprof/profile?seconds=30`.
4. Inspect whether clients are issuing `KEYS *` or very large values.
5. Compare latency with AOF disabled in a controlled local run.

## Mitigation

- Use `everysec` instead of `always` when durability policy allows.
- Bound or remove broad `KEYS` usage.
- Tune TTL cleanup batch size if cleanup appears in profiles.
- Move benchmark data to faster local storage if AOF writes dominate.

## Follow-up

If lock contention dominates after workload cleanup, prototype a sharded store
behind the existing store API and compare race-detector and benchmark results.

