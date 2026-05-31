# Use Cases

## Store a Session-Like Value

```text
SET session:123 user-42 EX 900
GET session:123
TTL session:123
```

The key is visible until expiration. Expiration is absolute in the AOF so replay
does not extend the session after restart.

## Cache a Stable Configuration Value

```text
SET config:feature-a enabled
PERSIST config:feature-a
GET config:feature-a
```

Persistent keys have `TTL = -1` and only leave the store through `DEL`.

## Diagnose a Local Latency Spike

1. Run `INFO` over TCP to inspect connected clients and command errors.
2. Scrape `/metrics` to compare key count, expired key count, and AOF errors.
3. Capture `/debug/pprof/profile?seconds=30` on a trusted admin interface.
4. Follow the high-latency runbook before changing the locking strategy.

## Recover After Process Restart

1. Load the most recent JSON snapshot when present.
2. Replay AOF records in order.
3. Skip corrupted records and ignore a partial trailing record.
4. Surface replay counters in `INFO` and Prometheus metrics.

