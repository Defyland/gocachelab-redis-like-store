# ADR 0003 - Use Lazy and Background TTL Expiration

## Status

Accepted

## Context

TTL keys must disappear from client-visible state when expired, but immediate
per-key timers would add scheduling overhead and complexity for large keysets.

## Options Considered

1. Lazy expiration only.
2. Per-key timers.
3. Min-heap of expirations.
4. Lazy expiration plus background cleanup batches.

## Decision

Use lazy expiration on public observations and background cleanup batches.

## Consequences

Positive:

- Public reads never return expired keys.
- Cleanup cost is bounded per tick.
- No heap bookkeeping in the first implementation.

Negative:

- Expired physical keys can consume memory until observed or cleaned.
- Expiry storms can create cleanup lag.
- `KEYS` may pay cleanup cost while scanning.

