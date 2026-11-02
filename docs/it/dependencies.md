# Dipendenze e fonti

[English](../dependencies.md) · [Español](../es/dependencies.md) · **Italiano**

Versioni controllate per lo sprint il 2026-10-31. `go.mod`/`go.sum` fissano versioni/hash; `go mod verify` controlla la cache. Valutare aggiornamenti con test, senza `@latest` in CI.

| Strumento | Versione usata | Fonte ufficiale |
| --- | --- | --- |
| Go Windows portabile | 1.27.3 | https://go.dev/dl/ |
| Bubble Tea | v2.1.1 | https://pkg.go.dev/charm.land/bubbletea/v2@v2.1.1 |
| Bubbles | v2.2.1 | https://pkg.go.dev/charm.land/bubbles/v2@v2.2.1 |
| Wish | v2.0.5 | https://pkg.go.dev/charm.land/wish/v2@v2.0.5 |
| pgx | v5.12.0 | https://github.com/jackc/pgx |
| WebSocket | v1.8.15 | https://pkg.go.dev/github.com/coder/websocket@v1.8.15 |
| PostgreSQL Windows/Docker | 17.11 | https://www.enterprisedb.com/download-postgresql-binaries |
| Rust del bridge locale | 1.99.1 | https://rust-lang.org/install/ |
| vodozemac (Olm 1:1) | 0.11.1 | https://github.com/matrix-org/vodozemac |

Le API Charm v2 usano `charm.land`: View restituisce `tea.View` e i tasti `tea.KeyPressMsg`. Sorgenti scaricati, compilazione e PTY SSH reale hanno verificato le API; il gateway usa il timeout handshake di `charm.land/ssh`.

`crypto/Cargo.lock` fissa le dipendenze. CI verifica formato/test con Rust 1.99.1. Il bridge locale è distribuito accanto alla TUI. `scripts/collect-licenses.ps1` conserva licenze, avvisi delle librerie standard e indice. I testi supplementari sono descritti in [third_party](third_party/README.md). CI estrae pacchetti, controlla avvisi ed esegue il bridge sui tre sistemi.

L'archivio Go Windows coincide con SHA-256 ufficiale: `f26081f438319757f8a1b46d7d0b2c5ec7583d4038e485e0aa725cfde7885fff`.

PostgreSQL proviene da EDB Windows 17.11. SHA-256 locale: `80379b2c04d51c30225532e0ae04509899141e9957ed096fe749d7fd9df8f82f`. Identifica l'archivio ricevuto; non è una firma EDB verificata indipendentemente.

Non si riusa codice wacon-mcp o WhatsApp. Non sono state adottate affermazioni di conformità legale del contesto iniziale senza verifica attuale. I conflitti di nome restano ricerca, non parere giuridico.
