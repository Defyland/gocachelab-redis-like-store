# API Documentation

## Transport

The cache data API is a TCP line protocol on `GOCACHELAB_TCP_ADDR`, default
`127.0.0.1:7379`. Each command is one line ending in `\r\n` or `\n`.

Arguments are space-separated. Double quotes support spaces and escaped
characters: `\"`, `\\`, `\n`, `\r`, and `\t`.

RESP is intentionally deferred. RESP input returns an error instead of being
silently misparsed.

## Response Format

The server uses Redis-like response prefixes:

| Prefix | Meaning | Example |
| --- | --- | --- |
| `+` | Simple string | `+OK` |
| `-` | Error | `-ERR unknown command NOPE` |
| `:` | Integer | `:1` |
| `$` | Bulk string | `$3\r\nAda` |
| `*` | Array of bulk strings | `*2 ...` |

## Commands

### PING

```text
PING
+PONG
```

`PING message` returns the message as a bulk string.

### SET

```text
SET user:1 "Ada Lovelace"
+OK

SET session:1 token EX 900
+OK
```

Optional TTL forms are `EX seconds` and `PX milliseconds`.

### GET

```text
GET user:1
$12
Ada Lovelace

GET missing
$-1
```

### DEL

```text
DEL user:1 session:1
:2
```

### EXISTS

```text
EXISTS user:1 missing
:1
```

### EXPIRE

```text
EXPIRE user:1 30
:1
```

Returns `0` when the key is missing or already expired.

### TTL

```text
TTL user:1
:30
TTL persistent:key
:-1
TTL missing
:-2
```

### PERSIST

```text
PERSIST user:1
:1
```

Returns `0` when the key is missing or has no expiration.

### KEYS

```text
KEYS user:* LIMIT 100
*1
$6
user:1
```

`KEYS` is intentionally bounded by `GOCACHELAB_KEYS_LIMIT`, default `1000`.

### INFO

Returns a bulk string with server, client, keyspace, command, persistence, and
memory counters.

### SAVE

Writes a JSON snapshot atomically to `GOCACHELAB_SNAPSHOT_PATH`.

### QUIT

Returns `+OK` and closes the connection.

## Error Examples

```text
GET
-ERR GET expects 1 argument

*1
-ERR RESP protocol is not supported by this build
```

## Idempotency Rules

`SET`, `DEL`, `EXPIRE`, and `PERSIST` are safe to replay from AOF because records
are applied in order and expiration is stored as an absolute timestamp.

