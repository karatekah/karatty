# TERMLINK (nombre provisional)

Mensajería propia desde la terminal, con un cliente nativo y acceso SSH. Repositorio privado de trabajo de karatekah. Sin integración ni código de WhatsApp o wacon-mcp.

## Sprint 1

Objetivo: dos usuarios invitados conversan entre SSH y el cliente instalable, salen y recuperan el historial desde PostgreSQL; un tercero sin membresía no puede leer, enviar ni suscribirse.

Go, monolito modular, PostgreSQL y Bubble Tea v2/Bubbles v2/Wish v2. Ambos accesos ejecutan las mismas operaciones de dominio y comparten TUI. HTTPS/WSS y SSH protegen el transporte; **el servidor puede leer los mensajes; no hay E2EE**.

## Ejecutar

Sigue la [guía local para Windows, Linux y macOS](docs/quickstart.md). Incluye PostgreSQL con Docker, certificados TLS locales, creación de invitaciones, acceso nativo/SSH y prueba de un tercero sin membresía. El entorno completo en Docker se comprueba en CI. No hay servidor público ni servicios contratados.

```sh
go build ./cmd/...
go vet ./...
# Con una base PostgreSQL local y TERMLINK_TEST_DATABASE_URL configurado:
go test -count=1 ./...
```

## Implementado

- Invitaciones con caducidad/usos y registro con prueba de posesión Ed25519; desafío de un solo uso, sesión temporal y revocación administrativa.
- Canales asignados por invitación y chats directos; autorización por sesión/clave/usuario/membresía en cada operación.
- Mensajes persistentes, secuencia transaccional por conversación, ACK de commit, deduplicación y conflicto al cambiar el contenido de un reintento.
- Historial paginado y eventos WSS con reconexión por secuencia; gateway Wish v2 autenticado al mismo núcleo y TUI Bubble Tea v2/Bubbles v2 compartida.
- Límites de tamaño, tasas, conexiones, páginas y escritura; controles de terminal rechazados en entrada y filtrados en pantalla. Logs operativos sin cuerpos/credenciales.
- Pruebas de integración con PostgreSQL real, TLS, WebSocket, cliente ejecutable y SSH con PTY; CI con detector de carreras en Linux, compilación Windows/Linux/macOS y prueba Docker.

El cliente conserva tokens solo en memoria; solicita la clave en cada arranque. La TUI retiene 500 mensajes en pantalla y muestra las primeras 100 conversaciones. El registro inicial se realiza con el CLI; después SSH no requiere instalar TERMLINK. El tiempo real usa sondeos de 250 ms del registro PostgreSQL; no necesita Redis.

## Pendiente antes de abrir la alpha

Recuperación y segunda clave, eliminación de cuentas, purga de **30 días** decidida por el usuario, retención/restauración de respaldos y revisión de seguridad/privacidad. No hay purga automática de datos de desarrollo. No hay archivos, búsqueda, notificaciones, organizaciones ni facturación. El [roadmap](docs/roadmap.md) distingue estas tareas del sprint. TERMLINK tiene conflictos de nombre conocidos y requiere investigación de marca/dominio antes de publicación comercial.

## Estructura y documentación

`cmd/server`, `cmd/ssh-gateway`, `cmd/termlink`; `internal/auth`, `chat`, `storage`, `security`, `transport`; `pkg/tui`, `client`, `protocol`; `migrations`, `tests`, `deployments`, `docs`. Los límites reales de paquetes se explican en el [ADR 0001](docs/adr/0001-foundation.md), y la autenticación en el [ADR 0002](docs/adr/0002-authentication.md).

[Protocolo](docs/protocol.md) · [Seguridad](docs/security.md) · [Roadmap](docs/roadmap.md) · [Fuentes y versiones](docs/dependencies.md)

Código privado sin licencia pública concedida por este repositorio. La elección de licencia comercial o de distribución es una decisión pendiente del propietario. Las dependencias conservan sus licencias propias.
