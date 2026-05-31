# Runbook - TTL Cleanup Lag

## Signals

- `physical_keys` is much larger than `keys`.
- Memory grows after a large expiry wave.
- GET latency rises because lazy expiration performs deletion work.

## Triage

1. Run `INFO` before and after the cleanup interval.
2. Compare `expired_keys_total` movement with expected TTL churn.
3. Check cleanup configuration: interval and batch size.
4. Capture CPU profile if cleanup appears to dominate runtime.

## Mitigation

- Increase `GOCACHELAB_TTL_CLEANUP_BATCH`.
- Decrease `GOCACHELAB_TTL_CLEANUP_INTERVAL`.
- Avoid synchronized TTLs in client workloads when possible.
- Use `PERSIST` for keys that should not churn.

## Follow-up

If cleanup remains a bottleneck, evaluate a min-heap expiration index or sharded
cleanup workers. Both add bookkeeping cost and should be justified by measured
lag.

