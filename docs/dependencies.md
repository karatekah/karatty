# Dependencies and sources

**English** · [Español](es/dependencies.md) · [Italiano](it/dependencies.md)

Versions were checked for the sprint on 2026-10-31. Exact module versions/hashes live in `go.mod`/`go.sum`; `go mod verify` checks the cache. Review upgrades with tests rather than using `@latest` in CI.

| Tool | Version used | Official source |
| --- | --- | --- |
| Portable Go for Windows | 1.27.4 | https://go.dev/dl/ |
| Bubble Tea | v2.1.1 | https://pkg.go.dev/charm.land/bubbletea/v2@v2.1.1 |
| Bubbles | v2.2.1 | https://pkg.go.dev/charm.land/bubbles/v2@v2.2.1 |
| Wish | v2.0.5 | https://pkg.go.dev/charm.land/wish/v2@v2.0.5 |
| pgx | v5.12.0 | https://github.com/jackc/pgx |
| WebSocket | v1.8.15 | https://pkg.go.dev/github.com/coder/websocket@v1.8.15 |
| PostgreSQL Windows binaries / Docker | 17.11 | https://www.enterprisedb.com/download-postgresql-binaries |
| Rust for the local crypto bridge | 1.99.1 | https://rust-lang.org/install/ |
| vodozemac (Olm 1:1) | 0.11.1 | https://github.com/matrix-org/vodozemac |

Charm v2 APIs use `charm.land`: View returns `tea.View`; keystrokes use `tea.KeyPressMsg`. Downloaded source, compilation and a real SSH PTY were used to check APIs. The prototype gateway uses the built-in `charm.land/ssh` handshake timeout.

`crypto/Cargo.lock` pins bridge dependencies. CI runs Rust 1.99.1 format/tests. `karatty-crypto` runs beside the local TUI and ships with it. `scripts/collect-licenses.ps1` preserves dependency licenses, standard-library notices and a package index. Pinned supplemental texts are documented in [third_party](../third_party/README.md). CI extracts packages, checks notices and executes the Go–Rust bridge on each platform.

The downloaded Windows Go archive matched the SHA-256 published by its official endpoint: `f26081f438319757f8a1b46d7d0b2c5ec7583d4038e485e0aa725cfde7885fff`.

PostgreSQL came from EDB's Windows 17.11 binary link. Locally calculated SHA-256: `80379b2c04d51c30225532e0ae04509899141e9957ed096fe749d7fd9df8f82f`. This fingerprints the received archive; it is not an independently verified EDB signature.

No wacon-mcp or WhatsApp code is reused. Original-context legal-compliance claims were not adopted without current verification. Naming conflicts remain research evidence, not a legal determination.
