# Local Olm bridge

**English** · [Español](../docs/es/crypto/README.md) · [Italiano](../docs/it/crypto/README.md)

The `karatty-crypto` executable is part of the local alpha client. It uses vodozemac 0.11.2 to create Olm identities, replenish one-time keys, and encrypt/decrypt messages. It accepts one JSON object on stdin and returns one on stdout. Rust tests cover bidirectional exchange, tampering, replay, out-of-order delivery and bounded forward secrecy. `pkg/cryptobridge` tests Go invocation and signatures for initial and replenished keys.

The `account` and `session` fields contain unencrypted private key material while crossing **local process pipes**. Never log them, persist them unencrypted or send them to the server. The bridge itself does not verify contact identity. `pkg/secureclient` verifies signed key bundles and fingerprints, stops on identity changes and atomically persists the advanced ratchet and pending envelope before transmission. `pkg/localstate` protects the archive with a local passphrase, Argon2id and XChaCha20-Poly1305.

Run `cargo test --locked` from `crypto/`. Build with `cargo build --locked --release` and keep the resulting executable beside `karatty`. See [ADR 0003](../docs/adr/0003-e2ee-1to1.md) and [verification](../docs/verification.md) for integration evidence and limits. Karatty has no independent audit.
