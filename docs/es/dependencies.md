# Dependencias y fuentes

[English](../dependencies.md) · **Español** · [Italiano](../it/dependencies.md)

Comprobadas para este sprint el 2026-10-31. Se fijan versiones exactas y hashes de módulos en go.mod/go.sum; `go mod verify` comprueba la caché. Se debe revisar cualquier actualización con pruebas, no usar `@latest` en CI.

| Herramienta | Versión usada | Fuente oficial |
| --- | --- | --- |
| Go portable Windows | 1.27.3 | https://go.dev/dl/ |
| Bubble Tea | v2.1.1 | https://pkg.go.dev/charm.land/bubbletea/v2@v2.1.1 |
| Bubbles | v2.2.1 | https://pkg.go.dev/charm.land/bubbles/v2@v2.2.1 |
| Wish | v2.0.5 | https://pkg.go.dev/charm.land/wish/v2@v2.0.5 |
| pgx | v5.12.0 | https://github.com/jackc/pgx |
| WebSocket | v1.8.15 | https://pkg.go.dev/github.com/coder/websocket@v1.8.15 |
| PostgreSQL portable Windows y Docker | 17.11 | https://www.enterprisedb.com/download-postgresql-binaries |
| Rust del prototipo criptográfico | 1.99.1 | https://rust-lang.org/install/ |
| vodozemac (Olm 1:1) | 0.11.1 | https://github.com/matrix-org/vodozemac |

Las API v2 de Charm usan `charm.land`; View devuelve tea.View y las pulsaciones son tea.KeyPressMsg. Las API se contrastaron con el código de las versiones descargadas y se verificaron mediante compilación y un PTY real sobre SSH. El gateway usa el timeout de handshake incorporado en charm.land/ssh.

`crypto/Cargo.lock` fija las dependencias del puente criptográfico. CI comprueba formato y pruebas con Rust 1.99.1. `karatty-crypto` se ejecuta localmente junto con la TUI; el empaquetado nativo incluye ambos ejecutables. `scripts/collect-licenses.ps1` conserva los textos suministrados por las dependencias, los avisos de las bibliotecas estándar y un índice dentro de cada paquete. Los textos complementarios fijados están documentados en [third_party](third_party/README.md). CI extrae la descarga, comprueba sus avisos y ejecuta el puente Go–Rust en cada plataforma.

El archivo Go de Windows descargado del endpoint oficial fue validado contra el SHA-256 anunciado allí: `f26081f438319757f8a1b46d7d0b2c5ec7583d4038e485e0aa725cfde7885fff`.

PostgreSQL procede del enlace de binarios de Windows de EDB para 17.11. SHA-256 calculado localmente: `80379b2c04d51c30225532e0ae04509899141e9957ed096fe749d7fd9df8f82f`. Es una huella del archivo recibido, no una firma publicada por EDB que se haya verificado independientemente.

No se reutiliza código de wacon-mcp ni de WhatsApp. No se incorporaron afirmaciones de cumplimiento legal del contexto original sin verificación actual. Los conflictos de nombre se registran como evidencia de investigación pendiente, no como un dictamen jurídico.
