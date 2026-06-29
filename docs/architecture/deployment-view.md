# Deployment View

## Local Binary

```text
go run ./cmd/gocachelab
```

Data lands in `./data/appendonly.aof` and `./data/snapshot.json`.

## Docker Compose

```text
docker compose up --build
```

Ports:

- `7379`: TCP command protocol.
- `8080`: admin HTTP.

Volume:

- `gocachelab-data`: AOF and snapshot files.

## Trust Boundary

Both listeners are intended for trusted local or private network use. The admin
listener exposes pprof and must be blocked from untrusted clients.

## Hosted demo intentionally omitted

This repository does not provide a Railway demo. As an R&D asset for studying
cache internals, a public hosted deploy would hide the intended private-network
trust boundary, expose an unauthenticated TCP cache interface, and weaken the
durability story if the runtime does not offer the same local-disk semantics
used by the AOF and snapshot files.

The truthful runnable surfaces are:

- the local developer binary on a trusted machine;
- Docker Compose on a private network with the named data volume;
- a future infrastructure deployment only when it preserves private networking,
  persistent storage, and restricted admin access.
