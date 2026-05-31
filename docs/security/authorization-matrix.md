# Authorization Matrix

Gocachelab v1 has no user identity model. Access control is delegated to the
deployment boundary.

| Surface | Intended actor | Control in v1 | Required hardening for exposed deployments |
| --- | --- | --- | --- |
| TCP cache commands | Trusted application or developer | Bind to loopback/private network | ACLs, TLS, command quotas |
| `/healthz` and `/readyz` | Supervisor or load balancer | Admin listener binding | Network policy |
| `/metrics` | Prometheus scraper | Admin listener binding | Network policy and scrape auth |
| `/debug/pprof/*` | Operator | Admin listener binding | Strong auth, temporary enablement, audit |
| AOF and snapshot files | Process user | File mode `0600` for AOF | Encrypted disk, backup policy |

The absence of in-process auth is explicit, not accidental. It keeps the first
implementation focused on cache-node mechanics and makes the deployment boundary
mandatory in the runbooks.

