# Events and Durability Records

Gocachelab does not publish domain events, outbox messages, webhooks, or
analytics events. The only append-only stream is the internal AOF.

## AOF Envelope

Each record has:

- Prefix: `GCL-AOF-1`
- Payload byte length
- CRC32 checksum
- Inline command payload

Example payload:

```text
SET session:1 token
EXPIREAT session:1 1780254900000000000
```

## Compatibility Rules

- New durable command names require replay support before writers emit them.
- Existing record header fields must not change without a new prefix.
- Partial trailing records are ignored because they represent interrupted writes.
- Corrupted records are counted and skipped; operators should inspect the AOF
  corruption runbook before trusting recovered state.

## Consumer Expectations

The only consumer is the local replay process during startup. It must be
idempotent for repeated `SET`, `DEL`, `EXPIREAT`, and `PERSIST` records.

