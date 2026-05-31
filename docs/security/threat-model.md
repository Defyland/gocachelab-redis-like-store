# Threat Model

## Assets

- Cached values in memory.
- AOF and snapshot files on disk.
- Admin pprof profiles and metrics.
- Availability of the TCP listener.

## Actors

- Trusted local developer or operator.
- Internal service client on a private network.
- Untrusted network client if the service is misbound to a public interface.
- Local user with filesystem access to the data directory.

## Trust Boundaries

- TCP and admin listeners are trusted-network surfaces.
- The local filesystem is trusted only to the Unix user running the process.
- pprof is operator-only and must not be routed publicly.

## Abuse Cases

- Send very large commands to consume memory.
- Use `KEYS *` repeatedly to hold the store lock.
- Expose pprof and leak heap or CPU profile data.
- Corrupt AOF files to create recovery uncertainty.
- Fill disk by issuing many mutating commands.

## Controls

- Command line size limit through `GOCACHELAB_MAX_LINE_BYTES`.
- Bounded `KEYS` through `GOCACHELAB_KEYS_LIMIT`.
- AOF files created with `0600` permissions.
- Values are not emitted in structured logs.
- Admin endpoints are separate from the TCP listener.
- Corrupted and partial AOF records are counted during replay.

## Residual Risks

- No authentication or ACLs in v1.
- No per-client rate limiting.
- No encryption for TCP traffic.
- No maxmemory eviction policy.
- No AOF compaction, so disk usage grows with write volume.

