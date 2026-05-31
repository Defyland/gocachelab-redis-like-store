# Runbook - Too Many Connections

## Signals

- `connected_clients` rises unexpectedly.
- New clients fail to connect.
- OS logs show file descriptor exhaustion.

## Triage

1. Run `INFO` to check active and accepted connection counters.
2. Inspect client retry behavior.
3. Check process file descriptor limits.
4. Confirm clients send `QUIT` or close idle sockets.

## Mitigation

- Restart abusive clients before restarting the cache node.
- Raise OS file descriptor limits for controlled benchmark runs.
- Bind the TCP listener to a private interface.

## Follow-up

Add in-process connection limits and idle deadlines before deploying outside a
local lab.

