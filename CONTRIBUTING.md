# Contributing to Karatty

**English** · [Español](docs/es/project/CONTRIBUTING.md) · [Italiano](docs/it/project/CONTRIBUTING.md)

This guide explains how to reproduce the experimental alpha environment and propose changes. Report functional problems in [issues](https://github.com/karatekah/karatty/issues); report vulnerabilities privately under [SECURITY.md](SECURITY.md).

## Set up

Use Go 1.27.4, Rust 1.99.1 and PostgreSQL 17.11. The [local guide](docs/quickstart.md) covers PostgreSQL, development certificates and clients. Integration tests need `TERMLINK_TEST_DATABASE_URL` pointing to a **development database** with permission to create schemas. Set `KARATTY_CRYPTO_BIN` to the compiled bridge for E2EE tests. Never target production.

```sh
go mod verify
cargo test --locked --manifest-path crypto/Cargo.toml
cargo build --locked --release --manifest-path crypto/Cargo.toml
# Set KARATTY_CRYPTO_BIN to the absolute path of crypto/target/release/karatty-crypto.
go vet ./...
go test -count=1 ./...
go build ./cmd/...
```

PowerShell users can run `./scripts/check.ps1`. Check documentation with `./scripts/check-docs.ps1`. CI also runs Linux race detection, native Windows/Linux/macOS builds and the Docker startup test. Without the database variable, integration tests are skipped; that local result does not validate them.

## Propose changes

Open an issue before changing behavior or protocol. Keep each change focused: describe the problem, affected files, validation and remaining limits. Use conventional commits such as `feat:`, `fix:`, `docs:`, `test:` and `chore:`. Migrations must preserve existing data or clearly document unavoidable loss.

E2EE, identity, retention and deletion affect security and privacy. Explain threat models, protocol compatibility and failure tests when changing them. Do not add home-grown cryptographic primitives or claim guarantees unsupported by evidence.

English is the primary documentation language. Update the complete Spanish and Italian counterparts when changing a page, preserve command/API identifiers, and follow [the translation guide](docs/translations.md).

## Security reports

Do not publish exploitable vulnerabilities or sensitive data in issues. Use **Security → Report a vulnerability** on GitHub and follow [SECURITY.md](SECURITY.md). Public bug reports must use test data and omit keys, tokens and private conversations.
