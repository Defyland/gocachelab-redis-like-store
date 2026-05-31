# Data Classification

| Data | Classification | Handling |
| --- | --- | --- |
| Keys | Internal metadata | Visible through `KEYS`, metrics only expose counts |
| Values | Potentially sensitive application data | Stored in memory, AOF, and snapshot; never logged |
| AOF records | Sensitive operational data | Local file, `0600`, backup only on trusted storage |
| Snapshot JSON | Sensitive operational data | Atomic local write, protect with filesystem policy |
| pprof profiles | Sensitive runtime data | Admin-only access |
| Metrics | Low to medium sensitivity | Exposes counts and errors, not values |

Operators should assume values may contain secrets because the cache does not
inspect payload semantics.

