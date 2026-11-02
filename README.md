# Karatty

Mensajería desde tu terminal: directos 1:1 con aceptación, identidad seudónima mediante clave SSH y cifrado de extremo a extremo en una TUI local.

**[v0.1.0-alpha.1](https://github.com/karatekah/termlink/releases/tag/v0.1.0-alpha.1) es una alpha experimental para uso local.** El código y los paquetes del cliente se distribuyen bajo Apache-2.0. No necesitas dominio, VPS ni correo: una máquina ejecuta el servidor y PostgreSQL, y cada participante instala el cliente. La configuración de LAN está documentada, pero la conversación entre dos equipos físicos aún no se ha comprobado. El alojamiento en Internet con `karatty.dev` queda para el futuro.

## Qué incluye la alpha

- Registro sin correo ni nombre real, con prueba de posesión de una clave Ed25519. Recuperación con un código guardado fuera del equipo.
- Chats directos que requieren aceptación; solicitudes, rechazo, bloqueo y conversaciones cerradas de solo lectura.
- Olm 1:1 con Double Ratchet mediante vodozemac 0.11.1. La TUI exige comparar la huella del contacto por otro medio y verificarla de nuevo si cambia.
- Archivo local cifrado con frase secreta, Argon2id y XChaCha20-Poly1305. Los envíos pendientes se guardan antes de transmitirlos y pueden recuperarse tras reiniciar.
- Sobres opacos en PostgreSQL, historial por secuencia, reconexión y deduplicación durante los 30 días de retención.
- Eliminación de cuenta y archivos locales, purga de mensajes vencidos y herramientas de respaldo cifrado y restauración.

## Probar y usar

Para ensayar ahora, sigue la [guía local y LAN](docs/quickstart.md): compilar, iniciar PostgreSQL y el servidor con TLS, y conversar con dos cuentas. La [guía de instalación y primer uso](docs/install.md) explica los paquetes, claves, huellas, reintentos y recuperación. La [matriz de verificación](docs/verification.md) distingue las pruebas automatizadas de las realizadas en equipos físicos.

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

La alpha usa solo la TUI instalada localmente y directos 1:1. El prototipo anterior por invitaciones y la TUI remota por SSH siguen en el código para desarrollo, envían texto legible al servidor y se excluyen de la alpha.

**Windows Defender:** clasificó una compilación del servidor como `Behavior:Win32/DefenseEvasion.A!ml`. En el equipo de prueba funcionó después de que su propietario restaurara el archivo de cuarentena, con Defender activo. No se ha confirmado un falso positivo ni validado esa detección en otros equipos. La descarga del cliente no incluye el servidor. Consulta la [evidencia y sus límites](docs/verification.md) antes de usar la alpha; no desactives la protección para instalarla.

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
| Alojar en Internet en el futuro | [Despliegue previsto](docs/deploy.md) · [Operación y respaldos](docs/operations.md) |
| Revisar la publicación | [Release](docs/release.md) · [Cambios](CHANGELOG.md) · [Roadmap](docs/roadmap.md) |
| Borradores para un futuro servicio alojado | [Privacidad](docs/privacy-draft.md) · [Condiciones](docs/terms-draft.md), pendientes de datos del operador |

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
