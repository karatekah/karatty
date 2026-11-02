# Karatty

Mensajería desde tu terminal: directos 1:1 con aceptación, identidad seudónima mediante clave SSH y cifrado de extremo a extremo en una TUI local.

**v0.1.0-alpha.1 está en preparación.** El repositorio sigue privado, no hay servicio público ni descarga publicada. El dominio elegido es `karatty.dev`; faltan su compra, el VPS y el contacto público del operador. El código, los paquetes y la documentación se publicarán juntos cuando se compruebe el despliegue.

## Qué incluye la candidata

- Registro sin correo ni nombre real, con prueba de posesión de una clave Ed25519. Recuperación con un código guardado fuera del equipo.
- Chats directos que requieren aceptación; solicitudes, rechazo, bloqueo y conversaciones cerradas de solo lectura.
- Olm 1:1 con Double Ratchet mediante vodozemac 0.11.1. La TUI exige comparar la huella del contacto por otro medio y verificarla de nuevo si cambia.
- Archivo local cifrado con frase secreta, Argon2id y XChaCha20-Poly1305. Los envíos pendientes se guardan antes de transmitirlos y pueden recuperarse tras reiniciar.
- Sobres opacos en PostgreSQL, historial por secuencia, reconexión y deduplicación durante los 30 días de retención.
- Eliminación de cuenta y archivos locales, purga de mensajes vencidos y herramientas de respaldo cifrado y restauración.

## Probar y usar

Para ensayar ahora, sigue el [entorno local](docs/quickstart.md), que incluye PostgreSQL, certificados de desarrollo y dos cuentas de prueba. La [guía de instalación y primer uso](docs/install.md) explica la descarga prevista, las claves, las huellas, los reintentos y la recuperación.

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

## Seguridad y límites

La alpha usa solo la TUI instalada localmente y directos 1:1. El prototipo anterior por invitaciones y la TUI remota por SSH siguen en el código para desarrollo, envían texto legible al servidor y se excluyen del despliegue público.

El secreto futuro protege mensajes cuyas claves ya se consumieron y borraron del ratchet. No protege un archivo local desbloqueado, una copia antigua del estado o un equipo comprometido. La verificación de huellas por otro medio es necesaria frente a sustituciones de identidad en el primer contacto. El servidor conoce los metadatos de entrega; un seudónimo no garantiza anonimato absoluto.

No hay auditoría independiente de Karatty. La [matriz de verificación](docs/verification.md) explica lo probado y sus límites. La retención es de 720 horas; el cliente solo puede limpiar un equipo cuando se usa. El archivo local está acotado a 16 MiB, la pantalla a 500 mensajes y el listado a 100 conversaciones. La alpha admite un dispositivo por cuenta. Grupos, SSH remoto, adjuntos, búsqueda y notificaciones quedan para después.

## Arquitectura y tecnologías

| Parte | Tecnología |
| --- | --- |
| Cliente y API | Go 1.27.3, monolito modular y contratos JSON compartidos |
| Interfaz | Bubble Tea v2 y Bubbles v2 |
| E2EE local | Rust 1.99.1, vodozemac 0.11.1 y Olm v1 |
| Archivo local | Argon2id y XChaCha20-Poly1305, bloqueo y reemplazo de archivo |
| Transporte | HTTPS y WSS; autorización por operación |
| Persistencia | PostgreSQL 17.11 y pgx, secuencias y ACK transaccionales |
| Operación prevista | Linux/systemd, TLS y copias `pg_dump` cifradas con `age` |

La [arquitectura](docs/architecture.md) describe paquetes y flujos. El [ADR de E2EE](docs/adr/0003-e2ee-1to1.md) documenta identidades, ratchet, archivo y modelo de amenazas. El módulo y repositorio conservan `termlink` como identificador interno.

## Documentación

| Necesidad | Guía |
| --- | --- |
| Instalar, conversar, actualizar y borrar | [Primer uso](docs/install.md) |
| Ejecutar y desarrollar localmente | [Entorno local](docs/quickstart.md) · [Contribuir](CONTRIBUTING.md) |
| Entender el sistema | [Arquitectura](docs/architecture.md) · [Protocolo](docs/protocol.md) · [Dependencias](docs/dependencies.md) |
| Revisar seguridad | [Garantías y límites](docs/security.md) · [Pruebas](docs/verification.md) · [Reportes privados](SECURITY.md) |
| Operar y respaldar | [Despliegue](docs/deploy.md) · [Operación](docs/operations.md) |
| Revisar la publicación | [Release](docs/release.md) · [Cambios](CHANGELOG.md) · [Roadmap](docs/roadmap.md) |
| Datos y condiciones del servicio | [Privacidad](docs/privacy-draft.md) · [Condiciones](docs/terms-draft.md), borradores pendientes de datos del operador |

## Desarrollo

```sh
go mod verify
cargo test --locked --manifest-path crypto/Cargo.toml
cargo build --locked --release --manifest-path crypto/Cargo.toml
go vet ./...
go test -count=1 ./...
go build ./cmd/...
```

Las pruebas de integración necesitan `TERMLINK_TEST_DATABASE_URL` hacia una base de desarrollo y `KARATTY_CRYPTO_BIN` con la ruta absoluta al puente compilado. Sin esas variables se omiten las pruebas que las requieren; una ejecución local verde no sustituye CI. Actions prueba PostgreSQL real, detector de carreras, Docker, Go y Rust, paquetes extraídos en tres sistemas y avisos de vulnerabilidades de las dependencias.

El código propio usa [Apache-2.0](LICENSE). Los paquetes conservan los textos y atribuciones de terceros junto con sus versiones fijadas.
