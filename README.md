# Karatty (alpha en preparación)

Mensajería propia desde la terminal, con un cliente nativo y acceso SSH. Repositorio privado de trabajo de karatekah. Sin integración ni código de WhatsApp o wacon-mcp.

## Sprint 1

Objetivo: dos usuarios invitados conversan entre SSH y el cliente instalable, salen y recuperan el historial desde PostgreSQL; un tercero sin membresía no puede leer, enviar ni suscribirse.

Go, monolito modular, PostgreSQL y Bubble Tea v2/Bubbles v2/Wish v2. Ambos accesos ejecutan las mismas operaciones de dominio y comparten TUI. HTTPS/WSS y SSH protegen el transporte; **el servidor puede leer los mensajes; no hay E2EE**.

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
- Registro público opcional con correo verificado; recuperación de clave SSH con correo y código de recuperación mostrado una sola vez. El servidor exige configurar el proveedor de correo para activar este modo.
- Eliminación autenticada de cuenta, claves, sesiones, membresías y mensajes propios; los mensajes ajenos permanecen en un chat directo archivado y de solo lectura.
- Canales asignados por invitación y chats directos que requieren aceptación del destinatario; solicitudes de 7 días, rechazo y bloqueo. Autorización por sesión/clave/usuario/membresía en cada operación.
- Mensajes persistentes, secuencia transaccional por conversación, ACK de commit y deduplicación durante 30 días, con conflicto al cambiar el contenido de un reintento dentro de ese plazo.
- Historial paginado y eventos WSS con reconexión por secuencia; gateway Wish v2 autenticado al mismo núcleo y TUI Bubble Tea v2/Bubbles v2 compartida.
- Límites de tamaño, tasas, conexiones, páginas y escritura; controles de terminal rechazados en entrada y filtrados en pantalla. Logs operativos sin cuerpos/credenciales.
- Pruebas de integración con PostgreSQL real, TLS, WebSocket, cliente ejecutable y SSH con PTY; CI con detector de carreras en Linux, compilación Windows/Linux/macOS y prueba Docker.

El cliente conserva tokens solo en memoria; solicita la clave en cada arranque. La TUI retiene 500 mensajes en pantalla y muestra las primeras 100 conversaciones. El registro inicial se realiza con el CLI; después SSH no requiere instalar Karatty. El tiempo real usa sondeos de 250 ms del registro PostgreSQL; no necesita Redis.

## Pendiente antes de abrir la alpha

Faltan la contratación/configuración del proveedor de correo y dominio, la prueba operativa de la purga de **30 días**, la retención/restauración de respaldos remotos y la revisión de seguridad/privacidad. La eliminación de cuentas ya funciona en la base activa, pero falta verificar su vencimiento en respaldos remotos. El historial ya oculta mensajes vencidos; la purga automática se activa con `--retention-30d` o automáticamente con `--public-signup`. No hay archivos, búsqueda, notificaciones, organizaciones ni facturación. El [roadmap](docs/roadmap.md) distingue estas tareas del sprint. Karatty es el nombre elegido para la alpha; la marca y el dominio aún requieren comprobación definitiva antes de publicar.

## Estructura y documentación

`cmd/server`, `cmd/ssh-gateway`, `cmd/termlink`; `internal/auth`, `chat`, `storage`, `security`, `transport`; `pkg/tui`, `client`, `protocol`; `migrations`, `tests`, `deployments`, `docs`. Los límites reales de paquetes se explican en el [ADR 0001](docs/adr/0001-foundation.md), y la autenticación en el [ADR 0002](docs/adr/0002-authentication.md).

[Protocolo](docs/protocol.md) · [Seguridad](docs/security.md) · [Roadmap](docs/roadmap.md) · [Fuentes y versiones](docs/dependencies.md)
[Operación y respaldos para la alpha](docs/operations.md)
[Borrador de privacidad (no publicable aún)](docs/privacy-draft.md)
[Preparación de ejecutables públicos](docs/release.md)

Código privado sin licencia pública concedida por este repositorio. La elección de licencia comercial o de distribución es una decisión pendiente del propietario. Las dependencias conservan sus licencias propias.
