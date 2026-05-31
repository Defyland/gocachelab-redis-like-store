# Aggregates

## Key Entry Aggregate

Fields:

- `key`: command-addressable string identifier.
- `value`: stored string payload.
- `expires_at`: optional absolute timestamp.

Commands:

- `SET` creates or replaces the aggregate.
- `GET`, `EXISTS`, `TTL`, and `KEYS` observe live state.
- `EXPIRE` changes the expiration boundary.
- `PERSIST` removes the expiration boundary.
- `DEL` deletes the aggregate.

Consistency boundary: a single command holds the store lock for the mutation or
for lazy cleanup needed by that observation.

## AOF Record Aggregate

Fields:

- Header prefix and payload length.
- CRC32 checksum.
- Inline durable command payload.

The AOF record is append-only. Replay applies supported durable commands in
order and reports records that cannot be trusted.

