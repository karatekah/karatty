# ADR 0002 — Invitations and possession proof

**English** · [Español](../es/adr/0002-authentication.md) · [Italiano](../it/adr/0002-authentication.md)

Accepted, 2026-10-31. Historical invitation-prototype decision.

Registration/login signs a random, single-use, short-lived challenge with the TERMLINK authentication domain and a public key bound in PostgreSQL. That legacy signing domain retains its spelling for protocol compatibility. An SSH signature is checked before consuming an invitation and registering the key. The hashed invitation has an expiry and remaining uses and is consumed under a lock in the same transaction. Private keys are never sent to the server.

SSH proves possession at handshake and accepts only previously associated keys. The prototype allows invitation signup followed by SSH access without installing a client; invitations can grant initial channel membership. This remote SSH flow is excluded from the alpha. Administration uses local database access rather than a public administration endpoint.

Random session tokens stay in client memory; the server stores hashes. Token files are not saved. Encrypted keys unlock through hidden terminal input and depend on filesystem permissions. Later work added email-free pseudonymous signup and self-service recovery using a one-use code; see [architecture](../architecture.md) and [protocol](../protocol.md). OS credential-store integration, multiple devices and assisted rotation remain future work.
