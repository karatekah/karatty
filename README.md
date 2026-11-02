# Karatty

**English** · [Español](docs/es/README.md) · [Italiano](docs/it/README.md)

Terminal messaging: accepted 1:1 direct chats, pseudonymous SSH-key identities and end-to-end encryption in a local TUI.

**[v0.1.0-alpha.1](https://github.com/karatekah/karatty/releases/tag/v0.1.0-alpha.1) is an experimental alpha for local use.** The project's code and client packages are distributed under Apache-2.0. No domain, VPS or email is required: one machine runs the server and PostgreSQL, and each participant installs the client. LAN configuration is documented, but a conversation between two physical machines has not been verified. Internet hosting with `karatty.dev` is future work.

## What the alpha includes

- Registration without email or a real name, using proof of possession of an Ed25519 SSH key. Recovery uses a code stored outside the device.
- Direct chats requiring acceptance, incoming requests, rejection, blocking and closed conversations with read-only history.
- Olm 1:1 with Double Ratchet through vodozemac 0.11.1. Contacts must compare fingerprints through another trusted channel and verify again after an identity change.
- A passphrase-encrypted local archive using Argon2id and XChaCha20-Poly1305. Pending sends are saved before transmission and survive restarts.
- Opaque envelopes in PostgreSQL, sequence-based history, reconnection and deduplication during the 30-day retention window.
- Account and local-file deletion, expired-message purging, and tools for encrypted backups and restoration.

## Try it

Follow the [local and LAN guide](docs/quickstart.md) to build, start PostgreSQL and the TLS server, and try two accounts. [Installation and first use](docs/install.md) covers packages, keys, fingerprints, retries and recovery. The [verification matrix](docs/verification.md) separates automated evidence from physical-machine tests.

Inside the TUI:

```text
Alice: /direct bob
Bob:   /requests
Bob:   /direct alice
Both: /myfingerprint
       Compare fingerprints through another channel.
Alice: /verify bob BOB_FINGERPRINT
Bob:   /verify alice ALICE_FINGERPRINT
```

After verification, type messages normally. Tab switches chats, Ctrl+L reloads and Ctrl+C exits. `/pending` lists saved, unconfirmed sends; `/retry ID` reuses the same envelope. An ACK confirms server persistence, not recipient reading.

## Security and limits

The alpha uses an installed local TUI and 1:1 direct chats. The older invitation prototype and remote SSH TUI remain in the source for development, expose readable content to the server and are outside this alpha.

**Windows Defender:** it classified one server build as `Behavior:Win32/DefenseEvasion.A!ml`. That build worked on the test machine after its owner restored it from quarantine, with Defender active. A false positive has not been confirmed, and other installations have not been validated. Client downloads do not include the server. Read the [evidence and limits](docs/verification.md); do not disable protection to install.

Forward secrecy covers earlier messages whose message keys have been consumed and deleted from the ratchet. It does not protect an unlocked local archive, an old state copy or a compromised device. Fingerprint comparison through another trusted channel is required against first-contact identity substitution. The server sees delivery metadata; a pseudonym does not guarantee absolute anonymity.

Karatty has no independent audit. Retention is 720 hours; local cleanup can only run while the client is used. The local archive is limited to 16 MiB, the screen to 500 messages and the conversation list to 100 entries. One device per account is supported. Groups, remote SSH, attachments, search and notifications are future work.

## Architecture and stack

| Component | Technology |
| --- | --- |
| Client and API | Go 1.27.3, modular monolith and shared JSON contracts |
| Terminal UI | Bubble Tea v2 and Bubbles v2 |
| Local E2EE | Rust 1.99.1, vodozemac 0.11.1 and Olm v1 |
| Local archive | Argon2id and XChaCha20-Poly1305, file locking and replacement |
| Transport | HTTPS and WSS; authorization on every operation |
| Storage | PostgreSQL 17.11 and pgx, transactional sequences and ACKs |
| Future hosting | Linux/systemd, TLS and `pg_dump` backups encrypted with `age` |

[Architecture](docs/architecture.md) explains packages and flows. [ADR 0003](docs/adr/0003-e2ee-1to1.md) describes identities, ratchets, storage and threats. The current Go module is `github.com/karatekah/karatty`. Historical release v0.1.0-alpha.1 used the previous repository/module name. Compatibility identifiers such as `cmd/termlink` and `TERMLINK_*` environment variables retain their existing spelling.

## Documentation

| Need | Guide |
| --- | --- |
| Install, chat, update and delete | [First use](docs/install.md) |
| Run and develop locally | [Local setup](docs/quickstart.md) · [Contributing](CONTRIBUTING.md) |
| Understand the system | [Architecture](docs/architecture.md) · [Protocol](docs/protocol.md) · [Dependencies](docs/dependencies.md) |
| Review security | [Security and privacy](docs/security.md) · [Verification](docs/verification.md) · [Private reports](SECURITY.md) |
| Future Internet hosting | [Deployment plan](docs/deploy.md) · [Operations and backups](docs/operations.md) |
| Release and future work | [Release process](docs/release.md) · [Release notes](docs/releases/v0.1.0-alpha.1.md) · [Changelog](CHANGELOG.md) · [Roadmap](docs/roadmap.md) |
| Future hosted-service drafts | [Privacy](docs/privacy-draft.md) · [Terms](docs/terms-draft.md); not active service policies |
| Maintain translations | [Documentation languages](docs/translations.md) |

## Development

```sh
go mod verify
cargo test --locked --manifest-path crypto/Cargo.toml
cargo build --locked --release --manifest-path crypto/Cargo.toml
go vet ./...
go test -count=1 ./...
go build ./cmd/...
```

Integration tests require `TERMLINK_TEST_DATABASE_URL` pointing to a development database and `KARATTY_CRYPTO_BIN` pointing to the compiled bridge. Tests requiring missing variables are skipped; a green local run does not replace CI. Actions uses real PostgreSQL, race detection, Docker, Go/Rust, packages extracted on three operating systems and dependency vulnerability checks.

The project's code uses [Apache-2.0](LICENSE). Packages preserve third-party license texts and pinned dependency versions.
