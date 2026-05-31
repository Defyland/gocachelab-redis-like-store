# Scalability

## Hot Paths

- `GET`: map lookup, TTL check, optional lazy deletion.
- `SET`: AOF append, map write, optional expiration metadata.
- `KEYS`: scan of the keyspace with optional lazy deletion.
- Startup replay: sequential AOF decode and command application.

## Read-Heavy Operations

`GET`, `EXISTS`, and `TTL` dominate read-heavy workloads. `GET` can use the read
lock when the key is live, but expired keys require deletion.

## Write-Heavy Operations

`SET`, `DEL`, `EXPIRE`, `PERSIST`, lazy expiration, and cleanup batches contend
on the same store lock.

## Fastest-Growing Data

The in-memory map and AOF file grow fastest. Snapshot size grows with live keys.
AOF size grows with every mutating command until compaction is implemented.

## Hot Partitions

A single key can become hot if many clients update it. With one process and one
lock, the first bottleneck is lock contention rather than a network partition.

## Horizontal Scaling

The current service does not scale horizontally because there is no replication
or consistent hashing layer. Horizontal scale requires partitioning keys across
nodes or adding a replica protocol with promotion rules.

## Sharding Path

The lowest-risk next step is an internal sharded map behind the same store API.
That reduces lock contention but complicates `KEYS`, snapshotting, and cleanup.

## What Can Be Asynchronous

- AOF compaction.
- Snapshot compression.
- Metrics export aggregation.
- TTL cleanup batches.

## What Must Not Be Eventual

- A single command's visible result on one node.
- Expired-key invisibility for public reads.
- Replay order within one AOF.

