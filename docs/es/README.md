<div align="center">

![Marca de terminal Karatty](../assets/karatty-mark.svg)

# Karatty

[English](../../README.md) · **Español** · [Italiano](../it/README.md)

**Conversaciones cifradas. Desde tu terminal.**

Una TUI local con directos 1:1 por aceptación, identidad SSH y cifrado de extremo a extremo.

[![v0.1.0-alpha.1](https://img.shields.io/badge/release-v0.1.0--alpha.1-4f46e5?style=flat-square)](https://github.com/karatekah/karatty/releases/tag/v0.1.0-alpha.1) [![CI](https://github.com/karatekah/karatty/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/karatekah/karatty/actions/workflows/ci.yml) [![Licencia Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-7ae2b5?style=flat-square)](../../LICENSE) [![Paquetes del cliente: Windows, Linux y macOS](https://img.shields.io/badge/platforms-Windows%20%7C%20Linux%20%7C%20macOS-64748b?style=flat-square)](install.md)

[Inicio rápido](quickstart.md) · [Descargar alpha](https://github.com/karatekah/karatty/releases/tag/v0.1.0-alpha.1) · [Documentación](#documentación) · [Roadmap](roadmap.md) · [Reportar un bug](https://github.com/karatekah/karatty/issues)

</div>

---

Ejecuta la **alpha experimental** en tu equipo: un servidor con PostgreSQL y un cliente local por participante. Sin dominio, VPS ni correo.

## Qué incluye la alpha

- **Identidad SSH.** Registro sin correo ni nombre real, con prueba de posesión de una clave Ed25519. Recuperación con un código guardado fuera del equipo.
- **Consentimiento primero.** Chats directos que requieren aceptación; solicitudes, rechazo, bloqueo y conversaciones cerradas de solo lectura.
- **E2EE local.** Olm 1:1 con Double Ratchet mediante vodozemac 0.11.2. La TUI exige comparar la huella del contacto por otro medio y verificarla de nuevo si cambia.
- **Archivo cifrado.** Archivo local cifrado con frase secreta, Argon2id y XChaCha20-Poly1305. Los envíos pendientes se guardan antes de transmitirlos y pueden recuperarse tras reiniciar.
- **Entrega persistente.** Sobres opacos en PostgreSQL, historial por secuencia, reconexión y deduplicación durante los 30 días de retención.
- **Ciclo de vida de datos.** Eliminación de cuenta y archivos locales, purga de mensajes vencidos y herramientas de respaldo cifrado y restauración.

## Probar y usar

Para ensayar ahora, sigue la [guía local y LAN](quickstart.md): compilar, iniciar PostgreSQL y el servidor con TLS, y conversar con dos cuentas. La [guía de instalación y primer uso](install.md) explica los paquetes, claves, huellas, reintentos y recuperación. La [matriz de verificación](verification.md) distingue las pruebas automatizadas de las realizadas en equipos físicos.

El flujo dentro de la TUI es:

```text
Alice: /direct bob
Bob:   /requests
Bob:   /direct alice
Ambos: /myfingerprint
       Comparan sus huellas por otro medio.
Alice: /verify bob HUELLA_DE_BOB
Bob:   /verify alice HUELLA_DE_ALICE
```

Después pueden escribir mensajes. Tab cambia de chat, Ctrl+L recarga y Ctrl+C sale. `/pending` muestra envíos guardados sin confirmar; `/retry ID` conserva el mismo sobre al reintentarlos. El ACK confirma persistencia en el servidor, no lectura por el destinatario.

## Arquitectura y tecnologías

| Parte | Tecnología |
| --- | --- |
| Cliente y API | Go 1.27.4, monolito modular y contratos JSON compartidos |
| Interfaz | Bubble Tea v2 y Bubbles v2 |
| E2EE local | Rust 1.99.1, vodozemac 0.11.2 y Olm v1 |
| Archivo local | Argon2id y XChaCha20-Poly1305, bloqueo y reemplazo de archivo |
| Transporte | HTTPS y WSS; autorización por operación |
| Persistencia | PostgreSQL 17.11 y pgx, secuencias y ACK transaccionales |
| Operación prevista | Linux/systemd, TLS y copias `pg_dump` cifradas con `age` |

La [arquitectura](architecture.md) describe paquetes y flujos. El [ADR de E2EE](adr/0003-e2ee-1to1.md) documenta identidades, ratchet, archivo y modelo de amenazas. El módulo Go actual es `github.com/karatekah/karatty`. La versión histórica v0.1.0-alpha.1 usaba el nombre anterior del repositorio y del módulo. Por compatibilidad se conservan rutas como `cmd/termlink` y variables `TERMLINK_*`.

## Documentación

| Necesidad | Guía |
| --- | --- |
| Instalar, conversar, actualizar y borrar | [Primer uso](install.md) |
| Ejecutar y desarrollar localmente | [Entorno local](quickstart.md) · [Contribuir](project/CONTRIBUTING.md) |
| Entender el sistema | [Arquitectura](architecture.md) · [Protocolo](protocol.md) · [Dependencias](dependencies.md) |
| Revisar seguridad | [Garantías y límites](security.md) · [Pruebas](verification.md) · [Reportes privados](project/SECURITY.md) |
| Alojar en Internet en el futuro | [Despliegue previsto](deploy.md) · [Operación y respaldos](operations.md) |
| Revisar la publicación | [Release](release.md) · [Notas de la versión](releases/v0.1.0-alpha.1.md) · [Cambios](project/CHANGELOG.md) · [Roadmap](roadmap.md) |
| Borradores para un futuro servicio alojado | [Privacidad](privacy-draft.md) · [Condiciones](terms-draft.md), borradores no vigentes |
| Mantener las traducciones | [Idiomas de la documentación](translations.md) |

## Desarrollo

<details>
<summary>Compilar y probar</summary>

```sh
go mod verify
cargo test --locked --manifest-path crypto/Cargo.toml
cargo build --locked --release --manifest-path crypto/Cargo.toml
go vet ./...
go test -count=1 ./...
go build ./cmd/...
```

Las pruebas de integración necesitan `TERMLINK_TEST_DATABASE_URL` hacia una base de desarrollo y `KARATTY_CRYPTO_BIN` con la ruta absoluta al puente compilado. Sin esas variables se omiten las pruebas que las requieren; una ejecución local verde no sustituye CI. Actions prueba PostgreSQL real, detector de carreras, Docker, Go y Rust, paquetes extraídos en tres sistemas y avisos de vulnerabilidades de las dependencias.

</details>

---

El código propio usa [Apache-2.0](../../LICENSE). Los paquetes conservan los textos y atribuciones de terceros junto con sus versiones fijadas.
