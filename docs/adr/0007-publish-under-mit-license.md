# ADR 0007 - Publish the Repository Under the MIT License

## Status

Accepted

## Context

GoCacheLab is positioned as a didactic systems repository. The parser, storage
engine, benchmarks, and runbooks are intentionally documented for external
study, but no explicit license means the repo is technically "look, don't
touch."

## Options Considered

1. Keep the default all-rights-reserved posture.
2. Publish under the MIT License.
3. Publish under a reciprocal or commercially restrictive license.

## Decision

Publish the repository under the MIT License and state that clearly in the
README.

## Consequences

Positive:

- Readers can fork and adapt the implementation with a well-understood reuse
  boundary.
- The repo becomes a practical learning asset instead of a documentation-only
  reference.

Negative:

- Downstream forks may diverge without upstream contribution.
- License clarity raises the bar on keeping third-party notices accurate.
