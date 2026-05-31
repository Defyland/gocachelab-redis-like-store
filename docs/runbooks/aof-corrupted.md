# Runbook - AOF Corrupted

## Signals

- `aof_replay_corrupted_total` is non-zero in `INFO` or `/metrics`.
- Startup logs report replay counters that differ from the expected write count.

## Triage

1. Stop writers to prevent more records from being appended.
2. Copy `appendonly.aof` and `snapshot.json` before manual inspection.
3. Check whether corruption is isolated to one record or repeated.
4. Compare recovered key count against the last known benchmark or smoke output.

## Mitigation

- If corruption is a trailing interrupted write, the partial counter should be
  non-zero and recovery can usually continue.
- If corruption is in the middle of the file, restore from a trusted backup or
  accept that later records may have replayed over an uncertain base state.
- Run `SAVE` after recovery only when the recovered state is trusted.

## Prevention

- Use reliable storage.
- Use `GOCACHELAB_AOF_FSYNC=always` for durability experiments where write
  latency is less important than crash-loss reduction.
- Add AOF compaction and backup verification in the next durability milestone.

