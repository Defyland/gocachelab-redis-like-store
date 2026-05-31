# Glossary

| Term | Meaning |
| --- | --- |
| Key entry | A string key, string value, and optional expiration timestamp. |
| Live key | A key whose expiration is absent or still in the future. |
| Physical key | A key still present in the in-memory map, including expired keys awaiting cleanup. |
| Lazy expiration | Removing an expired key when a command tries to observe it. |
| Background cleanup | Periodic batch deletion of expired physical keys. |
| AOF record | A checksummed durable command payload used for replay. |
| Partial record | A trailing AOF record cut off during an interrupted write. |
| Corrupted record | A record with invalid header, checksum, payload, or replay command. |
| Snapshot | JSON dump of live entries written atomically by `SAVE`. |
| Cleanup lag | Time between expiration and physical deletion. |

