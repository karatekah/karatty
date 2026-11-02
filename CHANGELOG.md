# Changelog

**English** · [Español](docs/es/project/CHANGELOG.md) · [Italiano](docs/it/project/CHANGELOG.md)

## v0.1.0-alpha.1 — experimental alpha

- Local TUI for 1:1 direct chats with acceptance, rejection and blocking.
- Pseudonymous registration using an Ed25519 SSH key, without email; recovery through an offline code.
- Olm encryption with vodozemac 0.11.1, fingerprints compared through another channel and re-verification after an identity change.
- Argon2id/XChaCha20-Poly1305 local archive, outbox persisted before sending and archive migration after account recovery.
- Opaque PostgreSQL envelopes, sequence-based history and deduplication during 30-day retention.
- Account/local-state deletion, expired-message purging and encrypted backup/restore tools.
- Native client/crypto-bridge packages, hashes, source commit and third-party license texts.
- Local use with a self-hosted TLS server; certificates with IP SANs and a two-account guide. Domain, VPS, mailboxes and hosted service remain future work.
- Native Windows server/two-TUI trial passed registration, acceptance, fingerprints, E2EE messages and history reopening after the server was restored from Defender quarantine, with protection active. Two-machine LAN and the detection on other installations remain unverified.

## Documentation and repository migration

The repository and current Go module are now `karatekah/karatty`. English is primary, with complete Spanish/Italian pages and per-page language links. Published v0.1.0-alpha.1 tags, binaries and checksums are unchanged; its source snapshot retains the previous internal module identifier. Compatibility command paths and environment-variable names remain as documented.
