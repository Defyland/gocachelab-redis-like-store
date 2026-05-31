# ADR 0006 - Start With a Simple Inline Protocol

## Status

Accepted

## Context

RESP compatibility is useful, but the first value of this repo is the cache-node
mechanics around command execution, TTL, AOF, and concurrency.

## Options Considered

1. Full RESP parser first.
2. Inline protocol only.
3. Inline protocol with explicit RESP rejection.

## Decision

Start with an inline protocol and reject RESP input explicitly.

## Consequences

Positive:

- Parser is small enough to audit.
- Tests focus on command semantics and failure handling.
- RESP can be added later behind the same command model.

Negative:

- Existing Redis clients cannot connect directly.
- Multi-bulk binary-safe values are not supported in v1.

