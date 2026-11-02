# Security and privacy

**English** · [Español](es/security.md) · [Italiano](it/security.md)

Required tests cover membership isolation, revocation, invitations, challenge signatures/reuse, sequence ordering, deduplication and persistence. Never log message bodies, tokens, invitations, signatures or private keys. Operational output is limited to generic startup/error logs and aggregate counters.

## Transport and limits

Plaintext prototype messages are limited to 4096 UTF-8 bytes, without controls or invisible formatting; E2EE envelopes to 12288 bytes. Pages contain at most 100 messages. Connections, handshakes, request/send rates and write times are bounded. Slow consumers disconnect and resynchronize against the durable log; the client bounds screen history.

TLS is required except explicitly enabled loopback-only HTTP development mode. Prototype SSH uses a persistent host key whose fingerprint must be verified; do not disable `StrictHostKeyChecking`. Listeners default to loopback.

The [verification matrix](verification.md) covers E2EE, identity loss/change, archives, transport limits and package extraction. It records Defender's server detection and the subsequent native Windows trial after the user restored the file, with protections active. This does not establish a false positive or validate other installations. Two-machine LAN testing is pending. Karatty has no independent audit.

Open registration uses a pseudonym and SSH key, without email. The local TUI encrypts direct chats with Olm after fingerprint verification. Invitation/remote-SSH modes lack E2EE and are excluded from the alpha. Future Internet hosting needs provider, backup and policy verification.

## Local state and recovery

Argon2id and XChaCha20-Poly1305 encrypt the archive with a local passphrase requested on opening and never sent to the server. Expired entries are removed when the client opens/saves; an offline device cannot run cleanup. After SSH-key recovery, the client migrates readable history into a new encrypted file and removes the old file only after saving/synchronizing the new device. Older sends lacking ACK remain uncertain for manual review. Cleanup while the program is not running cannot be guaranteed.

Reopening a direct chat requires both memberships and never restores revoked access. Recovery uses an offline code and a new SSH key. The encrypted client stores each pending envelope and client ID before transmission; retries preserve exactly that envelope. The legacy client only keeps pending sends in memory, so scripts must reuse `--id`.

## Process-level bounds

| Resource | Limit |
| --- | --- |
| API | 10 requests/s per IP, burst 30 |
| Sends | 2 messages/s per user, burst 10 |
| Prototype SSH | 1 authentication attempt/s per IP, burst 10 |
| Rate-limiter identities | 4096 |
| HTTP connections / SSH connections | 128 / 64 |
| SSH subscriptions/channels | 64 per process |
| Authentication challenges | At most 4096, two-minute TTL |
| WSS heartbeat | Every 15 s; write/heartbeat timeout 5 s |
| SSH handshake / maximum session | 10 s / 8 h |

These limits apply per process. Review shared enforcement and capacity before adding processes or exposing the service to the Internet.

## Consent and anonymity

A direct chat opens only after the recipient's reciprocal call. Requests expire after 7 days; a user may have 10 active outgoing requests and is limited to 1/minute, burst 3, per process. Recipients may reject or block. Blocking closes an existing chat and prevents new messages/requests; it does not erase previously delivered copies.

Recovery codes contain 192 random bits, are stored only as hashes and rotate when used. Recovery challenges expire after two minutes and allow one attempt. Recovery revokes prior SSH keys and sessions. Losing both key and code leaves no administrative email-recovery route. Pseudonyms, public keys, contact relationships and delivery metadata can identify a person; pseudonymity is not absolute anonymity.
