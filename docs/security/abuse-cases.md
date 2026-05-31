# Abuse Cases

## Command Flood

Impact: CPU saturation, goroutine growth, and lock contention.

Controls: private network binding, OS limits, future connection quotas.

## Large Command Payload

Impact: memory pressure while reading client input.

Controls: `GOCACHELAB_MAX_LINE_BYTES` and command parser validation.

## Unbounded Key Scan

Impact: `KEYS` scan holds the store lock and can raise latency for other
clients.

Controls: bounded `KEYS` limit and runbook guidance.

## AOF Disk Fill

Impact: mutating commands fail or recovery becomes slow.

Controls: disk monitoring and future AOF compaction.

## Public pprof Exposure

Impact: runtime data disclosure and profiling overhead.

Controls: bind admin listener to loopback or private admin network only.

