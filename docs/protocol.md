# Protocol v1

**English** · [Español](es/protocol.md) · [Italiano](it/protocol.md)

JSON over HTTPS. All endpoints except authentication challenge/proof, recovery and health require `Authorization: Bearer SESSION_TOKEN`. Tokens are not accepted in URLs. Responses use `Cache-Control: no-store`. Errors: 400 invalid input, 401 authentication, 403 membership, 409 ID/content conflict, 429 rate limit. Internal errors do not expose queries or data.

Authentication returns `public_mode`. When true, the local client requires encrypted messaging; `--legacy` is restricted to the invitation prototype for local development.

| Operation | Route |
| --- | --- |
| Registration/login challenge | POST /v1/auth/challenge |
| Serialized SSH signature (JSON encodes []byte as base64) | POST /v1/auth/proof |
| Begin recovery with pseudonym and new public key | POST /v1/recovery/start |
| New-key signature and recovery code | POST /v1/recovery/complete |
| End/revoke session | DELETE /v1/session |
| Delete authenticated account | DELETE /v1/account |
| Accessible conversations, at most 100 | GET /v1/conversations |
| Request direct chat or accept incoming request | POST /v1/direct |
| Incoming direct requests | GET /v1/direct/requests |
| Reject incoming request | DELETE /v1/direct/requests/{username} |
| Block user and close direct chat | PUT /v1/direct/blocks/{username} |
| Unblock user | DELETE /v1/direct/blocks/{username} |
| Upload Olm identity and signed one-time public keys | POST /v1/e2ee/device |
| Read peer public identity after acceptance | GET /v1/e2ee/peer/{username} |
| Claim one peer one-time public key | POST /v1/e2ee/peer/{username}/claim |
| Ascending sequence history | GET /v1/conversations/{id}/messages?after=0&limit=100 |
| Send stable client_id and body | POST /v1/conversations/{id}/messages |
| Sequence events with Bearer header | WSS /v1/conversations/{id}/events?after=0 |

## Persistence and history

`protocol.Message` is also the persistence ACK: server ID, `client_id`, conversation, sender, username, sequence, body and creation time, returned after commit. It does not indicate reading or delivery to every session. Different senders may use the same client ID; one sender reuses exactly the same ID for the same message/conversation retry.

Deduplication is guaranteed for the message's 30-day retention. Purging deletes content and identifiers; a later send with that ID is new. Purging is automatic at startup/hourly with `--public-signup`, or manual/explicitly enabled with `server --retention-30d` in development.

`Page` contains `messages` and `next`. The legacy client keeps its last delivered sequence in memory and filters duplicate events. The secure client persists its ratchet, archive and server cursor; a reopened TUI rebuilds its view from the archive while server synchronization uses the durable cursor. Sequences are assigned under a conversation lock in the INSERT transaction. Purging/deletion may leave gaps; `next_seq` never moves backwards. History hides messages older than 720 hours before physical purging. Empty batches still require authorization. Each query returns at most 100 messages; slow consumers hit the write timeout and reconnect to the durable log.

`Conversation.closed` marks a direct chat archived after a participant is deleted. The remaining member can read authorized retained history but cannot send. Deletion removes that account's keys, sessions, memberships and authored messages, not other participants' messages or remote copies already delivered. Backup expiry is a separate operator policy.

## E2EE key distribution

The experimental `e2ee` endpoints follow [ADR 0003](adr/0003-e2ee-1to1.md). `POST /device` requires an SSH-key signature over `KARATTY-E2EE-IDENTITY-v1\nUSER_ID\nED25519\nCURVE25519`. Each one-time key requires an Olm Ed25519 signature over `KARATTY-OTK-v1\nCURVE25519\nONE_TIME_KEY`. Keys use unpadded standard base64.

Identities cannot be silently replaced; account recovery removes the old one. Only accepted direct-chat members can read/claim their peer's keys. Claimed keys cannot reactivate when uploaded again. The official client compares fingerprints through another channel, encrypts using Olm and stores ratchets in the encrypted local archive.

## Open registration and envelopes

`server --public-signup` disables invitation signup and registers pseudonyms with Ed25519 public keys, without email. It lists only direct chats and rejects channel history, events and sends even for older channels. Direct `body` values must be JSON envelopes with `version:1`, `session_id`, `sender_identity` and `message:{type,body}`; Olm key/body encoding is unpadded standard base64. The API validates **structure**, not whether a modified sender encrypted correctly. Official clients produce and consume envelopes locally.

`auth/challenge` uses signing domain `KARATTY-PUBLIC-REGISTER-v1`; `auth/proof` consumes it in one attempt and requires the new key's signature. The response contains a temporary session and a 192-bit recovery code shown once. Invitation registration also returns that code.

`recovery/start` takes a pseudonym and new public key; its response does not reveal whether the pseudonym exists. `recovery/complete` requires that key's signature and the saved code, revokes earlier keys/sessions and rotates the code. Challenges expire in two minutes.

## Prototype SSH and listeners

Prototype SSH authenticates public-key possession and creates a domain session per SSH channel. A PTY shell opens a server-side TUI. Allowed commands: `list`, `requests`, `direct USERNAME`, `decline USERNAME`, `block USERNAME`, `unblock USERNAME`, `history CONVERSATION AFTER`, `send CONVERSATION CLIENT_ID` with body on stdin. It does not execute system commands. **This access is not E2EE and is outside the alpha.** Combining `--public-signup` and `--ssh-host-key` is rejected.

Open registration defaults to loopback. Explicit `--public-network` requires TLS and rejects `--dev` and a remote SSH TUI. See [local/LAN setup](quickstart.md) and the [future deployment plan](deploy.md). Hosted infrastructure/operations remain future work.
