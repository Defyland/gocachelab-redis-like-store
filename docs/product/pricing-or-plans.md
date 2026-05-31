# Pricing or Plans

Gocachelab is not sold as a hosted product. The relevant "plan" decision is
operational footprint:

| Plan | Intended use | Cost shape |
| --- | --- | --- |
| Local process | Portfolio review, learning, benchmark experiments | One Go binary and local disk |
| Docker Compose | Repeatable demos and smoke tests | One container and one named volume |
| Hardened private node | Internal lab service only | Adds network policy, process supervision, backup, and metrics scraping |

Managed Redis remains the right plan when teams need mature replication,
eviction, memory tuning, client ecosystem support, or service-level guarantees.

