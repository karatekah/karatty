# Karatty (alpha en preparación)

Mensajería desde una TUI instalable. La primera alpha se limitará a chats directos 1:1 que requieren aceptación, registro seudónimo sin correo y cifrado de extremo a extremo. Este repositorio sigue privado mientras se completa esa versión; el código, la documentación y los ejecutables se publicarán juntos cuando pase las comprobaciones de lanzamiento.

## Sprint 1

El primer sprint demostró que dos usuarios invitados podían conversar entre SSH y el cliente instalable, salir y recuperar el historial desde PostgreSQL; un tercero sin membresía no podía leer, enviar ni suscribirse. Ahora se prepara la alpha pública con un alcance más pequeño y E2EE obligatorio.

Go, monolito modular, PostgreSQL y Bubble Tea v2/Bubbles v2/Wish v2. El prototipo actual ofrece TUI local y por SSH, con HTTPS/WSS y SSH para el transporte. **El servidor todavía puede leer los mensajes; no hay E2EE y esta versión no debe abrirse al público.** La alpha usará solo la TUI local para chats directos 1:1; SSH remoto y grupos quedan para después.

## Ejecutar

Sigue la [guía local para Windows, Linux y macOS](docs/quickstart.md). Incluye PostgreSQL con Docker, certificados TLS locales, creación de invitaciones, acceso nativo/SSH y prueba de un tercero sin membresía. CI comprueba el arranque en Docker y el flujo HTTPS con el cliente nativo; las pruebas de integración cubren SSH por separado. No hay servidor público ni servicios contratados.

```sh
go build ./cmd/...
go vet ./...
# Con una base PostgreSQL local y TERMLINK_TEST_DATABASE_URL configurado:
go test -count=1 ./...
```

## Implementado

- Invitaciones con caducidad/usos y registro con prueba de posesión Ed25519; desafío de un solo uso, sesión temporal y revocación administrativa.
- Registro público opcional con seudónimo y clave SSH, sin correo ni nombre real; recuperación de clave con un código aleatorio mostrado una sola vez. Perder la clave y el código impide recuperar la cuenta.
- Eliminación autenticada de cuenta, claves, sesiones, membresías y mensajes propios; los mensajes ajenos permanecen en un chat directo archivado y de solo lectura.
- Canales asignados por invitación y chats directos que requieren aceptación del destinatario; solicitudes de 7 días, rechazo y bloqueo. Autorización por sesión/clave/usuario/membresía en cada operación.
- Mensajes persistentes, secuencia transaccional por conversación, ACK de commit y deduplicación durante 30 días, con conflicto al cambiar el contenido de un reintento dentro de ese plazo.
- Historial paginado y eventos WSS con reconexión por secuencia; gateway Wish v2 autenticado al mismo núcleo y TUI Bubble Tea v2/Bubbles v2 compartida.
- Límites de tamaño, tasas, conexiones, páginas y escritura; controles de terminal rechazados en entrada y filtrados en pantalla. Logs operativos sin cuerpos/credenciales.
- Pruebas de integración con PostgreSQL real, TLS, WebSocket, cliente ejecutable y SSH con PTY; CI con detector de carreras en Linux, compilación Windows/Linux/macOS y prueba Docker.

El cliente conserva tokens solo en memoria; solicita la clave en cada arranque. La TUI retiene 500 mensajes en pantalla y muestra las primeras 100 conversaciones. El registro inicial se realiza con el CLI. En el prototipo, SSH permite entrar sin instalar el cliente; esa vía no formará parte de la alpha. El tiempo real usa sondeos de 250 ms del registro PostgreSQL; no necesita Redis.

## Pendiente antes de abrir la alpha

Faltan E2EE verificable en chats directos 1:1, eliminar el acceso a mensajes en claro del servidor y retirar SSH remoto y canales grupales de la alpha pública. También faltan el dominio y VPS, la prueba operativa de la purga de **30 días**, la retención/restauración de respaldos remotos y la revisión de seguridad/privacidad. La eliminación de cuentas ya funciona en la base activa, pero falta verificar su vencimiento en respaldos remotos. El historial ya oculta mensajes vencidos; la purga automática se activa con `--retention-30d` o automáticamente con `--public-signup`. No hay archivos, búsqueda, notificaciones, organizaciones ni facturación. El [roadmap](docs/roadmap.md) distingue estas tareas del sprint. Karatty es el nombre elegido para la alpha; la marca y el dominio aún requieren comprobación definitiva antes de publicar.

## Estructura y documentación

`cmd/server`, `cmd/ssh-gateway`, `cmd/termlink`; `internal/auth`, `chat`, `storage`, `security`, `transport`; `pkg/tui`, `client`, `protocol`; `migrations`, `tests`, `deployments`, `docs`. Los límites reales de paquetes se explican en el [ADR 0001](docs/adr/0001-foundation.md), y la autenticación en el [ADR 0002](docs/adr/0002-authentication.md).

[Arquitectura y tecnologías](docs/architecture.md) · [Protocolo](docs/protocol.md) · [Seguridad](docs/security.md) · [Roadmap](docs/roadmap.md) · [Fuentes y versiones](docs/dependencies.md)
[Operación y respaldos para la alpha](docs/operations.md)
[Borrador de privacidad (no publicable aún)](docs/privacy-draft.md)
[Preparación de la publicación](docs/release.md) · [Contribuir](CONTRIBUTING.md)

La [guía de publicación](docs/release.md) reúne las comprobaciones para abrir este repositorio y distribuir una versión verificable. Aún no se ha elegido una licencia pública; hasta entonces, la visibilidad futura del código no concede permisos de reutilización. Las dependencias conservan sus licencias propias.
