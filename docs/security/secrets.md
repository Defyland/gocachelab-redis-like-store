# Secrets

Gocachelab has no required external secrets in v1.

## Environment Variables

Configuration is passed through environment variables for addresses, data paths,
cleanup settings, and AOF fsync policy. These variables do not need secret
storage unless paths or deployment metadata are considered sensitive.

## Cached Values

The service treats values as opaque. Clients must not store long-lived secrets
unless filesystem permissions, backup encryption, and memory inspection risks are
accepted.

## Logging Policy

Structured logs may include subsystem, operation, address, and error fields.
They must not include cached values.

