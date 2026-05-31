#!/usr/bin/env sh
set -eu

required_files='
README.md
openapi.yaml
Dockerfile
compose.yaml
go.mod
docs/engineering-case-study.md
docs/spec-driven/senior-readiness-spec.md
docs/spec-driven/implementation-plan.md
docs/spec-driven/verification-report.md
docs/product/problem.md
docs/product/personas.md
docs/product/use-cases.md
docs/product/non-goals.md
docs/product/roadmap.md
docs/domain/glossary.md
docs/domain/bounded-contexts.md
docs/domain/aggregates.md
docs/domain/invariants.md
docs/domain/state-machines.md
docs/api/README.md
docs/events/README.md
docs/security/threat-model.md
docs/scalability.md
docs/operational-cost.md
benchmarks/baseline.md
'

required_dirs='
docs/spec-driven
docs/product
docs/domain
docs/adr
docs/architecture
docs/benchmarks
docs/api
docs/diagrams
docs/events
docs/runbooks
docs/security
benchmarks/results
'

for path in $required_dirs; do
  test -d "$path" || {
    echo "missing directory: $path" >&2
    exit 1
  }
done

for path in $required_files; do
  test -f "$path" || {
    echo "missing file: $path" >&2
    exit 1
  }
done

for heading in \
  'What is this product?' \
  'Problem it solves' \
  'Target users' \
  'Main features' \
  'Architecture overview' \
  'Testing strategy' \
  'Performance benchmarks' \
  'Observability' \
  'Security considerations' \
  'Roadmap'
do
  grep -q "$heading" README.md || {
    echo "README missing heading: $heading" >&2
    exit 1
  }
done

for adr in \
  docs/adr/0001-build-the-store-in-go.md \
  docs/adr/0002-use-rwmutex-for-v1-concurrency.md \
  docs/adr/0003-use-lazy-and-background-ttl-expiration.md \
  docs/adr/0004-use-aof-durability-before-replication.md \
  docs/adr/0005-expose-pprof-and-prometheus-metrics.md \
  docs/adr/0006-start-with-simple-inline-protocol.md
do
  test -f "$adr" || {
    echo "missing ADR: $adr" >&2
    exit 1
  }
done

if grep -R "TODO\\|TBD\\|production-ready" README.md docs benchmarks 2>/dev/null; then
  echo "docs contain forbidden placeholder or unevidenced maturity language" >&2
  exit 1
fi

echo "repository structure validation passed"

