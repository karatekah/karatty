# ADR 0001 — Shared core and PostgreSQL

**English** · [Español](../es/adr/0001-foundation.md) · [Italiano](../it/adr/0001-foundation.md)

Accepted, 2026-10-31. This records the original foundation decision; E2EE identity/key tables were added later under [ADR 0003](0003-e2ee-1to1.md).

One Go module builds an API, SSH gateway and client. SSH verifies key possession and creates a session against the same domain core. API/SSH authorize each operation; TUI components share a small interface.

PostgreSQL stores users, public keys, sessions, invitations, authentication challenges, conversations, memberships and messages. Unused device/read-receipt tables were not added to the initial sprint. Persistence ACKs use the message response/ID; read receipts and multiple devices remain future work.

A conversation lock assigns transactional sequences. Conversation/sender/client-ID uniqueness deduplicates retries while the message is retained, for 30 days. Different content under that ID conflicts within the window. Expiry removes the ID, making a later send new. Client clocks are not trusted.

Initial real time reads the durable log every 250 ms through WSS or a core adapter. Every batch rechecks session, key, account and membership. This closes the history/subscription race, supports several processes and avoids requiring Redis. The tradeoff is periodic database work and up to 250 ms latency; measure before adding LISTEN/NOTIFY.

Reads/operations retain shared authorization locks during transactions; revocation waits for active operations and subsequent batches fail. Revoked keys/sessions close subscriptions at the next poll. The gateway also reevaluates idle SSH sessions.

Separate users/messages/realtime packages are added only when they have distinct responsibilities. Sprint operations live in auth/chat/transport; empty directories are not created merely by convention.
