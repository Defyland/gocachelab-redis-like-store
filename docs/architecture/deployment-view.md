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

