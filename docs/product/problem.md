# Product Problem

Gocachelab exists to make cache-node engineering visible. Redis hides the hard
parts behind a mature implementation; most backend challenge repos hide them
behind HTTP CRUD. This project gives reviewers a small service where they can
inspect protocol parsing, shared memory, expiration, durability, recovery, and
observability in one place.

The product problem is not "replace Redis." The product problem is "teach and
demonstrate how a Redis-like node behaves under operational pressure."

## Core Workflow

1. An engineer starts the process locally or in Docker.
2. A TCP client sends cache commands such as `SET`, `GET`, `EXPIRE`, and `SAVE`.
3. The node serves concurrent clients through one shared store.
4. Mutating commands are appended to the AOF.
5. Expired keys are removed lazily on access and in cleanup batches.
6. Operators inspect `INFO`, `/metrics`, and pprof when latency or memory rises.

## Business Value

For a portfolio or hiring challenge, the value is evidence: the codebase shows
systems reasoning instead of only framework wiring. For a team, the value is a
readable reference implementation for cache-node trade-offs.

