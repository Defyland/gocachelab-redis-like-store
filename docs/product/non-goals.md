# Non-Goals

- Replacing Redis in real deployments.
- Wire compatibility with existing Redis clients in v1.
- Cluster membership, replication, failover, or sharding.
- Eviction policies such as LRU, LFU, volatile-only eviction, or maxmemory.
- Authentication and ACLs for internet-facing deployments.
- Multi-type data structures such as lists, sets, streams, and sorted sets.
- Strong fsync-on-every-command durability as the default latency profile.

These are intentionally excluded to keep the first version focused on the
systems mechanics that can be tested and explained in a compact codebase.

