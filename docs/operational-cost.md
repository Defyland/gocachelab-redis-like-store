# Operational Cost

## Infrastructure Components

- One Go process.
- One TCP listener.
- One admin HTTP listener.
- One data directory with AOF and snapshot files.
- Optional Prometheus and Grafana for metrics.

## Non-Financial Cost

The main cost is operational attention: disk growth, memory growth, pprof
exposure, and interpreting recovery reports.

## Debugging Complexity

Debugging is local-first: `INFO`, `/metrics`, pprof, logs, AOF file inspection,
and snapshot inspection. The upside is a small surface. The downside is that
there is no managed service layer to absorb failures.

## Deployment Complexity

The binary is simple to run, but safe binding matters. Exposing either listener
to an untrusted network changes the threat model.

## Backup and Retention

AOF and snapshots must be backed up together. A snapshot without its following
AOF records can lose writes. An AOF without compaction increases storage cost
over time.

## Monitoring Burden

Minimum alerts:

- High `connected_clients`.
- Growing `physical_keys - keys` gap.
- Non-zero AOF corruption or partial replay counters.
- Disk usage in the data directory.
- Rising command error rate.

## Vendor Lock-in Risk

There is no vendor dependency. The trade-off is that mature Redis features must
be built or consciously excluded.

## Simpler Alternatives Rejected

- In-memory map only: no durability evidence.
- Snapshot only: loses recent writes and does not test replay.
- Managed Redis: solves the product problem but removes the engineering evidence.

